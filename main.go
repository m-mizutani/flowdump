package main

import log "github.com/sirupsen/logrus"

func main() {
	opts := newFlowDumpOpts()
	opts.deviceName = "en0"
	err := loop(opts)
	if err != nil {
		log.WithField("error", err).Error("Error in main loop")
	}
}
