package inputs

import (
	"bytes"
	"io"
	"net/http"
)

type TestList interface {
	FeedJobs(chan map[string]interface{})
}

func getReaderFromUrl(url string) (*bytes.Reader, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf := &bytes.Buffer{}

	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}
