package contexts

import (
	"context"
	"reflect"
)

func Combine(multi ...context.Context) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		cases := make([]reflect.SelectCase, 0, len(multi))
		for _, vv := range multi {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(vv.Done()),
			})
		}
		chosen, _, _ := reflect.Select(cases)
		switch chosen {
		default:
			cancel()
		}
	}()

	return ctx
}
