package gokit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Registry map[string]reflect.Type

func (r Registry) Register(objects ...interface{}) {
	for i := range objects {
		name, typ, err := r.Name(objects[i])
		if err != nil {
			panic(err)
		}

		if _, ok := r[name]; ok {
			return
		}

		r[name] = typ
	}
}

func (r Registry) JSON(name string, body []byte) (interface{}, error) {
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

func (r Registry) Has(object interface{}) (string, error) {
	name, _, err := r.Name(object)
	if err != nil {
		return name, err
	}

	if _, ok := r[name]; !ok {
		return name, fmt.Errorf("object %s not found in reqistry", name)
	}

	return name, nil
}

func (r Registry) Name(object interface{}) (string, reflect.Type, error) {
	typ := reflect.TypeOf(object)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Name() == typ.Kind().String() {
		return "", typ, fmt.Errorf("registry accepts only named types")
	}

	return strings.Replace(typ.Name(), "*", "", -1), typ, nil

}

func (r Registry) Names() []string {
	var o []string
	for n := range r {
		o = append(o, n)
	}
	return o
}

func examiner(t reflect.Type, depth int) {
	fmt.Println(strings.Repeat("\t", depth), "Type is", t.Name(), "and kind is", t.Kind())
	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		fmt.Println(strings.Repeat("\t", depth+1), "Contained type:")
		examiner(t.Elem(), depth+1)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Println(strings.Repeat("\t", depth+1), "Field", i+1, "name is", f.Name, "type is", f.Type.Name(), "and kind is", f.Type.Kind())
			if f.Tag != "" {
				fmt.Println(strings.Repeat("\t", depth+2), "Tag is", f.Tag)
				fmt.Println(strings.Repeat("\t", depth+2), "tag1 is", f.Tag.Get("tag1"), "tag2 is", f.Tag.Get("tag2"))
			}
		}
	}
}
