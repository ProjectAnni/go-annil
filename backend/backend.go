package backend

import (
	"io"
)

type AudioType uint8

const (
	FLAC    AudioType = 0
	MP3     AudioType = 1
	UNKNOWN AudioType = 2
)

type Backend interface {
	ListCatalogs() []string
	GetCover(catalog string) (io.ReadCloser, error)
	GetAudio(catalog string, track uint8) (AudioType, io.ReadCloser, error)
}
