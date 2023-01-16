package shell_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/deweppro/go-utils/shell"
)

func TestUnit_ShellCall(t *testing.T) {
	sh := shell.New()
	sh.SetDir("/tmp")
	sh.SetEnv("LANG", "en_US.UTF-8")

	out, err := sh.Call(context.TODO(), "ping -c 2 google.ru")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(string(out))
}

func TestUnit_ShellCallParallelContext(t *testing.T) {
	out := &bytes.Buffer{}

	sh := shell.New()
	sh.SetDir("/tmp")
	sh.SetEnv("LANG", "en_US.UTF-8")
	sh.SetWriter(out)
	err := sh.CallParallelContext(context.TODO(), "ping -c 2 google.ru", "ping -c 2 yandex.ru", "ls -la")
	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Println(out.String())
}
