package gokit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Registry map[string]reflect.Type

func (r Registry) Register(objects ...interface{}) ([]string, error) {
	var names []string
	for i := range objects {
		name, err := r.Add(objects[i], "")
		if err != nil {
			return names, err
		}

		names = append(names, name)
	}

	return names, nil
}

func (r Registry) Add(object interface{}, name string) (string, error) {
	n, typ, err := r.compose(object)
	if err != nil {
		return name, err
	}

	if name == "" {
		name = n
	}

	r[name] = typ

	return name, nil
}

func (r Registry) Exists(name string) bool { _, ok := r[name]; return ok }

func (r Registry) Has(object interface{}) (string, error) {
	name, _, err := r.compose(object)
	if err != nil {
		return name, err
	}

	if _, ok := r[name]; !ok {
		return name, fmt.Errorf("object %s not found in reqistry", name)
	}

	return name, nil
}

func (r Registry) Names() []string {
	var o []string
	for n := range r {
		o = append(o, n)
	}
	return o
}

func (r Registry) Info(name string) string {
	////t, ok := r[name]
	////if !ok {
	////	return "undefined"
	////}
	////
	////
	////fmt.Println(name)
	////var nfo string
	////for i := 0; i < t.NumField(); i++ {
	////	nfo += t.Field(i).Name + ":" + t.Field(i).Type.Name() + ", "
	////}
	////
	////return fmt.Sprintf("{%s}", strings.TrimSuffix(nfo, ", "))
	////return ""
	//
	//z, e1 := r.Decode(name, []byte(""))
	//if e1 != nil {
	//	fmt.Println(name, e1)
	//}
	//o, e2 := json.Marshal(z)
	//if e2 != nil {
	//	fmt.Println(name, e2)
	//}
	return "TBD..."
}

func (r Registry) Decode(name string, body []byte) (interface{}, error) {
	name = strings.Title(name)
	o, ok := r[name]
	if !ok {
		return nil, fmt.Errorf("%s definition not found", name)
	}

	value := reflect.New(o)
	if len(body) > 0 {
		if err := json.Unmarshal(body, value.Interface()); err != nil {
			fmt.Println(string(body))
			return nil, err
		}
	}

	return value.Elem().Interface(), nil
}

func (r Registry) Encode(object interface{}) (string, []byte, error) {
	var (
		name string
		body []byte
		err  error
	)

	if name, err = r.Has(object); err != nil {
		return name, nil, err
	}

	body, err = json.Marshal(object)

	return name, body, err
}

func (r Registry) compose(object interface{}) (string, reflect.Type, error) {
	var typ reflect.Type
	var ok bool
	if typ, ok = object.(reflect.Type); !ok {
		typ = reflect.TypeOf(object)
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Name() == typ.Kind().String() {
		return "", typ, fmt.Errorf("only named types allowed")
	}

	name := strings.Replace(typ.Name(), "*", "", -1)

	return strings.Title(name), typ, nil
}
