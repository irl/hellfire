package inputs

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"strconv"
)

type AlexaTopsitesList string

const AlexaTopsitesURL string = "http://s3.amazonaws.com/alexa-static/top-1m.csv.zip"

func (l *AlexaTopsitesList) FeedJobs(jobs chan map[string]interface{}) {
	var topsites *csv.Reader

	urlReader, err := getReaderFromUrl(AlexaTopsitesURL)
	if err != nil {
		log.Fatalf("Unable to get <%s>: %s", AlexaTopsitesURL, err)
	}

	zr, err := zip.NewReader(urlReader, int64(urlReader.Len()))
	if err != nil {
		log.Fatalf("Unable to read zip: %s", err)
	}

	for _, zf := range zr.File {
		if zf.Name == "top-1m.csv" {
			f, _ := zf.Open()
			topsites = csv.NewReader(bufio.NewReader(f))
		}
	}
	if topsites == nil {
		panic("There was no top-1m.csv in the zip archive.")
	}
	for {
		record, err := topsites.Read()
		if err == io.EOF {
			break
		}

		r := make(map[string]interface{})
		r["domain"] = record[1]
		r["rank"], _ = strconv.Atoi(record[0])
		jobs <- r
	}
}
