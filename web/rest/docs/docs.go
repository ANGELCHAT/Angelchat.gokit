package docs

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Doc struct {
	*Options
	readme *readme
	off    bool
}

func Documentation(oo ...Option) *Doc {
	d := &Doc{Options: newOptions(oo...)}
	d.readme = &readme{d}

	return d
}

func (d *Doc) Read(routers ...*mux.Router) error {
	var first = routers[0]

	register := func(path string, e http.HandlerFunc, method string) {
		path = strings.TrimSpace(fmt.Sprintf("%s%s", d.HttpPrefix, path))
		if err := first.
			HandleFunc(path, e).
			Methods(method).
			GetError(); err != nil {
			info("%s", fmt.Errorf("[%s] %s: %s", method, path, err))
			//return err
		}

		//return nil
	}

	//register("/docs", d.all, "GET")
	register("/docs/activate/{action}", d.activate, "GET")
	register("/docs/readme.apib", d.readme.apib(d.Storage), "GET")
	register("/docs/readme.md", d.readme.markdown(d.Storage), "GET")
	register("/docs/readme.html", d.readme.html(d.Storage), "GET")
	register("/docs/{endpoint}", d.describeEndpoint, "PATCH")
	register("/docs/{endpoint}/activate", d.activateEndpoint, "GET")
	register("/docs/{endpoint}/{object}/{field}", d.describeField, "PATCH")

	for _, router := range routers {
		if err := d.read(router); err != nil {
			return err
		}
		router.Use(d.auth("dps-admin", "mL9j1dAM"))
		router.Use(d.observe(router))
	}

	d.off = true

	return nil
}

func (d *Doc) read(r *mux.Router) error {
	var (
		merged      []Endpoint
		fromCode    = make(map[string]Endpoint)
		fromStorage = d.Storage.All()
	)

	// read all registered http handlers with it's definitions and prepare
	// Endpoint list.
	err := r.Walk(func(mr *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		var (
			p, _ = mr.GetPathTemplate()
			m, _ = mr.GetMethods()
			id   = id(mr)
		)

		if len(m) == 0 {
			return nil
		}

		fromCode[id] = Endpoint{ID: id, Method: m[0], Path: p}

		return nil
	})

	if err != nil {
		return err
	}

	// It compares endpoints from storage with endpoints from code definition.
	// When endpoints from different sources are met, then we merge them together.
	// Thanks to that we will get new Endpoint definition with fresh structure.
	for _, se := range fromStorage {
		_, ok := fromCode[se.ID]
		if !ok {
			if err := d.Storage.Delete(se.ID); err != nil {
				return err
			}
		}
		delete(fromCode, se.ID)
	}

	// add new definitions from code
	for _, re := range fromCode {
		merged = append(merged, re)
	}

	debug("processed %d from code, %d from storage, %d merged in result",
		len(fromCode), len(fromStorage), len(merged))

	return d.Storage.Store(merged...)
}

func (d *Doc) auth(user, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/v2/docs/") {
				u, p, _ := r.BasicAuth()
				if u != user || p != password {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
					http.Error(w, "Not authorized", 401)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (d *Doc) observe(mr *mux.Router) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if d.off {
				next.ServeHTTP(w, r)
				return
			}

			if r == nil {
				return
			}

			var (
				buff     = &bytes.Buffer{}
				begin    = time.Now()
				request  = new(http.Request)
				endpoint mux.RouteMatch
				response = &response{w, &bytes.Buffer{}, 0}
			)

			// take endpoint definition from mux router.
			if !mr.Match(r, &endpoint) {
				return
			}

			id := id(endpoint.Route)

			*request = *r
			// Drain endpoint request body into buffer, remaining request body
			// data untouched.
			if r.Body != nil {
				r.Body = ioutil.NopCloser(io.TeeReader(r.Body, buff))
				request.Body = ioutil.NopCloser(buff)
			}

			// call another request decorator, down to last one.
			next.ServeHTTP(response, r)

			// process endpoint...
			go func(id string) {
				e := d.Storage.Load(id)
				if !e.Eager {
					return
				}
				tn := time.Now()
				e.Visits++
				e.Definition = fields.extend(e.Definition, newDefinition(endpoint, response, request))
				te := time.Since(tn)

				//o := e.Definition.Response
				//fmt.Println(e.String())

				//fmt.Println(e.GoString())
				//fmt.Println(o.Root())
				//fmt.Println(o.JSON(0))

				if err := d.Storage.Store(e); err != nil {
					info("%s", err)
					return
				}

				debug("[%s] %s in %s(calculation) %s(total)", r.Method, r.URL.Path, te, time.Since(begin))

			}(id)

		})
	}
}

func (d *Doc) activate(w http.ResponseWriter, r *http.Request) {
	d.off = mux.Vars(r)["action"] == "off"

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (d *Doc) activateEndpoint(w http.ResponseWriter, r *http.Request) {
	e := d.Storage.Load(mux.Vars(r)["endpoint"])
	e.Eager = !e.Eager

	if err := d.Storage.Store(e); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (d *Doc) describeEndpoint(w http.ResponseWriter, r *http.Request) {
	type request struct{ Usage, Access, Description string }
	var in request

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		info("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	endpoint := d.Storage.Load(mux.Vars(r)["endpoint"])
	endpoint.Usage = in.Usage
	endpoint.Description = in.Description
	endpoint.Access = strings.Split(in.Access, ",")

	if err := d.Storage.Store(endpoint); err != nil {
		info("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *Doc) describeField(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Description, Example, Type string
		Required, Ignore           bool
	}

	var (
		req      = request{}
		err      = json.NewDecoder(r.Body).Decode(&req)
		endpoint = d.Storage.Load(mux.Vars(r)["endpoint"])
		object   = mux.Vars(r)["object"]
		field    = mux.Vars(r)["field"]
	)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if endpoint.ID == "" {
		http.NotFound(w, r)
		return
	}

	//fields = endpoint.Definition.part(object)
	ff := endpoint.Definition.part(object)
	f, pos := fields.find(field, ff)
	if pos == -1 {
		http.NotFound(w, r)
		return
	}

	f.Type = req.Type
	f.Description = req.Description
	f.Example = req.Example
	f.Ignored = req.Ignore
	f.Required = req.Required
	ff[pos] = f

	if err := d.Storage.Store(endpoint); err != nil {
		info("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (d *Doc) all(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", "application/json")

	json.NewEncoder(res).Encode(d.Storage.All())
}

type response struct {
	r    http.ResponseWriter
	body *bytes.Buffer
	code int
}

func (w *response) Write(b []byte) (int, error) {
	_, err := w.body.Write(b)
	if err != nil {
		info("response body error: %s", err)
	}
	return w.r.Write(b)
}

func (w *response) WriteHeader(statusCode int) {
	w.code = statusCode
	w.r.WriteHeader(statusCode)
}

func (w *response) Header() http.Header {
	return w.r.Header()
}

func id(r *mux.Route) string {
	p, _ := r.GetPathTemplate()
	m, _ := r.GetMethods()
	return hash(m, p)
}

func hash(words ...interface{}) string {
	var text string
	for _, w := range words {
		text += fmt.Sprintf("%+v", w)
	}

	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
