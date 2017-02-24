package apk

import (
	"io/ioutil"
	"testing"
)

func TestUnmarshalArsc(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/resources.arsc")
	if err != nil {
		t.Fatal(err)
	}
	if err = UnmarshalArsc(data); err != nil {
		t.Fatal(err)
	}
}
