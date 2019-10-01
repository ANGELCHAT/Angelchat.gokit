package server

type Endpoint interface {
	Do(*Request)
}

type EndpointFunc func(*Request)

func (f EndpointFunc) Do(r *Request) { f(r) }
