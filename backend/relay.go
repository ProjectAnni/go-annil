package backend

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type RelayBackend struct {
	Url   string
	Token string
}

func NewRelay(url string, token string) *RelayBackend {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return &RelayBackend{
		Url:   url,
		Token: token,
	}
}

func (e *RelayBackend) ListCatalogs() []string {
	req, err := http.NewRequest(http.MethodGet, e.Url+"albums", nil)
	ret := make([]string, 0)
	if err != nil {
		return ret
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ret
	}
	if res.StatusCode != http.StatusOK {
		return ret
	}
	defer res.Body.Close()
	_ = json.NewDecoder(res.Body).Decode(&ret)
	return ret
}

func (e *RelayBackend) GetCover(catalog string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, e.Url+catalog+"/cover", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", e.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("invalid response code")
	}
	return res.Body, nil
}

func (e *RelayBackend) GetAudio(catalog string, track uint8) (AudioType, io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, e.Url+catalog+"/"+strconv.Itoa(int(track)), nil)
	if err != nil {
		return UNKNOWN, nil, err
	}
	req.Header.Set("Authorization", e.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return UNKNOWN, nil, err
	}
	if res.StatusCode != http.StatusOK {
		return UNKNOWN, nil, errors.New("invalid response code")
	}
	t := UNKNOWN
	if res.Header.Get("Content-Type") == "audio/flac" {
		t = FLAC
	} else if res.Header.Get("Content-Type") == "audio/mp3" {
		t = MP3
	}
	return t, res.Body, nil
}
