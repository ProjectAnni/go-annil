package backend

import (
    "bytes"
    "errors"
    "io"
    "io/ioutil"
    "os"
    "path"
    "strconv"
    "strings"
)

type FileBackend struct {
    Backend
    rootDir string
}

func NewFileBackend(pathIn string) (*FileBackend, error) {
    f, err := os.Open(pathIn)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    info, err := f.Stat()
    if err != nil {
        return nil, err
    }
    if !info.IsDir() {
        return nil, errors.New("not a directory")
    }
    return &FileBackend{rootDir: path.Clean(pathIn)}, nil
}

func (b *FileBackend) ListCatalogs() []string {
    f, err := os.Open(b.rootDir)
    if err != nil {
        return []string{}
    }
    defer f.Close()
    catalogs, err := f.Readdirnames(0)
    return catalogs
}

func (b *FileBackend) GetCover(catalog string) (io.ReadCloser, error) {
    f, err := os.Open(b.rootDir + "/" + catalog + "/cover.jpg")
    if err != nil {
        return ioutil.NopCloser(bytes.NewReader([]byte{})), err
    }
    return f, nil
}

func (b *FileBackend) GetAudio(catalog string, track uint8) (AudioType, io.ReadCloser, error) {
    dir, err := os.Open(b.rootDir + "/" + catalog)
    if err != nil {
        return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), err
    }

    info, err := dir.Stat()
    if err != nil {
        return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), err
    }

    if !info.IsDir() {
        return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), errors.New("not a directory")
    }

    names, err := dir.Readdirnames(0)
    if err != nil {
        return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), err
    }
    var name = ""

    var prefix = ""
    if track < 10 {
        prefix = "0" + strconv.Itoa(int(track))
    } else {
        prefix = strconv.Itoa(int(track))
    }

    for _, n := range names {
        if strings.HasPrefix(n, prefix) {
            name = n
        }
    }

    if name == "" {
        return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), err
    }

    audType := UNKNOWN
    if strings.HasSuffix(name, ".flac") {
        audType = FLAC
    } else if strings.HasSuffix(name, ".mp3") {
        audType = MP3
    }

    f, err := os.Open(b.rootDir + "/" + catalog + "/" + name)
    if err != nil {
        return UNKNOWN, ioutil.NopCloser(bytes.NewReader([]byte{})), err
    }
    return audType, f, nil
}
