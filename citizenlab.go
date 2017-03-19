package hellfire // import "pathspider.net/hellfire"

import (
	"fmt"
	"log"
	"strings"
)

type CitizenLabCountryList struct {
	TestList
	country  string
	filename string
}

// URL format string to download the latest Citizen Lab test list from. The %s
// will be replaced with either the two letter country code, or with "global"
// as specified in the call to SetCountry before the job feeder is activated.
const CitizenLabCountryListURL string = "https://raw.githubusercontent.com/citizenlab/test-lists/master/lists/%s.csv"

func (l *CitizenLabCountryList) SetFilename(filename string) {
	l.filename = filename
}

// The SetCountry method allows selection of the Citizen Lab test list to use.
// The full list of countries available can be found at
// https://github.com/citizenlab/test-lists/tree/master/lists. This method will
// also accept "global" as a country name, selecting the global test list.
//
// This function must be called before FeedJobs() or the application will panic.
func (l *CitizenLabCountryList) SetCountry(country string) {
	country = strings.ToLower(country)
	if country == "global" || len(country) == 2 {
		l.country = country
	} else {
		panic("Country code must be two characters, or 'global'.")
	}
}

func (l *CitizenLabCountryList) FeedJobs(jobs chan map[string]interface{}) {
	var citizenLabList *CSVList

	if l.filename == "" {
		if l.country == "" {
			panic("The country to use for the Citizen Lab test was not specified")
		}
		listUrl := fmt.Sprintf(CitizenLabCountryListURL, l.country)
		urlReader, err := getReaderFromUrl(listUrl)
		if err != nil {
			log.Fatalf("Unable to get <%s>: %s", listUrl, err)
		}

		citizenLabList = CSVListFromReader(urlReader)
	} else {
		// Note that country codes starting with X are "private use"
		// and XF in this case is to indicate that a file was used.
		// BUG(irl): Maybe a hint could be provided on the command line
		// later.
		l.SetCountry("xf")
		citizenLabList = CSVListFromFile(l.filename)
	}
	citizenLabList.FeedJobs(jobs)
}
