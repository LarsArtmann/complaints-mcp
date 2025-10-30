// Package afero provides a filesystem in memory
package afero

import (
	"os"
	"time"
)

type AferoFs struct {
	fs *MemMapFs
}

func New() *AferoFs {
	return &AferoFs{fs: NewMemMapFs()}
}