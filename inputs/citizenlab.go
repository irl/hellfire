package inputs

import (
	"fmt"
	"log"
)

type CitizenLabCountryList struct {
	TestList
	country string
}

// URL format string to download the latest Citizen Lab test list from. The %s
// will be replaced with either the two letter country code, or with "global"
// as specified in the call to SetCountry before the job feeder is activated.
const CitizenLabCountryListURL string = "https://raw.githubusercontent.com/citizenlab/test-lists/master/lists/%s.csv"

// The SetCountry method allows selection of the Citizen Lab test list to use.
// The full list of countries available can be found at
// https://github.com/citizenlab/test-lists/tree/master/lists. This method will
// also accept "global" as a country name, selecting the global test list.
func (l *CitizenLabCountryList) SetCountry(country string) {
	// BUG(irl): Make lowercase
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
