package slug

import "testing"

func TestMake(t *testing.T) {
	if Make("Hello World!") != "hello-world" {
		t.Fatal("slug mismatch")
	}
}
