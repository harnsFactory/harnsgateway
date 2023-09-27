package randutil

import (
	"testing"
)

func TestInt63n(t *testing.T) {
	expect := Int63n()

	// time.Sleep(time.Second)

	actual := Int63n()

	if expect == actual {
		t.Errorf("actual %v, expect %v", actual, expect)
	}
}

func TestStringN(t *testing.T) {
	expect := StringN(2)

	// time.Sleep(time.Second)

	actual := StringN(2)

	if expect == actual {
		t.Errorf("actual %v, expect %v", actual, expect)
	}
}