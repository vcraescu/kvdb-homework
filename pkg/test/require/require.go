package require

import (
	"fmt"
	"reflect"
	"testing"
)

func Equal(t *testing.T, expected, actual interface{}, msg ...interface{}) {
	t.Helper()

	if reflect.DeepEqual(actual, expected) {
		return
	}

	if len(msg) > 0 {
		t.Fatalf("\nGot\n\t%#v\nWant\n\t%#v\nMessage\n\t%s", actual, expected, fmt.Sprint(msg...))
	}

	t.Fatalf("\nGot\n\t%#v\nWant\n\t%#v", actual, expected)
}

func Error(t *testing.T, err error, msg ...interface{}) {
	t.Helper()

	if err != nil {
		return
	}

	Equal(t, err, nil, msg...)
}

func NoError(t *testing.T, err error, msg ...interface{}) {
	t.Helper()

	if err == nil {
		return
	}

	Equal(t, nil, err, msg...)
}

func True(t *testing.T, actual bool, msg ...interface{}) {
	t.Helper()

	Equal(t, true, actual, msg...)
}

func False(t *testing.T, actual bool, msg ...interface{}) {
	t.Helper()

	Equal(t, false, actual, msg...)
}
