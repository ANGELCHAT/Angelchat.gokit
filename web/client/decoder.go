package client

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"
)

// Decoder is reading output of Endpoint::Do and then transform it to given v
type Decoder func(r io.Reader, v interface{}) error

// JsonDecode is golang default transformation of json string (r) into given struct (v)
func JsonDecode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func CsvDecodeNative(r io.Reader, v interface{}) error {
	e, ok := v.(CSVRecord)
	if !ok {
		return errors.New("csv: type " + reflect.TypeOf(v).String() + " does" +
			" not implement Row(map[string]string)")
	}

	c := csv.NewReader(r)

	// read header
	h, err := c.Read()
	if err != nil {
		return err
	}

	// read rows
	for {
		rec, err := c.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.New("csv: unexpected error while reading content")
		}

		l := map[string]string{}
		for i, k := range h {
			l[k] = rec[i]
		}

		e.Row(l)
	}
	return nil
}

func CsvDecodeScannerAsync(r io.Reader, v interface{}) error {
	e, ok := v.(CSVRecord)
	if !ok {
		return errors.New("csv: type " + reflect.TypeOf(v).String() + " does" +
			" not implement Row(map[string]string)")
	}

	for l := range newCSV(r).ReadAsync() {
		e.Row(l)
	}

	return nil
}

// CSVRecord
type CSVRecord interface {
	Row(v map[string]string)
}

type csvReader struct {
	Scanner *bufio.Scanner
	m       sync.Mutex
}

func (r csvReader) ReadAsync() <-chan map[string]string {
	c := make(chan map[string]string)

	var wg sync.WaitGroup
	r.Scanner.Scan()
	t := r.Scanner.Text()
	h := strings.Split(t, `,`)

	for r.Scanner.Scan() {
		wg.Add(1)
		go func(b string) {
			defer wg.Done()
			o := strings.Split(b, `,`)
			x := make(map[string]string)
			for i, k := range h {
				k = strings.Replace(k, `"`, "", -1)
				v := strings.Replace(o[i], `"`, "", -1)
				x[k] = v
			}
			c <- x
		}(r.Scanner.Text())
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	return c
}

func (r csvReader) Read() ([][]byte, error) {
	if r.Scanner.Scan() {
		b := r.Scanner.Bytes()
		l := len(b)
		if l <= 2 {
			return nil, errors.New("malformed csv")
		}

		return bytes.Split(b[1:l-1], []byte(`","`)), nil
	}
	if err := r.Scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

func newCSV(r io.Reader) csvReader {
	return csvReader{
		Scanner: bufio.NewScanner(r),
	}
}
