package main

import (
	"flag"
	"fmt"
	"inputs"
	"net"
	"os"
	"sync"
)

func main() {
	// FIXME: Use docopt https://github.com/docopt/docopt.go
	topsites := flag.Bool("topsites", false, "Use the Alexa Topsites list")
	citizenlab := flag.String("citizenlab", "none", "Use the Citizen Lab test list for specified country")
	flag.Parse()

	var testList inputs.TestList

	if *topsites && (*citizenlab == "none") {
		panic("You can only select one source")
	} else if !*topsites && (*citizenlab == "none") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *topsites {
		testList = new(inputs.AlexaTopsitesList)
	} else {
		l := new(inputs.CitizenLabCountryList)
		l.SetCountry(*citizenlab)
		testList = l
	}
	performLookups(testList)
}

func lookupWorker(id int, waitGroup *sync.WaitGroup,
	jobs chan map[string]interface{},
	results chan map[string]interface{}) {
	waitGroup.Add(1)
	fmt.Println("Spawning worker ", id)
	go func(id int, jobs chan map[string]interface{}, results chan map[string]interface{}) {
		defer waitGroup.Done()
		for job := range jobs {
			if job["domain"] == "" {
				fmt.Println("Worker", id, "cascading shutdown signal")
				jobs <- make(map[string]interface{})
				break
			}
			fmt.Println("Worker", id, "got a job: ", job)
			ips, _ := net.LookupIP(job["domain"].(string))
			job["ips"] = ips
			results <- job
		}
	}(id, jobs, results)
}

func outputPrinter(waitGroup *sync.WaitGroup, results chan map[string]interface{}) {
	defer waitGroup.Done()
	for {
		result := <-results
		if result["domain"] == nil {
			break
		}
		fmt.Println("Result: ", result["domain"], result["ips"])
	}
}

func performLookups(testList inputs.TestList) {
	var waitGroup sync.WaitGroup
	var jobs chan map[string]interface{} = make(chan map[string]interface{})
	var results chan map[string]interface{} = make(chan map[string]interface{})

	// Spawn lookup workers
	for i := 0; i < 300; i++ {
		lookupWorker(i, &waitGroup, jobs, results)
	}

	// Spawn output printer
	// The output printer doesn't lock the wait group until after the
	// workers are done - maybe we should have two
	go outputPrinter(&waitGroup, results)

	// Submit jobs
	testList.FeedJobs(jobs)
	waitGroup.Wait()
	close(jobs)

	// Shutdown the output printer
	waitGroup.Add(1)
	shutdownSentinel := make(map[string]interface{})
	results <- shutdownSentinel
	close(results)
	waitGroup.Wait()
}
