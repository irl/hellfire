
// Hellfire is a parallelised DNS resolver. It builds effects lists for input
// to PATHspider measurements. For sources where the filename is optional, the
// latest source will be downloaded from the Internet when the filename is
// omitted.
// 
// BASIC USAGE
//
//  Usage:
//    hellfire --topsites [--file=<filename>]
//    hellfire --cisco [--file=<filename>]
//    hellfire --citizenlab (--country=<cc>|--file=<filename>)
//    hellfire --opendns (--list=<name>|--file=<filename>)
//    hellfire --csv --file=<filename>
//    hellfire --txt --file=<filename>
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
	"fmt"
	"github.com/docopt/docopt-go"
	"net"
	"pathspider.net/hellfire/inputs"
	"sync"
	"time"
)

func main() {
	usage := `Hellfire: PATHspider Effects List Resolver

Hellfire is a parallelised DNS resolver. It builds effects lists for input to
PATHspider measurements. For sources where the filename is optional, the latest
source will be downloaded from the Internet when the filename is omitted.

Usage:
  hellfire --topsites [--file=<filename>]
  hellfire --cisco [--file=<filename>]
  hellfire --citizenlab (--country=<cc>|--file=<filename>)
  hellfire --opendns (--list=<name>|--file=<filename>)
  hellfire --csv --file=<filename>
  hellfire --txt --file=<filename>

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
		testList.(*inputs.CitizenLabCountryList).SetCountry(arguments["--country"].(string))
	} else if arguments["--opendns"].(bool) {
		testList = new(inputs.OpenDNSList)
		testList.(*inputs.OpenDNSList).SetListName(arguments["--list"].(string))
	}

	if testList != nil {
		performLookups(testList)
	} else {
		panic("An error occured building the input provider")
	}
}

func lookupWorker(id int, lookupWaitGroup *sync.WaitGroup,
	jobs chan map[string]interface{},
	results chan map[string]interface{}) {
	lookupWaitGroup.Add(1)
	fmt.Println("Spawning worker ", id)
	go func(id int, lookupWaitGroup *sync.WaitGroup, jobs chan map[string]interface{}, results chan map[string]interface{}) {
		defer lookupWaitGroup.Done()
		for job := range jobs {
			if job["domain"] == nil {
				jobs <- make(map[string]interface{})
				fmt.Println("Worker", id, "cascaded shutdown signal")
				break
			}
			fmt.Println("Worker", id, "got a job: ", job)
			var attempt int
			var ips []net.IP
			for attempt = 0; attempt <= 3; attempt++ {
				ips, _ = net.LookupIP(job["domain"].(string))
				if len(ips) == 0 {
					time.Sleep(1)
				} else {
					break
				}
			}
			job["ips"] = ips
			job["attempts"] = attempt
			results <- job
		}
		fmt.Println("Worker", id, "exiting")
	}(id, lookupWaitGroup, jobs, results)
}

func outputPrinter(outputWaitGroup *sync.WaitGroup, results chan map[string]interface{}) {
	outputWaitGroup.Add(1)
	fmt.Println("Spawning output printer")
	go func(results chan map[string]interface{}) {
		defer outputWaitGroup.Done()
		for {
			result := <-results
			if result["domain"] == nil {
				fmt.Println("Output printer got shutdown signal")
				break
			}
			fmt.Println("Result: ", result["domain"], result["ips"], result["attempts"])
		}
	}(results)
}

func performLookups(testList inputs.TestList) {
	var lookupWaitGroup sync.WaitGroup
	var outputWaitGroup sync.WaitGroup
	jobs := make(chan map[string]interface{}, 1)
	results := make(chan map[string]interface{})

	// Spawn lookup workers
	for i := 0; i < 300; i++ {
		lookupWorker(i, &lookupWaitGroup, jobs, results)
	}

	// Spawn output printer
	outputPrinter(&outputWaitGroup, results)

	// Submit jobs
	testList.FeedJobs(jobs)
	jobs <- make(map[string]interface{})
	lookupWaitGroup.Wait()
	<-jobs // Read last shutdown sentinel from the queue left by the
	// final worker to exit
	// https://blog.golang.org/pipelines - This is a better way
	fmt.Println("Workers now all finished")
	close(jobs)

	// Shutdown the output printer
	results <- make(map[string]interface{})
	outputWaitGroup.Wait()
	close(results)
}
