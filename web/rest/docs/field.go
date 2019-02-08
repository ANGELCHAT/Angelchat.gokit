package docs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type (
	_fields struct{}
)

var fields = _fields{}

func (c _fields) build(path string, in interface{}) map[string]Field {
	rec := make(map[string]Field)
	switch v := in.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return nil
		}

		f := c.new(path, "struct")
		rec[f.ID] = f

		for n, r := range v {
			if path != "" {
				n = fmt.Sprintf("%s|%s", path, n)
			}

			for id, f := range c.build(n, r) {
				rec[id] = f
			}
		}

	case []interface{}:
		f := c.new(path, "slice")
		rec[f.ID] = f

		for _, r := range v {
			for id, f := range c.build(path, r) {
				rec[id] = f
			}
		}

		if r, ok := rec[f.ID]; ok {
			r.Array = true
			rec[f.ID] = r
		}

	case []map[string]interface{}:
		f := c.new(path, "slice")
		rec[f.ID] = f
		for _, r := range v {
			for id, f := range c.build(path, r) {
				rec[id] = f
			}
		}
		if r, ok := rec[f.ID]; ok {
			r.Array = true
			rec[f.ID] = r
		}

	default:
		if v == nil {
			return rec
		}

		f := c.new(path, fmt.Sprintf("%T", v), fmt.Sprintf("%v", v))
		rec[f.ID] = f
	}

	return rec
}

func (c _fields) new(path, kind string, examples ...string) Field {
	var s string
	if len(examples) > 0 {
		s = examples[0]
	}

	var typ = kind
	if kind == "struct" {
		i := strings.LastIndex(path, "|")
		if i > 0 {
			typ = strings.Title(path[i+1:])
		} else {
			typ = strings.Title(path)
		}
	}

	if strings.Contains(typ, "float") ||
		strings.Contains(typ, "int") ||
		strings.Contains(typ, "uint") {
		typ = "number"
	}

	if strings.Contains(typ, "bool") {
		typ = "boolean"
	}

	if strings.Contains(typ, "struct") {
		typ = "object"
	}

	return Field{
		ID:       hash(path),
		Path:     path,
		Type:     typ,
		Example:  s,
		Array:    kind == "slice",
		Object:   kind == "struct",
		LastUsed: time.Now().Format("2006/01/02 15:04:05"),
	}
}

func (c _fields) body(r io.Reader, name string) []Field {
	var (
		v   interface{}
		out []Field
	)

	if r == nil {
		return nil
	}

	if err := json.NewDecoder(r).Decode(&v); err != nil && err != io.EOF {
		info("fields decoding %s\n%s", err)
		return nil
	}

	for _, f := range c.build(name, v) {
		out = append(out, f)
	}
	return out
}

func (c _fields) query(u url.URL, name string) []Field {
	var ff []Field
	if len(u.Query()) == 0 {
		return ff
	}

	ff = append(ff, c.new(name, "struct"))
	for n := range u.Query() {
		ff = append(ff, c.new(fmt.Sprintf("%s|%s", name, n), "string"))
	}
	return ff
}

func (c _fields) headers(hh http.Header, name string) []Field {
	var ff []Field
	ff = append(ff, c.new(name, "struct"))
	for n, v := range hh {
		if len(v) == 0 {
			continue
		}
		f := c.new(fmt.Sprintf("%s|%s", name, n), "string")
		f.Example = v[0]
		f.Ignored = true
		ff = append(ff, f)
	}
	return ff
}

func (c _fields) parameters(r *mux.Route, name string) []Field {
	var (
		vv   []Field
		p, _ = r.GetPathTemplate()
		re   = regexp.MustCompile(`\{([^\{\}]*)\}`)
		pp   = re.FindAllString(p, -1)
	)

	if len(pp) == 0 {
		return nil
	}

	vv = append(vv, c.new(name, "struct"))
	for _, n := range pp {
		n = strings.Trim(n, "{")
		n = strings.Trim(n, "}")
		vv = append(vv, c.new(name, "struct"), c.new(fmt.Sprintf("%s|%s", name, n), "string"))
	}

	return vv
}

