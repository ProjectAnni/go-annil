package backend

import (
    "bytes"
    "io"
    "io/ioutil"
)

type AudioType uint8

const (
    FLAC    AudioType = 0
    MP3     AudioType = 1
    UNKNOWN AudioType = 2
)

type Backend struct {
}

func (*Backend) ListCatalogs() []string {
    return []string{}
}

func (*Backend) GetCover(catalog string) (io.ReadCloser, error) {
    return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
}

func (*Backend) GetAudio(catalog string, track uint8) (AudioType, io.ReadCloser, error) {
    return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), nil
}
