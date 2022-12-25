package ioutils

import (
	"io"

	errors "github.com/deweppro/go-errors"
)

func ReadAll(r io.ReadCloser) ([]byte, error) {
	b, err := io.ReadAll(r)
	err = errors.Wrap(err, r.Close())
	if err != nil {
		return nil, err
	}
	return b, nil
}
