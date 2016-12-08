package inputs

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/url"
)

type CitizenLabCountryList struct {
	TestList
	country string
}

const CitizenLabCountryListURL string = "https://raw.githubusercontent.com/citizenlab/test-lists/master/lists/%s.csv"

func (l *CitizenLabCountryList) SetCountry(country string) {
	// FIXME: Make lowercase
	if country == "global" || len(country) == 2 {
		l.country = country
	} else {
		panic("Country code must be two characters, or 'global'.")
	}
}

func (l *CitizenLabCountryList) FeedJobs(jobs chan map[string]interface{}) {
	listUrl := fmt.Sprintf(CitizenLabCountryListURL, l.country)
	urlReader, err := getReaderFromUrl(listUrl)
	if err != nil {
		log.Fatalf("Unable to get <%s>: %s", listUrl, err)
	}

	r := csv.NewReader(bufio.NewReader(urlReader))
	header, err := r.Read()
	if err != nil || len(header) < 2 {
		panic("Error reading the header from the CSV (There may not be a list for the specified country)")
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		r := make(map[string]interface{})
		fmt.Println("Record length is ", len(record))
		for idx, name := range header {
			r[name] = record[idx]
		}

		u, _ := url.Parse(r["url"].(string))
		r["domain"] = u.Host

		jobs <- r
	}
}
