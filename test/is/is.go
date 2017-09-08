package is

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func True(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf(
			"\033[31m%s:%d: "+msg+"\033[39m\n\n",
			append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

//  ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf(
			"\033[31m%s:%d: unexpected error: %s\033[39m\n\n",
			filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func Equal(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf(
			"\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n",
			filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func Err(tb testing.TB, err error, msg string) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf(
			"\033[31m%s:%d: error %s is expected\033[39m\n\n ",
			filepath.Base(file), line, msg)
		tb.FailNow()
	}
}

func NotErr(tb testing.TB, err error) {
	Ok(tb, err)
}
