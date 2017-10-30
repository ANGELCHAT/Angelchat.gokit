package platform

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sokool/gokit/log"
)

type Serializer struct {
	object map[string]reflect.Type
}

func (s *Serializer) Marshal(n string, v interface{}) ([]byte, error) {
	if _, ok := s.object[n]; !ok {
		return []byte{}, fmt.Errorf("object '%s' is not registerd", n)
	}

	//data, err := gocsv.MarshalBytes(v)
	//data, err := binary.Marshal(v)
	data, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (s *Serializer) Unmarshal(n string, data []byte) (interface{}, error) {
	t, ok := s.object[n]
	if !ok {
		return nil, fmt.Errorf("object %s is not registerd", n)
	}

	v := reflect.New(t).Interface()

	//if err := gocsv.UnmarshalBytes(data, v); err != nil {
	//if err := binary.Unmarshal(data, v); err != nil {
	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Serializer) Register(vs ...interface{}) error {
	for _, v := range vs {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		_, has := s.object[t.Name()]
		if has {
			return fmt.Errorf("object %s name already registerd", t.Name())
		}

		s.object[t.Name()] = t
		log.Debug("serializer", "%s registered", t.Name())
	}

	return nil
}

func NewSerializer(vs ...interface{}) *Serializer {

	s := &Serializer{
		object: make(map[string]reflect.Type),
	}

	s.Register(vs...)
	return s
}
