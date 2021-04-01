package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/stdlog"
)

var logger log.Logger
var address = "https://zkb-finance.mdgms.com/home/indices/detail.html?FI_ID_NOTATION=30535364"

func main() {
	// Check command line param
	logger = stdlog.GetFromFlags()
	var re = regexp.MustCompile(`(?m)\<p\sclass\=\"fi-rate\"\>Aktuell\s*\<span\sclass\=\"number\"\>\s*CHF\s*(.*[-.1234567890])`)
	logger.Debugf("Started goZKBSaron.go")
	myClient := &http.Client{
		Timeout: time.Second * 20,
	}
	resp, err := myClient.Get(address)
	if err != nil {
		logger.Debugf("Error Client get on addresse: %s Error: %s", address, err)
	} else {

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Criticalf("Error :", err)
		}
		for i, match := range re.FindAllSubmatch(body, -1) {
			logger.Debugf("%s", match[i+1])
		}

		//ipStr = string(body[:])
	}
}
