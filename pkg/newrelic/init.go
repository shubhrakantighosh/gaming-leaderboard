package newrelic

import (
	"log"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

var NRApp *newrelic.Application

// InitNewRelic initializes New Relic using provided app name & license key
func InitNewRelic(appName, licenseKey string) {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		log.Fatalf("failed to initialize New Relic: %v", err)
	}

	// Wait for agent to connect
	if err := app.WaitForConnection(5 * time.Second); err != nil {
		log.Fatalf("New Relic failed to connect: %v", err)
	}

	NRApp = app
	log.Println("New Relic initialized successfully")
}
