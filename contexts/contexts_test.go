package contexts_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/deweppro/go-utils/contexts"
)

func TestUnit_Combine(t *testing.T) {
	cc, cancel := context.WithCancel(context.Background())
	c := contexts.Combine(context.Background(), context.Background(), cc)
	if c == nil {
		t.Fatalf("contexts.Combine returned nil")
	}

	select {
	case <-c.Done():
		t.Fatalf("<-c.Done() == it should block")
	default:
	}

	cancel()
	<-time.After(time.Second)

	select {
	case <-c.Done():
	default:
		t.Fatalf("<-c.Done() it shouldn't block")
	}

	if got, want := fmt.Sprint(c), "context.Background.WithCancel"; got != want {
		t.Fatalf("contexts.Combine() = %q want %q", got, want)
	}
}
