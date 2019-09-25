package docs

type Endpoints interface {
	Store(...Endpoint) error
	Load(string) Endpoint
	Delete(string) error
	All() []Endpoint

	//StorePayload(Definition) error
	//Payload(string) Definition
}

type repository struct{ l map[string]*Endpoint }

func newRepository() Endpoints {
	return &repository{make(map[string]*Endpoint)}
}

func (r *repository) Store(ee ...Endpoint) error {
	for _, e := range ee {
		z := Endpoint(e)
		r.l[e.ID] = &z
	}
	return nil
}

func (r *repository) Load(id string) Endpoint {
	e, ok := r.l[id]
	if !ok {
		return Endpoint{}
	}
	return *e
}

func (r *repository) Delete(id string) error {
	delete(r.l, id)
	return nil
}

func (r *repository) All() []Endpoint {
	var ee []Endpoint
	for _, e := range r.l {
		//fmt.Println(e)
		ee = append(ee, *e)
	}

	return ee
}

func (r *repository) StorePayload(p Definition) error {
	return nil
}

func (r *repository) Payload(endpoint string) Definition {
	return Definition{}
}
