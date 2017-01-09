package inputs // import "pathspider.net/hellfire/inputs"

import (
	"archive/zip"
	"log"
)

type CiscoUmbrellaList struct {
	TestList
}

// URL to download the latest Cisco Umbrella list from
const CiscoUmbrellaURL string = "http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip"

func (l *CiscoUmbrellaList) FeedJobs(jobs chan map[string]interface{}) {
	var topsites *CSVList

	urlReader, err := getReaderFromUrl(CiscoUmbrellaURL)
	if err != nil {
		log.Fatalf("Unable to get <%s>: %s", CiscoUmbrellaURL, err)
	}

	zr, err := zip.NewReader(urlReader, int64(urlReader.Len()))
	if err != nil {
		log.Fatalf("Unable to read zip: %s", err)
	}

	for _, zf := range zr.File {
		if zf.Name == "top-1m.csv" {
			f, _ := zf.Open()
			topsites = CSVListFromReader(f)
			break
		}
	}

	if topsites == nil {
		panic("Did not find top-1m.csv in the zip archive")
	}

	topsites.SetHeader([]string{"rank", "domain"})
	topsites.FeedJobs(jobs)
}
