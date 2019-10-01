package docs

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Endpoint struct {
	ID, Method, Path, Usage, Description string
	Visits                               uint64
	Eager                                bool
	Access                               []string
	Definition                           Definition
}

func (e Endpoint) Id() string { return e.ID }

func (e Endpoint) Roles() string { return strings.Join(e.Access, ",") }

func (e Endpoint) GoString() string {
	b, x := json.MarshalIndent(e, "", "\t")
	if x != nil {
		info("%s", x)
	}
	return fmt.Sprintf("%T%s", e, b)
}

func (e Endpoint) String() string { return fmt.Sprintf("%s: [%s] {%s}", e.Path, e.Method, e.Usage) }

type endpoints []Endpoint

func (s endpoints) Len() int { return len(s) }

func (s endpoints) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type byPath struct{ endpoints }

func (s byPath) Less(i, j int) bool {
	a := s.endpoints[i].Path + s.endpoints[i].Method
	b := s.endpoints[j].Path + s.endpoints[j].Method
	return a < b
}
