package pgp_test

import (
	"bytes"
	"crypto"
	"testing"

	"github.com/deweppro/go-utils/cert/pgp"
)

func TestUnit_PGP(t *testing.T) {
	conf := pgp.Config{
		Name:    "Test Name",
		Email:   "Test Email",
		Comment: "Test Comment",
	}
	crt, err := pgp.NewCert(conf, crypto.MD5, 1024, "tool", "dewep utils")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(string(crt.Private), string(crt.Public))

	in := bytes.NewBufferString("Hello world")
	out := &bytes.Buffer{}

	sig := pgp.New()
	if err = sig.SetKey(crt.Private, ""); err != nil {
		t.Fatalf(err.Error())
	}
	sig.SetHash(crypto.MD5, 1024)
	if err = sig.Sign(in, out); err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(out.String())
}
