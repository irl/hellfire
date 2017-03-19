// Hellfire is a parallelised DNS resolver. It builds effects lists for input
// to PATHspider measurements. For sources where the filename is optional, the
// latest source will be downloaded from the Internet when the filename is
// omitted.
//
// BASIC USAGE
//
//  Usage:
//    hellfire --topsites [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
//    hellfire --cisco [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
//    hellfire --citizenlab [--country=<cc>|--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
//    hellfire --opendns [--list=<name>|--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
//    hellfire --csv --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
//    hellfire --txt --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
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
	"github.com/docopt/docopt-go"
	"pathspider.net/hellfire"
	"strings"
)

func main() {
	usage := `Hellfire: PATHspider Effects List Resolver

Hellfire is a parallelised DNS resolver. It builds effects lists for input to
PATHspider measurements. For sources where the filename is optional, the latest
source will be downloaded from the Internet when the filename is omitted.

Usage:
  hellfire --topsites [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
  hellfire --cisco [--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
  hellfire --citizenlab [--country=<cc>|--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
  hellfire --opendns [--list=<name>|--file=<filename>] [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
  hellfire --csv --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]
  hellfire --txt --file=<filename> [--output=<individual|array|oneeach>] [--type=<host|ns|mx>] [--canid=<canid address>]

Options:
  -h --help                           Show this screen.
  --version                           Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, "Hellfire dev", false)

	var listName string
	var listVariant string
	var listFilename string

	//BUG(irl): CSV type is ignored
	//BUG(irl): TXT type is ignored
	//BUG(irl): Can't select SRV lookup yet

	if arguments["--topsites"].(bool) {
		listName = "topsites"
	} else if arguments["--cisco"].(bool) {
		listName = "cisco"
	} else if arguments["--citizenlab"].(bool) {
		listName = "citizenlab"
		if arguments["--country"] != nil {
			listVariant = arguments["--country"].(string)
		} else {
			listVariant = "global"
		}
	} else if arguments["--opendns"].(bool) {
		listName = "opendns"
		if arguments["--list"] != nil {
			listVariant = arguments["--list"].(string)
		} else {
			listVariant = "top"
		}
	}

	if arguments["--file"] != nil {
		listFilename = arguments["--file"].(string)
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

	var canidAddress string
	if arguments["--canid"] == nil {
		canidAddress = ""
	} else {
		canidAddress = arguments["--canid"].(string)
	}

	testListOptions := strings.Join([]string{listName, listVariant, listFilename}, ";")
	hellfire.PerformLookups(testListOptions, lookupType, outputType, canidAddress)
}
