package docs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

type Field struct {
	ID, Path, Type, Description, Example string
	Array, Object, Ignored, Required     bool
	LastUsed                             string
}

func (f Field) String() string {
	b, _ := json.MarshalIndent(f, "", "  ")
	return string(b)
}

func (f Field) NameR(n string) string { return strings.Replace(f.Path, n+"|", "", -1) }

func (f Field) NormalisedPath() string {
	i := strings.Index(f.Path, "|")
	if i > 0 {
		return strings.Replace(f.Path[i+1:], "|", ".", -1)
	}
	return f.Path
}

func (f Field) Name() string {
	if i := strings.LastIndex(f.Path, "|"); i > 0 {
		return f.Path[i+1:]
	}

	return f.Path
}

func (f Field) IndentedName() string {
	n := ""
	for i := 0; i < strings.Count(f.Path, "|")*2; i++ {
		n += " "
	}
	return n + f.Name()
}

//func (f Field) XPath() string { return strings.Replace(f.Path, "|", ".", -1) }
func (f Field) Belongs() string {
	if i := strings.LastIndex(f.Path, "|"); i > 0 {
		return f.Path[:i]
	}

	return f.Path
}

func (f Field) IsRoot() bool { return !strings.Contains(f.Path, "|") }

func (f Field) IsPart(of string) bool {
	if p := strings.Index(f.Path, of); p < 0 {
		return false
	}

	if !strings.Contains(f.Path[len(of):], "|") {
		return true
	}

	return false
}

type Object []Field

func (o Object) Root() Field {
	for _, f := range o {
		if f.IsRoot() {
			return f
		}
	}

	return Field{}
}

func (o Object) JSON(indent int) string { return fields.json(o.Root().Path, o, indent) }

func (o Object) Filter(k string) Object {
	var out Object
	for _, f := range o {
		if strings.Contains(f.Path, k) {
			out = append(out, f)
		}
	}

	return out
}

func (o Object) Group() map[string]Object {
	var (
		out = map[string]Object{}
	)

	for _, f := range o {
		if f.Object {
			var niu Object
			niu = append(niu, f)
			fmt.Println()
			fmt.Println(f.Path)
			fmt.Println("============")
			for _, f2 := range o {
				if f2.IsPart(f.Path + "|") {
					fmt.Println("  ", f2.Path[len(f.Path):], f2.Object)
					niu = append(niu, f2)
				}
			}

			out[f.Path] = niu
		}
	}

	for k, f := range fields.objects(o) {
		fmt.Println(k)
		fmt.Println(f.Vars)
		fmt.Println("=====================")
	}
	os.Exit(1)
	return out
}

func (o Object) Len() int { return len(o) }

func (o Object) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o Object) ByName() Object {
	sort.Sort(byName{o})
	return o
}

type Definition struct {
	Headers, Parameters, Query, Request, Response Object
}

func newDefinition(e mux.RouteMatch, w *response, r *http.Request) Definition {
	var res []Field

	if w.Header().Get("content-type") == "application/json" {
		res = fields.body(w.body, "response")
	}

	return Definition{
		Headers:    fields.headers(r.Header, "header"),
		Parameters: fields.parameters(e.Route, "parameter"),
		Query:      fields.query(*r.URL, "query"),
		Request:    fields.body(r.Body, "request"),
		Response:   res,
	}
}

func (d Definition) part(what string) []Field {
	var fields []Field
	switch what {
	case "response":
		fields = d.Response
	case "request":
		fields = d.Request
	case "header":
		fields = d.Headers
	case "parameter":
		fields = d.Parameters
	case "query":
		fields = d.Query
	}

	return fields
}

func (d Definition) Objects(f string) map[string]*object {
	return fields.objects(d.part(f))
}

type byName struct{ Object }

func (s byName) Less(i, j int) bool { return s.Object[i].Path < s.Object[j].Path }
