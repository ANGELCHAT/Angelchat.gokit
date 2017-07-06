package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"

	"github.com/colchis/investor/internal/platform/log"
)

// Wrapper wraps Endpoint with extra behavior such as logging, decoding, encoding, instrumentation, load balancing...
type Wrapper func(Endpoint) Endpoint

type TraceInfo struct {
	Timestamp        time.Time
	Endpoint         url.URL
	Headers          http.Header
	DNSLookup        time.Duration
	TCPConnection    time.Duration
	TLSHandshake     time.Duration
	ServerProcessing time.Duration
	ContentTransfer  time.Duration
	ContentSize      int
	Total            time.Duration
}

// URL takes complete string of url to connect by Endpoint
func URL(u, method string) Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			nu, uerr := url.Parse(u)
			if uerr != nil {
				res, _ := e.Do(r)
				return res, uerr
			}
			r.URL = nu
			r.Method = method
			return e.Do(r)
		})
	}
}

func JSONRequest() Wrapper {
	return Request("application/json", jsonEncode)
}

// Request is wrapper for input encoding into io.Writer. When structure is given as
// request then it's encoded by Encoder for given Content-Type (kind) request header.
func Request(kind string, en Encoder) Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {

			if r.Header.Get("Content-type") != "" {
				return e.Do(r)
			}
			r.Header.Set("Content-type", kind)

			b := &bytes.Buffer{}
			if ern := en(b, r.Context().Value("in")); ern != nil {
				res, rer := e.Do(r)
				if rer != nil {
					return res, rer
				}

				return res, ern
			}

			r.Body = ioutil.NopCloser(b)

			return e.Do(r)
		})
	}
}

func JSONResponse() Wrapper {
	return Response("application/json", JsonDecode)
}

func CSVResponse() Wrapper {
	return Response("text/csv", CsvDecodeScannerAsync)
}

// Response
func Response(kind string, d Decoder) Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Accept") == "" {
				r.Header.Set("Accept", kind)
			}

			res, err := e.Do(r)
			if err != nil {
				return res, err
			}

			if res.Header.Get("Content-type") != kind {
				return res, errors.New("wrong response content-type header, expected " + kind)
			}

			// copy out body stream to s
			b := &bytes.Buffer{}
			s := io.TeeReader(res.Body, b)
			res.Body = ioutil.NopCloser(b)

			// decode out into given value
			if err := d(s, r.Context().Value("out")); err != nil {
				return res, err
			}

			return res, err
		})
	}
}

// Authorization
func Authorization(token string) Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") == "" {
				r.Header.Set("Authorization", token)
			}

			res, err := e.Do(r)
			if err != nil {
				return res, err
			}

			if res.StatusCode == http.StatusUnauthorized {
				return res, fmt.Errorf("Wrong authorization token")
			}

			return res, err
		})
	}
}

// Logging
func Logging() Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			res, err := e.Do(r)

			log.Debug("HTTP.request.url", "[%s] %s", r.Method, r.URL)
			log.Debug("HTTP.request.headers", "%v", r.Header)
			in := r.Context().Value("in")
			if in != nil {
				log.Debug("HTTP.request.body", "%v", in)
			}

			if res != nil {
				//copy out body stream to s
				b := &bytes.Buffer{}
				s := io.TeeReader(res.Body, b)
				res.Body = ioutil.NopCloser(b)

				o, _ := ioutil.ReadAll(s)
				log.Debug("HTTP.response.status", "%v", res.Status)
				log.Debug("HTTP.response.headers", "%v", res.Header)
				log.Debug("HTTP.response.size", "%.2fKB", float64(len(o))/1024)
				//log.Debug("HTTP.response.body", "%s", string(o))
			}

			return res, err
		})
	}
}

