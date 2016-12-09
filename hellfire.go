package main

import (
	"flag"
	"fmt"
	"pathspider.net/hellfire/inputs"
	"net"
	"os"
	"time"
	"sync"
)

func main() {
	// FIXME: Use docopt https://github.com/docopt/docopt.go
	topsites := flag.Bool("topsites", false, "Use the Alexa Topsites list")
	citizenlab := flag.String("citizenlab", "none", "Use the Citizen Lab test list for specified country")
	flag.Parse()

	var testList inputs.TestList

	if *topsites && (*citizenlab != "none") {
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
			for attempt = 0 ; attempt <= 3 ; attempt++ {
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
	var jobs chan map[string]interface{} = make(chan map[string]interface{}, 1)
	var results chan map[string]interface{} = make(chan map[string]interface{})

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
	<- jobs // Read last shutdown sentinel from the queue left by the
		// final worker to exit
		// https://blog.golang.org/pipelines - This is a better way
	fmt.Println("Workers now all finished")
	close(jobs)

	// Shutdown the output printer
	results <- make(map[string]interface{})
	outputWaitGroup.Wait()
	close(results)
}
