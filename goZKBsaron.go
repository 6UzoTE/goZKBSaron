package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/stdlog"
	client "github.com/influxdata/influxdb1-client/v2"
)

var logger log.Logger

//INFLUXconn is the connection Struct
type INFLUXconn struct {
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
	TLS  bool   `json:"tls"`
}

//READ JSON INFLUX DB configuration
func readConfJSON() INFLUXconn {
	var influxc INFLUXconn
	var bytefile []byte
	var err error
	bytefile, err = ioutil.ReadFile("influx_config.json")
	logger.Debugf("Try to read private config influx_config.json")
	if err != nil {
		logger.Debugf(err.Error())
		logger.Debugf("Error: Reding PRIVATE INFLUX Configuration influx_config.json")
		bytefile, err = ioutil.ReadFile("influx.json")
		if err != nil {
			logger.Criticalf(err.Error())
			logger.Criticalf("Error: Reading PUBLIC INFLUX Configuration influx.json ")
			os.Exit(1)
		}
	}
	json.Unmarshal(bytefile[:], &influxc)
	logger.Debugf("INFLUX Configuration read")
	return influxc
}

//Write to INFLUX DB
func writeInflux(saron float64) {
	var hoststring string
	inc := readConfJSON()
	logger.Debugf("Host: %s", inc.Host)
	logger.Debugf("User: %s", inc.User)
	//logger.Debugf("Pass: %s", inc.Pass)
	logger.Debugf("TLS: %t", inc.TLS)

	if inc.TLS {
		hoststring = strings.Join([]string{"https", inc.Host}, "://")
	} else {
		hoststring = strings.Join([]string{"http", inc.Host}, "://")
	}
	logger.Debugf("HOSTSRTRING: %s", hoststring)

	// Write influxdb
	// Create a new influx HTTPClient -
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:      hoststring,
		Username:  inc.User,
		Password:  inc.Pass,
		UserAgent: "goZKBSaron.go"})

	if err != nil {
		logger.Criticalf("Error: %s", err)
	}
	defer c.Close()
	bp, errb := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "saron",
		Precision: "m"})

	if errb != nil {
		logger.Criticalf("Error: %s", errb)
	}
	tags := map[string]string{"ZKB": "Saron"}
	field := map[string]interface{}{"interest": saron}
	mtime := time.Now()
	pt, errN := client.NewPoint("sarondaily", tags, field, mtime)

	if errN != nil {
		logger.Criticalf("Error creating newpoint : %s", errN)
	}
	bp.AddPoint(pt)
	//Cleanup
	if err := c.Write(bp); err != nil {
		logger.Criticalf("Error creating newpoint : %s", err)
	}

}

//READ (scrape) current interest rate from webpage
func readZkb() float64 {
	var conver error
	var re = regexp.MustCompile(`(?m)\<p\sclass\=\"fi-rate\"\>Aktuell\s*\<span\sclass\=\"number\"\>\s*CHF\s*(.*[-.1234567890])`)
	var address = "https://zkb-finance.mdgms.com/home/indices/detail.html?FI_ID_NOTATION=30535364"
	var floatSaron float64
	myClient := &http.Client{
		Timeout: time.Second * 20,
	}
	resp, err := myClient.Get(address)
	if err != nil {
		logger.Criticalf("Error Client get on addresse: %s Error: %s", address, err)
		floatSaron = 101
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		//removed for backward compatiblity Go 1.11 (Debian)
		//myClient.CloseIdleConnections()
		if err != nil {
			logger.Criticalf("Error :", err)
		}
		for i, match := range re.FindAllSubmatch(body, -1) {
			//logger.Debugf("%s %d", match[i+1], i)
			if i <= 0 {
				floatSaron, conver = strconv.ParseFloat(string(match[i+1]), 64)
			} else {
				floatSaron = 101
			}
			if conver != nil {
				logger.Debugf("Error conversion: %s", conver)
				floatSaron = 101
			}
		}
		if floatSaron == 101 {
			logger.Debugf("Error: Float %f EXITING", floatSaron)
		}
	}
	return floatSaron
}

func main() {
	// Check command line param
	var persistInflux bool
	flag.BoolVar(&persistInflux, "influxdb", true, "Persist to influxdb (true/false) default=true")
	flag.Parse()
	logger = stdlog.GetFromFlags()
	logger.Debugf("Started goZKBSaron.go")
	logger.Debugf("Retrieve Value for SARON")
	saron := readZkb()
	if saron > 100 {
		logger.Criticalf("Error: Saron value to high: %f")
		os.Exit(101)
	} else {
		logger.Debugf("Retrieved SARON: %f", saron)
	}
	if persistInflux {
		logger.Debugf("Persisting to InfluxDB --influxdb=%t", persistInflux)
		writeInflux(saron)
	} else {
		logger.Debugf("Not persisting to InfluxDB --influxdb=%t", persistInflux)
	}
}
