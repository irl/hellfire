// Hellfire is a parallelised DNS resolver. It builds effects lists for input
// to PATHspider measurements. For sources where the filename is optional, the
// latest source will be downloaded from the Internet when the filename is
// omitted.
//
// BASIC USAGE
//
//  Usage:
//    hellfire --topsites [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
//    hellfire --cisco [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
//    hellfire --citizenlab (--country=<cc>|--file=<filename>) [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
//    hellfire --opendns (--list=<name>|--file=<filename>) [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
//    hellfire --csv --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
//    hellfire --txt --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
//
//  Options:
//    -h --help     Show this screen.
//    --version     Show version.
//
// OUTPUT TYPES
//
// * "individual" - One record output per IP address looked up, discarding no
// addresses.
// * "array" - One record output per domain name, with an array of all
// addresses resolved.
// * "oneeach" - One record output per IP address, only printing one IPv4 and
// one IPv6 at most for each domain.
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

type LookupQueryResult struct {
	attempts int
	result   interface{}
}

func main() {
	usage := `Hellfire: PATHspider Effects List Resolver

Hellfire is a parallelised DNS resolver. It builds effects lists for input to
PATHspider measurements. For sources where the filename is optional, the latest
source will be downloaded from the Internet when the filename is omitted.

Usage:
  hellfire --topsites [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
  hellfire --cisco [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
  hellfire --citizenlab (--country=<cc>|--file=<filename>) [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
  hellfire --opendns (--list=<name>|--file=<filename>) [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
  hellfire --csv --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]
  hellfire --txt --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>]

Options:
  -h --help                           Show this screen.
  --version                           Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, "Hellfire dev", false)

	var testList inputs.TestList
	//BUG(irl): Filenames are ignored
	//BUG(irl): CSV type is ignored
	//BUG(irl): TXT type is ignored
	//BUG(irl): Can't actually select NS, MX or SRV lookup yet
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

		var lookupType string
		supportedLookupTypes := []string{"host", "mx", "ns"}
		if arguments["--type"] != nil {
			for _, supportedType := range supportedLookupTypes {
				if arguments["--type"].(string) == supportedType {
					lookupType = arguments["--type"].(string)
				}
			}
			if lookupType == "" {
				panic("Unsupported lookup type requested.")
				//BUG(irl): Should list the supported types.
			}
		} else {
			lookupType = "host"
		}

		var outputType string
		supportedOutputTypes := []string{"individual", "array", "oneeach"}
		if arguments["--output"] != nil {
			for _, supportedType := range supportedOutputTypes {
				if arguments["--output"].(string) == supportedType {
					outputType = arguments["--output"].(string)
				}
			}
			if outputType == "" {
				panic("Unsupported lookup type requested.")
				//BUG(irl): Should list the supported types.
			}
		} else {
			outputType = "individual"
		}

		performLookups(testList, lookupType, outputType)
	} else {
		panic("An error occured building the input provider")
	}
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
	lookupType string) {
	lookupWaitGroup.Add(1)
	go func(id int, lookupWaitGroup *sync.WaitGroup, jobs chan map[string]interface{}, results chan map[string]interface{}, lookupType string) {
		defer lookupWaitGroup.Done()
		for job := range jobs {
			if job["domain"] == nil {
				jobs <- make(map[string]interface{})
				break
			}
			lookupResult := makeQuery(job["domain"].(string),
				lookupType)
			job["ips"] = lookupResult.result
			job["lookupAttempts"] = lookupResult.attempts
			job["lookupType"] = lookupType
			results <- job
		}
	}(id, lookupWaitGroup, jobs, results, lookupType)
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
				ips := result["ips"].([]net.IP)
				delete(result, "ips")
				for _, ipo := range ips {
					ip := ipo.String()
					result["ip"] = ip
					b, _ := json.Marshal(result)
					fmt.Println(string(b))
					delete(result, "ip")
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
					result["ip"] = ip
					b, _ := json.Marshal(result)
					fmt.Println(string(b))
					delete(result, "ip")
				}
			}
		}
	}(results)
}

func performLookups(testList inputs.TestList, lookupType string, outputType string) {
	var lookupWaitGroup sync.WaitGroup
	var outputWaitGroup sync.WaitGroup
	jobs := make(chan map[string]interface{}, 1)
	results := make(chan map[string]interface{})

	// Spawn lookup workers
	for i := 0; i < 300; i++ {
		lookupWorker(i, &lookupWaitGroup, jobs, results, lookupType)
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
