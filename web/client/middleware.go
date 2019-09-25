package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

type Logger = func(message string, args ...interface{})

// Middleware wraps Endpoint with extra behavior such as logging, decoding, encoding,
// authentication, error handling, tracing...
type Middleware func(Endpoint) Endpoint

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
func URL(u, method string) Middleware {
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

func JSONRequest() Middleware {
	return Request("application/json", jsonEncode)
}

func FormRequest() Middleware {
	return Request("application/x-www-form-urlencoded", func(w io.Writer, v interface{}) error {
		d, ok := v.(url.Values)
		if !ok {
			return fmt.Errorf("application/x-www-form-urlencoded accept url.Value as a request object")
		}

		w.Write([]byte(d.Encode()))
		return nil
	})
}

// Request is wrapper for input encoding into io.Writer. When structure is given as
// request then it's encoded by Encoder for given Content-Type (kind) request header.
func Request(kind string, en Encoder) Middleware {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			input := r.Context().Value("in")
			if input == nil {
				return e.Do(r)
			}

			if r.Header.Get("Content-type") != "" {
				return e.Do(r)
			}

			r.Header.Set("Content-type", kind)
			b := &bytes.Buffer{}
			if ern := en(b, input); ern != nil {
				res, rer := e.Do(r)
				if rer != nil {
					return res, rer
				}

				return res, ern
			}

			r.ContentLength = int64(len(b.String()))
			r.Body = ioutil.NopCloser(b)

			return e.Do(r)
		})
	}
}

func JSONResponse() Middleware {
	return Response("application/json", JsonDecode)
}

func CSVResponse() Middleware {
	return Response("text/csv", CsvDecodeScannerAsync)
}

// Response
func Response(kind string, d Decoder) Middleware {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			output := r.Context().Value("out")
			if output == nil {
				return e.Do(r)
			}

			if r.Header.Get("Accept") == "" {
				r.Header.Set("Accept", kind)
			}

			res, err := e.Do(r)
			if err != nil {
				return res, err
			}

			//if res.Title.Get("Content-type") != kind {
			//	return res, errors.New("wrong response content-type header, expected " + kind + " has: " + res.Title.Get("Content-type"))
			//}

			// copy out body stream to s
			b := &bytes.Buffer{}
			s := io.TeeReader(res.Body, b)
			res.Body = ioutil.NopCloser(b)

			// decode out into given value
			if err := d(s, output); err != nil {
				return res, err
			}

			return res, err
		})
	}
}

// Authorization
func Authorization(token string) Middleware {
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

func Header(k, v string) Middleware {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {

			r.Header.Set(k, v)

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
func Logging(log Logger) Middleware {

	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			res, err := e.Do(r)

			log("HTTP.request.url", "[%s] %s", r.Method, r.URL)
			log("HTTP.request.headers", "%v", r.Header)
			in := r.Context().Value("in")
			if in != nil {
				log("HTTP.request.body", "%v", in)
			}

			if res != nil {
				//copy out body stream to s
				b := &bytes.Buffer{}
				s := io.TeeReader(res.Body, b)
				res.Body = ioutil.NopCloser(b)

				o, _ := ioutil.ReadAll(s)
				log("HTTP.response.status", "%v", res.Status)
				log("HTTP.response.headers", "%v", res.Header)
				log("HTTP.response.size", "%.2fKB", float64(len(o))/1024)
				log("HTTP.response.body", "%s\n", string(o))
			}

			return res, err
		})
	}
}

// Trace gives detailed information about HTTP call, it will look and measure
// all the HTTP parts such as tcp connection, dns lookup, tlc handshaking and
// body transfer.
func Trace(fn func(TraceInfo), log Logger) Middleware {
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
						log("HTTP.trace.details", "unable connect to host %v: %v", addr, err)
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
				log("HTTP.trace.details.error", err.Error())
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

			log("HTTP.trace.details",
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

func ResponseError(fn func(int, []byte) error) Middleware {
	return func(e Endpoint) Endpoint {
		return EndpointFunc(func(r *http.Request) (*http.Response, error) {
			res, err := e.Do(r)
			if err != nil {
				return res, err
			}

			if res.StatusCode < 200 || res.StatusCode >= 400 {
				if fn != nil {
					b := &bytes.Buffer{}
					data, _ := ioutil.ReadAll(io.TeeReader(res.Body, b))
					res.Body = ioutil.NopCloser(b)

					return res, fn(res.StatusCode, data)
				}

				return res, fmt.Errorf("[%s] %s [%s]", r.Method, r.URL, res.Status)
			}

			return res, err
		})
	}
}
