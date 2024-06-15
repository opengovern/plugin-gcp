package gcp

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

// run this test as
//
//	`TEST_INSTANCE_ID="<an-instance-id>" make testgcp`
func TestGetMetrics(t *testing.T) {

	//test variables
	id := os.Getenv("TEST_INSTANCE_ID")
	endtime := time.Now()
	starttime := endtime.Add(-24 * 1 * time.Hour) // 24 hours before current time

	log.Printf("running %s", t.Name())

	// creating and initializing client
	metric := NewCloudMonitoring(
		[]string{
			"https://www.googleapis.com/auth/monitoring.read",
		},
	)
	err := metric.InitializeClient(context.Background())
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}

	// creating the metric request for the instance
	request := metric.NewInstanceMetricRequest(
		"compute.googleapis.com/instance/cpu/utilization",
		id,
		starttime,
		endtime,
		60,
	)

	// execute the request
	resp, err := metric.GetMetric(request)
	if err != nil {
		t.Error(err)
	}

	log.Printf("metrics: %s", resp.GetMetric().String())
	log.Printf("resource: %s", resp.GetResource().String())
	log.Printf("# of points: %d", len(resp.Points))

	// for _, point := range resp.Points {
	// 	log.Printf("Point : %.10f", point.GetValue().GetDoubleValue())
	// }

	metric.CloseClient()
}
