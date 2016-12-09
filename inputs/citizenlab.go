package inputs

import (
	"fmt"
	"log"
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

	citizenLabList := CSVListFromReader(urlReader)
	citizenLabList.FeedJobs(jobs)
}
