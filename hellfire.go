// Hellfire is a parallelised DNS resolver. It builds effects lists for input
// to PATHspider measurements. For sources where the filename is optional, the
// latest source will be downloaded from the Internet when the filename is
// omitted.
//
// BASIC USAGE
//
//  Usage:
//    hellfire --topsites [--file=<filename>] [--all]
//    hellfire --cisco [--file=<filename>] [--all]
//    hellfire --citizenlab (--country=<cc>|--file=<filename>) [--all]
//    hellfire --opendns (--list=<name>|--file=<filename>) [--all]
//    hellfire --csv --file=<filename> [--all]
//    hellfire --txt --file=<filename> [--all]
//
//  Options:
//    -h --help     Show this screen.
//    --version     Show version.
//
// SEE ALSO
//
// The PATHspider website can be found at https://pathspider.net/.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/docopt/docopt-go"
	"net"
	"pathspider.net/hellfire/inputs"
	"strings"
	"sync"
	"time"
)

func main() {
	usage := `Hellfire: PATHspider Effects List Resolver

Hellfire is a parallelised DNS resolver. It builds effects lists for input to
PATHspider measurements. For sources where the filename is optional, the latest
source will be downloaded from the Internet when the filename is omitted.

Usage:
  hellfire --topsites [--file=<filename>] [--all]
  hellfire --cisco [--file=<filename>] [--all]
  hellfire --citizenlab (--country=<cc>|--file=<filename>) [--all]
  hellfire --opendns (--list=<name>|--file=<filename>) [--all]
  hellfire --csv --file=<filename> [--all]
  hellfire --txt --file=<filename> [--all]

Options:
  -h --help     Show this screen.
  --version     Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, "Hellfire dev", false)

	var testList inputs.TestList
	//BUG(irl): Filenames are ignored
	//BUG(irl): CSV type is ignored
	//BUG(irl): TXT type is ignored
	if arguments["--topsites"].(bool) {
		testList = new(inputs.AlexaTopsitesList)
	} else if arguments["--cisco"].(bool) {
		testList = new(inputs.CiscoUmbrellaList)
	} else if arguments["--citizenlab"].(bool) {
		testList = new(inputs.CitizenLabCountryList)
		if arguments["--country"] != nil {
			testList.(*inputs.CitizenLabCountryList).SetCountry(arguments["--country"].(string))
		}
	} else if arguments["--opendns"].(bool) {
		testList = new(inputs.OpenDNSList)
		if arguments["--list"] != nil {
			testList.(*inputs.OpenDNSList).SetListName(arguments["--list"].(string))
		}
	}

	if testList != nil {
		if arguments["--file"] != nil {
			testList.SetFilename(arguments["--file"].(string))
		}
		performLookups(testList, arguments["--all"].(bool))
	} else {
		panic("An error occured building the input provider")
	}
}

func lookupWorker(id int, lookupWaitGroup *sync.WaitGroup,
	jobs chan map[string]interface{},
	results chan map[string]interface{}) {
	lookupWaitGroup.Add(1)
	go func(id int, lookupWaitGroup *sync.WaitGroup, jobs chan map[string]interface{}, results chan map[string]interface{}) {
		defer lookupWaitGroup.Done()
		for job := range jobs {
			if job["domain"] == nil {
				jobs <- make(map[string]interface{})
				break
			}
			var attempt int
			var ips []net.IP
			for attempt = 1; ; attempt++ {
				ips, _ = net.LookupIP(job["domain"].(string))
				if len(ips) == 0 {
					time.Sleep(1)
				} else {
					break
				}
				if attempt == 3 {
					break
				}
			}
			job["ips"] = ips
			job["attempts"] = attempt
			results <- job
		}
	}(id, lookupWaitGroup, jobs, results)
}

func outputPrinter(outputWaitGroup *sync.WaitGroup, results chan map[string]interface{}, printAllResults bool) {
	outputWaitGroup.Add(1)
	go func(results chan map[string]interface{}) {
		defer outputWaitGroup.Done()
		for {
			result := <-results
			if result["domain"] == nil {
				break
			}
			if printAllResults {
				b, _ := json.Marshal(result)
				fmt.Println(string(b))
			} else {
				found4 := false
				found6 := false
				ips := result["ips"].([]net.IP)
				delete(result, "ips")
				for _, ipo := range ips {
					ip := ipo.String()
					if strings.Contains(ip, ".") {
						if found4 {
							continue
						} else {
							found4 = true
						}
					} else {
						if found6 {
							continue
						} else {
							found6 = true
						}
					}
					result["ip"] = ip
					b, _ := json.Marshal(result)
					fmt.Println(string(b))
					delete(result, "ip")
				}
			}
		}
	}(results)
}

func performLookups(testList inputs.TestList, printAllResults bool) {
	var lookupWaitGroup sync.WaitGroup
	var outputWaitGroup sync.WaitGroup
	jobs := make(chan map[string]interface{}, 1)
	results := make(chan map[string]interface{})

	// Spawn lookup workers
	for i := 0; i < 300; i++ {
		lookupWorker(i, &lookupWaitGroup, jobs, results)
	}

	// Spawn output printer
	outputPrinter(&outputWaitGroup, results, printAllResults)

	// Submit jobs
	testList.FeedJobs(jobs)
	jobs <- make(map[string]interface{})
	lookupWaitGroup.Wait()
	<-jobs // Read last shutdown sentinel from the queue left by the
	// final worker to exit
	// https://blog.golang.org/pipelines - This is a better way
	close(jobs)

	// Shutdown the output printer
	results <- make(map[string]interface{})
	outputWaitGroup.Wait()
	close(results)
}
