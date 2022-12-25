package random_test

import (
	"bytes"
	"testing"

	"github.com/deweppro/go-utils/random"
)

func TestUnit_Bytes(t *testing.T) {
	r1 := random.Bytes(5)
	r2 := random.Bytes(5)

	if len(r1) != 5 || len(r2) != 5 {
		t.Errorf("invalid len, is not 5")
	}
	if bytes.Equal(r1, r2) {
		t.Errorf("result is not random")
	}
}