// Trace gives detailed information about HTTP call, it will look and measure
// all the HTTP parts such as tcp connection, dns lookup, tlc handshaking and
// body transfer.
func Trace(fn func(TraceInfo)) Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			var t0, t1, t2, t3, t4 time.Time
			t0 = time.Now()

			trace := &httptrace.ClientTrace{
				DNSStart: func(_ httptrace.DNSStartInfo) {
					t0 = time.Now()
				},
				DNSDone: func(_ httptrace.DNSDoneInfo) {
					t1 = time.Now()
				},
				ConnectStart: func(_, _ string) {
					if t1.IsZero() {
						// connecting to IP
						t1 = time.Now()
					}
				},
				ConnectDone: func(net, addr string, err error) {
					if err != nil {
						log.Info("unable connect to host %v: %v", addr, err)
					}
					t2 = time.Now()
				},
				GotConn: func(i httptrace.GotConnInfo) {
					t3 = time.Now()
				},
				GotFirstResponseByte: func() { t4 = time.Now() },
			}

			r = r.WithContext(httptrace.WithClientTrace(r.Context(), trace))

			res, err := e.Do(r)

			if err != nil {
				log.Error("HTTP.trace.error", "%s", err.Error())
				return res, err
			}

			// in order to measure all LC timings, we need to drain whole
			// body and copy new into res.Body again
			b := &bytes.Buffer{}
			s := io.TeeReader(res.Body, b)
			res.Body = ioutil.NopCloser(b)
			rb, _ := ioutil.ReadAll(s)

			t5 := time.Now() // after read body

			if t0.IsZero() {
				// we skipped DNS
				t0 = t1
			}

			if t0.IsZero() && t2.IsZero() {
				t0 = t3
				t1 = t3
				t2 = t3
			}

			//connection timestamp
			var t time.Time
			if !t0.IsZero() {
				t = t0
			} else {
				t = t3
			}

			ti := TraceInfo{
				Timestamp:        t,
				DNSLookup:        t1.Sub(t0),
				TCPConnection:    t2.Sub(t1),
				TLSHandshake:     t3.Sub(t2),
				ServerProcessing: t4.Sub(t3),
				ContentTransfer:  t5.Sub(t4),
				Total:            t5.Sub(t0),
				Endpoint:         *r.URL,
				Headers:          r.Header,
				ContentSize:      len(rb),
			}

			token := r.Header.Get("Authorization")
			if len(token) >= 8 {
				token = token[:8]
			}

			log.Info("HTTP.trace.info", "[%s***] %v "+
				"started at %s, "+
				"done %s and duration %s",
				token,
				ti.Endpoint.String(),
				ti.Timestamp.Format("15:04:05.000"),
				ti.Timestamp.Add(ti.Total).Format("15:04:05.000"),
				ti.Total,
			)

			log.Debug("HTTP.trace.details",
				"\n\ttime: %s\n"+
					"\tdns: %s\n"+
					"\ttcp: %s\n"+
					"\ttls: %s\n"+
					"\tserver_processing: %s\n"+
					"\tcontent_read: %s\n"+
					"\ttotal: %s -> %s",
				ti.Timestamp.Format("2006-01-02 15:04:05.000"),
				ti.DNSLookup,
				ti.TCPConnection,
				ti.TLSHandshake,
				ti.ServerProcessing,
				ti.ContentTransfer,
				ti.Total,
				ti.Timestamp.Add(ti.Total).Format("15:04:05.000"),
			)

			if fn != nil {
				fn(ti)
			}

			return res, err
		})
	}
}

func ResponseError() Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			res, err := e.Do(r)

			if res.StatusCode < 200 || res.StatusCode >= 400 {
				return res, fmt.Errorf("[%s] %s [%s]", r.Method, r.URL, res.Status)
			}

			return res, err
		})
	}
}

func After(t time.Time) Wrapper {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			n := time.Now()
			if n.Before(t) {
				res, err := e.Do(r)
				if err != nil {
					return res, err
				}

				return res, errors.New("it's to early to call resurce, allowed after " + t.String())
			}

			return e.Do(r)
		})
	}
}

// Delay will hold Endpoint for a while, until given duration. It takes care
// for an endpoint to not be called often than given duration of time.
func Delay(limit time.Duration, offset time.Duration) Wrapper {
	var last time.Time
	var m sync.Mutex
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			called := time.Now()
			m.Lock()
			// ignore first call
			if !last.IsZero() {
				wait(limit, offset)
			}
			now := time.Now()

			if last.IsZero() {
				last = now
			}
			last = called

			//debug info to be sure what's is the state of each HTTP call
			log.Debug(""+
				"HTTP.delay: last %s, executed %s, done %s, idle %s",
				last.Format("15:04:05.000"),
				called.Format("15:04:05.000"),
				now.Format("15:04:05.000"),
				now.Sub(called))
			m.Unlock()

			return e.Do(r)
		})
	}
}

// wait waits for time occurence ie d=time.Second and o=0ms then will wait on
// every 1.000s, 2.000s, 3.000s. If you decide to create it with
// 2 * time.Second and 299ms then will wait on 1.299s, 3.299s, 5.299s...
// todo make it more reliable, especially on first few ticks and duration less than 1sec
func wait(limit time.Duration, offset time.Duration) {
	n := time.Now()
	x := n.Truncate(time.Second).Add(offset)
	if x.Before(n) {
		x = x.Add(limit)
	}
	<-time.NewTimer(x.Sub(n)).C
}
