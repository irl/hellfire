// Package inputs provides means of importing reconnaissance missions into
// hellfire. A number of input formats can be interpreted, including missions
// formatted using CSV and JSON schemas.
//
// Any method of importing missions must implement the TestList interface.
package inputs

import (
	"bytes"
	"io"
	"net/http"
)

// The TestList interface describes the methods that are used by hellfire
// to import missions.
type TestList interface {
	// The FeedJobs method should submit jobs, one at a time, into the
	// chan that has been passed to it. The map must contain one of the
	// keys "domain" or "url" that is either a fully-qualified domain
	// name or a URL with a fully-qualified domain name in the host
	// portion.
	//
	// If there is no value set for the "domain" key, the host portion of
	// the URL will be used for the lookup. If there is both a value for
	// "domain" and "url", the "url" value will be ignored and the "domain"
	// value used directly.
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
