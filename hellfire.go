package hellfire // import "pathspider.net/hellfire"

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type LookupQueryResult struct {
	attempts int
	result   []net.IP
}

func prepareTestList(testListOptions string) TestList {
	var testList TestList

	options := strings.Split(testListOptions, ";")

	if len(options) < 3 {
		return nil
	}

	if options[0] == "topsites" {
		testList = new(AlexaTopsitesList)
	} else if options[0] == "cisco" {
		testList = new(CiscoUmbrellaList)
	} else if options[0] == "citizenlab" {
		testList = new(CitizenLabCountryList)
		if options[1] != "" {
			testList.(*CitizenLabCountryList).SetCountry(options[1])
		}
	} else if options[0] == "opendns" {
		testList = new(OpenDNSList)
		if options[1] != "" {
			testList.(*OpenDNSList).SetListName(options[1])
		}
	} else {
		return nil
	}

	if options[2] != "" {
		testList.SetFilename(options[2])
	}

	return testList
}

func makeQuery(domain string, lookupType string) LookupQueryResult {
	result := []net.IP{}
	domains := []string{}
	lookupAttempt := 1

	//BUG(irl): Need to add support for MX lookups
	//BUG(irl): Need to add support for SRV lookups
	if lookupType == "host" {
		domains = append(domains, domain)
	} else if lookupType == "ns" {
		var nss []*net.NS
		for {
			nss, _ = net.LookupNS(domain)
			if len(nss) == 0 {
				time.Sleep(1)
			} else {
				break
			}
			lookupAttempt++
			if lookupAttempt == 4 {
				lookupAttempt = 3
				break
			}
		}
		for _, ns := range nss {
			domains = append(domains, ns.Host)
		}
	} else if lookupType == "mx" {
		var nss []*net.MX
		for {
			nss, _ = net.LookupMX(domain)
			if len(nss) == 0 {
				time.Sleep(1)
			} else {
				break
			}
			lookupAttempt++
			if lookupAttempt == 4 {
				lookupAttempt = 3
				break
			}
		}
		for _, ns := range nss {
			domains = append(domains, ns.Host)
		}
	}

	for _, d := range domains {
		var ips []net.IP
		for {
			ips, _ = net.LookupIP(d)
			if len(ips) == 0 {
				time.Sleep(1)
			} else {
				break
			}
			lookupAttempt++
			if lookupAttempt == 4 {
				lookupAttempt = 3
				break
			}
		}
		result = append(result, ips...)
	}
	return LookupQueryResult{lookupAttempt, result}
}

func lookupWorker(id int, lookupWaitGroup *sync.WaitGroup,
	jobs chan map[string]interface{},
	results chan map[string]interface{},
	lookupType string,
	canidAddress string) {
	lookupWaitGroup.Add(1)
	go func(id int,
		lookupWaitGroup *sync.WaitGroup,
		jobs chan map[string]interface{},
		results chan map[string]interface{},
		lookupType string,
		canidAddress string) {
		defer lookupWaitGroup.Done()
		for job := range jobs {
			if job["domain"] == nil {
				jobs <- make(map[string]interface{})
				break
			}
			lookupResult := makeQuery(job["domain"].(string),
				lookupType)
			job["hellfire_lookup_attempts"] = lookupResult.attempts
			job["hellfire_lookup_type"] = lookupType
			for _, ip := range lookupResult.result {
				thisResult := make(map[string]interface{})
				for key, value := range job {
					thisResult[key] = value
				}
				thisResult["ips"] = []net.IP{ip}
				if canidAddress != "" {
					thisResult["canid_info"] = GetAdditionalInfo(ip, canidAddress)
				}
				results <- thisResult
			}
		}
	}(id, lookupWaitGroup, jobs, results, lookupType, canidAddress)
}

func outputPrinter(outputWaitGroup *sync.WaitGroup, results chan map[string]interface{}, outputType string) {
	outputWaitGroup.Add(1)
	go func(results chan map[string]interface{}) {
		defer outputWaitGroup.Done()
		for {
			result := <-results
			if result["domain"] == nil {
				break
			}
			if outputType == "all" {
				b, _ := json.Marshal(result)
				fmt.Println(string(b))
			} else if outputType == "individual" {
				ips := result["ips"]
				if ips == nil {
					continue
				}
				delete(result, "ips")
				for _, ipo := range ips.([]net.IP) {
					ip := ipo.String()
					result["dip"] = ip
					b, _ := json.Marshal(result)
					fmt.Println(string(b))
					delete(result, "dip")
				}
			} else if outputType == "oneeach" {
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
					result["dip"] = ip
					b, _ := json.Marshal(result)
					fmt.Println(string(b))
					delete(result, "dip")
				}
			}
		}
	}(results)
}

func PerformLookups(testListOptions string, lookupType string, outputType string, canidAddress string) {
	var lookupWaitGroup sync.WaitGroup
	var outputWaitGroup sync.WaitGroup

	jobs := make(chan map[string]interface{}, 1)
	results := make(chan map[string]interface{})
        testList := prepareTestList(testListOptions)

	// Spawn lookup workers
	for i := 0; i < 300; i++ {
		lookupWorker(i, &lookupWaitGroup, jobs, results, lookupType, canidAddress)
	}

	// Spawn output printer
	outputPrinter(&outputWaitGroup, results, outputType)

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
