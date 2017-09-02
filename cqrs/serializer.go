package cqrs

import (
	"encoding/json"
	"fmt"
	//"github.com/alecthomas/binary"
)

type serializer struct {
	object map[string]structure
}

func (s *serializer) Marshal(n string, v interface{}) ([]byte, error) {
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

func (s *serializer) Unmarshal(n string, data []byte) (interface{}, error) {
	t, ok := s.object[n]
	if !ok {
		return nil, fmt.Errorf("object %s is not registerd", n)
	}

	v := t.Instance()

	//if err := gocsv.UnmarshalBytes(data, v); err != nil {
	//if err := binary.Unmarshal(data, v); err != nil {
	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}

	return v, nil
}

func newSerializer(es ...interface{}) *serializer {
	os := map[string]structure{}
	for _, v := range es {
		s := newStructure(v)
		os[s.Name] = s
	}

	return &serializer{
		object: os,
	}
}
