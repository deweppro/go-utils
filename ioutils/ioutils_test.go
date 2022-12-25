package ioutils_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	errors "github.com/deweppro/go-errors"
	"github.com/deweppro/go-utils/ioutils"
)

type mockReadCloser struct {
	Data     *bytes.Buffer
	ErrRead  error
	ErrClose error
}

func (v *mockReadCloser) Read(p []byte) (int, error) {
	if v.ErrRead != nil {
		return 0, v.ErrRead
	}
	return v.Data.Read(p)
}

func (v *mockReadCloser) Close() error {
	return v.ErrClose
}

func TestUnit_ReadAll(t *testing.T) {
	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Case1",
			args: args{
				r: &mockReadCloser{
					Data:     bytes.NewBuffer([]byte(`hello`)),
					ErrRead:  nil,
					ErrClose: nil,
				},
			},
			want:    []byte(`hello`),
			wantErr: false,
		},
		{
			name: "Case2",
			args: args{
				r: &mockReadCloser{
					Data:     nil,
					ErrRead:  errors.New("read error"),
					ErrClose: nil,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Case3",
			args: args{
				r: &mockReadCloser{
					Data:     nil,
					ErrRead:  errors.New("read error"),
					ErrClose: errors.New("close error"),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Case4",
			args: args{
				r: &mockReadCloser{
					Data:     bytes.NewBuffer([]byte(`hello`)),
					ErrRead:  nil,
					ErrClose: errors.New("close error"),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ioutils.ReadAll(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}
