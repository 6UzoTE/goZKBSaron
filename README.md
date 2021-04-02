# goZKBSaron

Scrape current SARON Interest rate (SWISS AVERAGE RATE OVERNIGHT ISIN: CH0049613687)
Value scraped from mdgms.com/ zkb.ch


Build:
  go build goZKBsaron.go

Usage example:
    Run in CRON without parameters set. Errors will be reported via eMail if set in CRON

Parameters:
Usage of goZKBsaron:
  -flushlog string
    	sets the flush trigger level (default "none")
  -influxdb
    	Persist to influxdb (true/false) default=true (default true)
  -log string
    	sets the logging threshold (default "info")
  -stderr
    	outputs to standard error (stderr)

Example:
  goZKBSaron --log=Debug --influxdb=false
