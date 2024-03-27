package fs

import "io"

type Storage interface {
	GetURL(k string) (string, error)
	Put(k string, v io.ReadSeeker) error
	Delete(k string) error
}
