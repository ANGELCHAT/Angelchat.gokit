package client

import (
	"encoding/json"
	"io"
)

// Encoder is used to transform v into io.Writer, usually that transformation is used when we want to pass
// some data into Endpoint, which through Do method send that (v) information across network.
type Encoder func(w io.Writer, v interface{}) error

// jsonEncode is a default golang implementation of struct (v) into JSON string (w)
func jsonEncode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
