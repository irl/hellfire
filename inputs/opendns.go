package inputs // import "pathspider.net/hellfire/inputs"

import (
	"fmt"
	"log"
	"strings"
)

type OpenDNSList struct {
	TestList
	listname string
	filename string
}

// URL format string to download the OpenDNS public domain list from. The %s
// will be replaced with the name of the list as specified in the call to
// SetListName before the job feeder is activated.
const OpenDNSListURL string = "https://raw.githubusercontent.com/opendns/public-domain-lists/master/opendns-%s-domains.txt"

func (l *OpenDNSList) SetFilename(filename string) {
	l.filename = filename
}

// The SetListName method allows selection of the OpenDNS list to use. This
// function accepts either "top" or "random" as the list name.
//
// This function must be called before FeedJobs() or the application will panic.
func (l *OpenDNSList) SetListName(listname string) {
	listname = strings.ToLower(listname)
	if listname == "top" || listname == "random" {
		l.listname = listname
	} else {
		panic("List name must be either \"top\" or \"random\".")
	}
}

func (l *OpenDNSList) FeedJobs(jobs chan map[string]interface{}) {
	var openDNSList *CSVList

	if l.filename == "" {
		if l.listname == "" {
			panic("The list name to use was not specified.")
		}
		listUrl := fmt.Sprintf(OpenDNSListURL, l.listname)
		urlReader, err := getReaderFromUrl(listUrl)
		if err != nil {
			log.Fatalf("Unable to get <%s>: %s", listUrl, err)
		}

		openDNSList = CSVListFromReader(urlReader)
	} else {
		openDNSList = CSVListFromFile(l.filename)
	}
	openDNSList.SetHeader([]string{"domain"})
	openDNSList.FeedJobs(jobs)
}
