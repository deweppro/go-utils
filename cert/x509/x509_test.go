package x509_test

import (
	"testing"
	"time"

	"github.com/deweppro/go-utils/cert/x509"
)

func TestUnit_X509(t *testing.T) {
	conf := &x509.Config{
		Organization: "DewepPro Test Inc.",
	}

	crt, err := x509.NewCertCA(conf, time.Hour*24*365*10, "DewepPro Root R1")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(string(crt.Private), string(crt.Public))

	crt, err = x509.NewCert(conf, time.Hour*24*90, 2, crt, "dewep.pro", "*.dewep.pro")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(string(crt.Private), string(crt.Public))
}
