package gcp

import (
	"context"
	"log"
	"os"
	"testing"
)

// run this test as
//
//	`TEST_INSTANCE_ID="<an-instance-id>" make testgcp`
func TestGetMetrics(t *testing.T) {

	id := os.Getenv("TEST_INSTANCE_ID")

	log.Printf("running %s", t.Name())
	metric := NewCloudMonitoring(
		[]string{
			"https://www.googleapis.com/auth/monitoring.read",
		},
	)
	err := metric.InitializeClient(context.Background())
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}

	request := metric.NewMetricRequest(
		"compute.googleapis.com/instance/cpu/utilization",
		id,
	)

	resp := metric.GetMetric(request)

	log.Printf("metrics: %s", resp.GetMetric().String())
	log.Printf("resource: %s", resp.GetResource().String())

	for _, point := range resp.Points {
		log.Printf("Point : %s", point.String())
	}

	metric.CloseClient()
}
