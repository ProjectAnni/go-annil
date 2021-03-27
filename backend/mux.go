package backend

import (
	"errors"
	"io"
)

func NewMultiplexer(backends []Backend) *Multiplexer {
	return &Multiplexer{Backends: backends}
}

type Multiplexer struct {
	Backends []Backend
}

func (be *Multiplexer) ListCatalogs() []string {
	m := make(map[string]bool)
	for _, b := range be.Backends {
		for _, cat := range b.ListCatalogs() {
			m[cat] = true
		}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (be *Multiplexer) GetCover(catalog string) (io.ReadCloser, error) {
	for _, b := range be.Backends {
		c, err := b.GetCover(catalog)
		if err == nil {
			return c, nil
		}
	}
	return nil, errors.New("no available")
}

func (be *Multiplexer) GetAudio(catalog string, track uint8) (AudioType, io.ReadCloser, error) {
	for _, b := range be.Backends {
		t, c, err := b.GetAudio(catalog, track)
		if err == nil {
			return t, c, nil
		}
	}
	return UNKNOWN, nil, errors.New("no available")
}
