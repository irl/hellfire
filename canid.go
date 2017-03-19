package hellfire // import "pathspider.net/hellfire"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

func GetAdditionalInfo(ip net.IP, canidAddress string) map[string]interface{} {
	url := fmt.Sprintf("http://%s/prefix.json?addr=%s", canidAddress, ip.String())

	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	var info map[string]interface{}
	json.Unmarshal([]byte(body), &info)

	return info
}

