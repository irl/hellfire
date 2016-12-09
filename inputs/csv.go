package inputs

import (
	"bufio"
	"encoding/csv"
	"io"
	"net/url"
	"os"
)

type CSVList struct {
	TestList
	reader io.Reader
	header []string
}

func CSVListFromFile(filename string) *CSVList {
	f, err := os.Open(filename)
	if err != nil {
		panic("Error opening file")
	}
	return CSVListFromReader(f)
}

func CSVListFromReader(reader io.Reader) *CSVList {
	l := new(CSVList)
	l.reader = reader
	return l
}

func (l *CSVList) SetHeader(header []string) {
	l.header = header
}

func (l *CSVList) FeedJobs(jobs chan map[string]interface{}) {
	reader := csv.NewReader(bufio.NewReader(l.reader))
	if reader == nil {
		panic("CSVList not initialised with a reader")
	}
	var header []string
	if l.header == nil {
		var err error
		header, err = reader.Read()
		if err != nil {
			panic("Error reading the header from the CSV")
		}
	} else {
		header = l.header
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		r := make(map[string]interface{})
		for idx, name := range header {
			r[name] = record[idx]
		}
		if r["domain"] == nil && r["url"] != nil {
			u, _ := url.Parse(r["url"].(string))
			r["domain"] = u.Host
		}
		jobs <- r
	}
}