func (c _fields) merge(to, from []Field) []Field {
	toM := c.mape(to)
	fromM := c.mape(from)

	if toM == nil {
		toM = make(map[string]Field)
	}

	for id, f := range fromM {
		if _, ok := toM[id]; ok {
			t := toM[id]
			t.LastUsed = f.LastUsed
			toM[id] = t
			continue
		}
		toM[id] = f
	}

	var out []Field
	for _, f := range toM {
		out = append(out, f)
	}

	return out
}

func (c _fields) extend(a, b Definition) Definition {
	return Definition{
		Headers:    c.merge(a.Headers, b.Headers),
		Parameters: c.merge(a.Parameters, b.Parameters),
		Query:      c.merge(a.Query, b.Query),
		Request:    c.merge(a.Request, b.Request),
		Response:   c.merge(a.Response, b.Response),
	}
}

type (
	object struct {
		Field Field
		Vars  []Field
	}
)

func (c _fields) objects(ff []Field) map[string]*object {
	objects := map[string]*object{}

	for _, f := range ff {
		parts := strings.Split(f.Path, "|")
		size := len(parts)
		if size == 0 {
			info("uuu nie wiem")
			return objects
		}

		for i := range parts {
			if i < size-1 {
				n := strings.Join(parts[:size-1], "|")
				if _, ok := objects[n]; ok {
					continue
				}

				if _, ok := objects[n]; !ok {
					objects[n] = &object{}
				}
				objects[n] = &object{}
			}
		}

		if size >= 2 {
			n := strings.Join(parts[:size-1], "|")
			o := objects[n]
			o.Vars = append(o.Vars, f)
		}

		if f.Object || f.Array {
			if _, ok := objects[f.Path]; !ok {
				objects[f.Path] = &object{}
			}
			objects[f.Path].Field = f

		}
	}

	return objects
}

func (c _fields) json(x string, ff []Field, p int) string {
	type obj map[string]interface{}
	type arr []interface{}

	jsn := obj{}

	oo := c.objects(ff)
	cache := map[string]bool{}
	add := func(o object) {
		if _, ok := cache[o.Field.Path]; ok {
			return
		}

		f := o.Field
		tmp := jsn
		parts := strings.Split(f.Path, "|")
		for _, n := range parts {
			//fmt.Println(n)
			if _, ok := tmp[n]; !ok {
				if f.Array && f.Object {
					tmp[n] = arr{obj{}}
				} else if f.Object {
					tmp[n] = obj{}
				}
			} else {
				if _, ok := tmp[n].(string); ok {
					tmp[n] = arr{f.Example}
					break
				}
			}
			switch tmp[n].(type) {
			case obj:
				tmp = tmp[n].(obj)

			case arr:
				x := tmp[n].(arr)
				tmp = x[0].(obj)

			default:
				//tmp = obj{}
				info("%s", fmt.Errorf("uups %s:%T not supported", n, tmp[n]))
			}

		}

		for _, v := range o.Vars {
			if v.Object || (v.Object && v.Array) || v.Ignored {
				continue
			}
			p := strings.Split(v.Path, "|")
			tmp[p[len(p)-1]] = v.Example
		}

		cache[o.Field.Path] = true

	}
	for _, o := range oo {
		parts := strings.Split(o.Field.Path, "|")
		for i := range parts {
			n := strings.Join(parts[:i+1], "|")
			if n == "" {
				n = parts[i]
			}

			m, ok := oo[n]
			if !ok {
				info("%s", fmt.Errorf("uups, field %s not found", n))
				break
			}

			add(*m)
		}
	}

	//b, _ := json.Marshal(jsn[x])
	b, _ := json.MarshalIndent(jsn[x], fmt.Sprintf(fmt.Sprintf("%%-%ds", p), " "), "  ")
	return string(b)
}

func (c _fields) find(id string, ff []Field) (Field, int) {
	var f Field
	for i := range ff {
		if ff[i].ID == id {
			return ff[i], i
		}
	}
	return f, -1
}

func (c _fields) mape(ff []Field) map[string]Field {
	out := map[string]Field{}
	for i := range ff {
		out[ff[i].ID] = ff[i]
	}
	return out
}
