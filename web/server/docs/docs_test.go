package docs
//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"io"
//	"net/http"
//	"net/http/httptest"
//	"path/filepath"
//	"runtime"
//	"testing"
//
//	"github.com/gorilla/mux"
//	"github.com/kr/pretty"
//	"github.com/livechat/developers-platform-service/internal"
//	"github.com/teris-io/shortid"
//)
//
//var tc1 = []interface{}{
//	map[string]interface{}{
//		"one": "1",
//		"two": "2",
//		"three": map[string]interface{}{
//			"four": 4,
//			"five": 5,
//		},
//	},
//	map[string]interface{}{
//		"six":   "6",
//		"seven": 7,
//	},
//	map[string]interface{}{
//		"three": map[string]interface{}{
//			"four": 4,
//			"five": 5,
//			"eight": []interface{}{
//				map[string]interface{}{"nine": "nine", "ten": "10",},
//				map[string]interface{}{"nine": "9", "ten": "ten",},
//			},
//		},
//	},
//	map[string]interface{}{
//		"eleven": "11",
//		"thirteen": []interface{}{
//			"13",
//		},
//	},
//}
//var tc2 = map[string]interface{}{
//	"one": "1",
//	"two": 2,
//	"three": []interface{}{
//		map[string]interface{}{"four": 4.0, "five": true},
//		map[string]interface{}{"four": -4.0, "five": false},
//	},
//	"six": map[string]interface{}{
//		"seven": "seven",
//		"eight": uint(8),
//	},
//	"nine": []interface{}{10, 11, 12},
//	"ten": map[string]interface{}{
//		"eleven": "el",
//		"points": []interface{}{
//			map[string]interface{}{"lat": 12.0, "lon": 52.42},
//			map[string]interface{}{"lat": 43.0, "lon": 51.12},
//		},
//	},
//	"tricks": []map[string]interface{}{
//		{"width": 581, "height": 611},
//		{"width": 571, "height": 913},
//		{"width": 815, "height": 115},
//		{"width": 573, "height": 341},
//	},
//	"stuff": map[string]interface{}{
//		"foo": map[string]interface{}{
//			"bip": "BIP",
//			"bas": map[string]interface{}{
//				"bis": "BIS",
//				"sas": "SAS",
//				"bar": map[string]interface{}{
//					"bop": "BOP",
//					"bep": []interface{}{"a", "b"},
//				},
//			},
//		},
//	},
//}
//
//type node struct{ group, resource, action string }
//
//func TestName(t *testing.T) {
//	//
//	//type item struct {
//	//	Name string
//	//}
//	//
//	//type product struct {
//	//	ID   string
//	//	Name string
//	//	Item item
//	//}
//	//
//	////p1 := product{fake.CharactersN(8), fake.Product()}
//	//srv := mux.NewRouter()
//	//
//	//srv.HandleFunc("/v2/products", receiver()).Methods("GET")
//	//srv.HandleFunc("/v2/products/{product}", receiver()).Methods("DELETE")
//	//
//	//docs := Documentation()
//	//isOK(t, docs.Read(srv))
//	//
//	//call(srv, "/v2/products", "GET", product{})
//	//
//	//time.Sleep(time.Second)
//	//fmt.Printf("%#v\n", docs.Storage.Load("7dbda3ac8d2d39b7105d9553b799749d"))
//	//for _, d := range doc.Storage.All() {
//	//	fmt.Println(d.GoString())
//	//}
//
//	var (
//	//y = []interface{}{
//	//	map[string]interface{}{
//	//		"test": map[string]interface{}{
//	//			"four": 4,
//	//			"five": 5,
//	//		},
//	//	},
//	//}
//	//a = fields.build("stuff", tc1)
//	//b = variables.build("stuff", "struct", tc2)
//	)
//
//	//variables.merge(a, b)
//
//	//a = variables.merge(a, Variables{
//	//	"product": Variables{
//	//		"id": "z",
//	//	},
//	//})
//	//a = variables.merge(a, Variables{
//	//	"stuff": "a",
//	//	"product": Variables{
//	//		"id": "some description",
//	//	},
//	//})
//	//
//	//a = variables.merge(a, Variables{
//	//	"product": Variables{
//	//		"items": []string{
//	//			"a", "b",
//	//		},
//	//	},
//	//})
//	//fmt.Printf("\nmerged\n%s", a)
//
//	x := fields.build("application", tc2)
//	pretty.Println(x)
//
//}
//
//func TestObject(t *testing.T) {
//	c, err := dps.NewContainer("config.json", "1")
//	isOK(t, err)
//
//	var (
//		//inf  = c.Logger.Log.Info
//		o = c.Applications.Published(0)
//		//n    = time.Now()
//		b, _ = json.Marshal(o)
//		l    = bytes.NewBuffer(b)
//		v    = fields.fromReader(l, "application")
//		//v = fields.build("test", tc2)
//	)
//
//	//inf("%s", time.Since(n))
//	//inf("%s", fields.objects())
//
//	//pretty.Println(v)
//	pretty.Println(fields.json("application", v))
//	//pretty.Println(fields.objects(v))
//	//inf("%s", v.String())
//	//fields.objects(v)
//}
//
////func TestRead(t *testing.T) {
////	a := DocsEndpoint{
////		Headers: Variables{
////			"Test|array": "ok",
////			"a": "lo",
////		},
////	}
////	b, _ := json.MarshalIndent(a, "", "\t")
////	fmt.Printf("%T%s\n\n", a, b)
////
////	b, _ = json.Marshal(a)
////	o := Variables{}.read("application", bytes.NewReader(b))
////
////	fmt.Println("")
////	pretty.Println(o)
////	pretty.Println(o.String())
////	fmt.Print(markdown("Header", a.Headers))
////	fmt.Print(markdown("DUPA", a.Headers))
////}
//
//func TestGrouping(t *testing.T) {
//
//	tc := []string{
//		"/v2/:GET",
//		"/v2/applications/algolia:POST",
//		"/v2/applications/all:GET",
//		"/v2/applications/installed/{license}:GET",
//		"/v2/applications/installed:GET",
//		"/v2/applications/published:GET",
//		"/v2/applications/submitted:GET",
//		"/v2/applications/{application}/authorization/replace:PUT",
//		"/v2/applications/{application}/authorization:PUT",
//		"/v2/applications/{application}/benefits:PUT",
//		"/v2/applications/{application}/buy:PUT",
//		"/v2/applications/{application}/clone:POST",
//		"/v2/applications/{application}/custom-url:PATCH",
//		"/v2/applications/{application}/display/details:PUT",
//		"/v2/applications/{application}/display:PUT",
//		"/v2/applications/{application}/elements/buttons/:POST",
//		"/v2/applications/{application}/elements/buttons/{button}:DELETE",
//		"/v2/applications/{application}/features:PUT",
//		"/v2/applications/{application}/image:POST",
//		"/v2/applications/{application}/inspect:DELETE",
//		"/v2/applications/{application}/inspect:PUT",
//		"/v2/applications/{application}/install/mark:PATCH",
//		"/v2/applications/{application}/install:DELETE",
//		"/v2/applications/{application}/install:PUT",
//		"/v2/applications/{application}/my:GET",
//		"/v2/applications/{application}/owner/{developer}:PUT",
//		"/v2/applications/{application}/payment:PUT",
//		"/v2/applications/{application}/publish:DELETE",
//		"/v2/applications/{application}/publish:PUT",
//		"/v2/applications/{application}/published:GET",
//		"/v2/applications/{application}/reviews/accept:PATCH",
//		"/v2/applications/{application}/reviews:GET",
//		"/v2/applications/{application}/reviews:PUT",
//		"/v2/applications/{application}/screenshots/commit:PUT",
//		"/v2/applications/{application}/slug/{slug}:PATCH",
//		"/v2/applications/{application}/stats:GET",
//		"/v2/applications/{application}/submission/{type}:PUT",
//		"/v2/applications/{application}/submission:DELETE",
//		"/v2/applications/{application}/submission:POST",
//		"/v2/applications/{application}/theme:PUT",
//		"/v2/applications/{application}/visibility/hide:PUT",
//		"/v2/applications/{application}/visibility/show:PUT",
//		"/v2/applications/{application}/webhook:PUT",
//		"/v2/applications/{application}/widgets/{widget}:DELETE",
//		"/v2/applications/{application}/widgets:POST",
//		"/v2/applications/{application}:DELETE",
//		"/v2/applications:GET",
//		"/v2/applications:POST",
//		"/v2/campaigns/all:GET",
//		"/v2/campaigns/{campaign}:DELETE",
//		"/v2/campaigns:GET",
//		"/v2/campaigns:PUT",
//		"/v2/categories/{category}:DELETE",
//		"/v2/categories/{category}:PUT",
//		"/v2/categories:GET",
//		"/v2/categories:POST",
//		"/v2/developers/:GET",
//		"/v2/developers/recent:GET",
//		"/v2/developers/{id}/:GET",
//		"/v2/developers/{id}/activate/{on|off}:PATCH",
//		"/v2/docs/activate/{on-off}:PATCH",
//		"/v2/docs/readme.html:GET",
//		"/v2/docs/readme.md:GET",
//		"/v2/docs/{endpoint}/{object}/{field}:PATCH",
//		"/v2/docs/{endpoint}:PATCH",
//		"/v2/docs:GET",
//		"/v2/events:GET",
//		"/v2/me:GET",
//		"/v2/me:PUT",
//		"/v2/recent:GET",
//		"/v2/register:POST",
//		"/v2/reviews:GET",
//		"/v2/survey:POST",
//	}
//	shortid.Generate()
//	//fmt.Println(hash("Test"))
//	fmt.Println(tc)
//}
//
////  ok fails the test if an err is not nil.
//func isOK(tb testing.TB, err error) {
//	if err != nil {
//		_, file, line, _ := runtime.Caller(1)
//		fmt.Printf(
//			"\033[31m%s:%d: unexpected error: %s\033[39m\n\n",
//			filepath.Base(file), line, err.Error())
//		tb.FailNow()
//	}
//}
//
//func receiver() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		io.Copy(w, r.Body)
//	}
//}
//
//func call(s *mux.Router, path, method string, v interface{}) {
//	var buff *bytes.Buffer
//	if v != nil {
//		z, _ := json.Marshal(v)
//		buff = bytes.NewBuffer(z)
//	}
//	req, _ := http.NewRequest(method, path, buff)
//
//	res := httptest.NewRecorder()
//	s.ServeHTTP(res, req)
//}
