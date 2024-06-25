package gcp

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// run this test as
//
//	`TEST_INSTANCE_ID="<an-instance-id>" make testgcp`
func TestGetMetrics(t *testing.T) {

	//test variables
	id := os.Getenv("TEST_INSTANCE_ID")
	endTime := time.Now()
	startTime := endTime.Add(-24 * 1 * time.Hour) // 24 hours before current time

	t.Logf("running %s", t.Name())

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
	memoryRequest := metric.NewTimeSeriesRequest(
		fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			"compute.googleapis.com/instance/memory/balloon/ram_used",
			id,
		),
		&monitoringpb.TimeInterval{
			EndTime:   timestamppb.New(endTime),
			StartTime: timestamppb.New(startTime),
		},
		&monitoringpb.Aggregation{
			AlignmentPeriod: &durationpb.Duration{
				Seconds: 60,
			},
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_NONE, // will represent all the datapoints in the above period, with a mean
		},
	)

	// execute the request
	resp, err := metric.GetMetric(memoryRequest)
	if err != nil {
		t.Error(err)
	}

	// log.Printf("metrics: %s", resp.GetMetric().String())
	// log.Printf("resource: %s", resp.GetResource().String())
	t.Logf("datapoints: %d", len(resp))
	// log.Println(resp.Unit)
	// log.Printf("Point 1 : %.0f %s", resp.GetPoints()[0].GetValue().GetDoubleValue(), resp.GetUnit())

	// for _, point := range resp.Points {
	// 	log.Printf("Point : %.0f", point.GetValue().GetDoubleValue())
	// }

	metric.CloseClient()
}

func TestGetDiskUsage(t *testing.T) {

	//test variables
	id := os.Getenv("TEST_INSTANCE_ID")
	// diskId := os.Getenv("DISK_ID")
	endTime := time.Now()
	startTime := endTime.Add(-24 * 1 * time.Hour) // 24 hours before current time

	t.Logf("running %s", t.Name())

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
	diskRequest := metric.NewTimeSeriesRequest(
		fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			"agent.googleapis.com/disk/percent_used",
			id,
			// diskId,
		),
		&monitoringpb.TimeInterval{
			EndTime:   timestamppb.New(endTime),
			StartTime: timestamppb.New(startTime),
		},
		&monitoringpb.Aggregation{
			AlignmentPeriod: &durationpb.Duration{
				Seconds: 60,
			},
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_NONE, // will represent all the datapoints in the above period, with a mean
		},
	)

	// execute the request
	resp, err := metric.GetMetric(diskRequest)
	if err != nil {
		t.Error(err)
	}

	// log.Printf("metrics: %s", resp.GetMetric().String())
	// log.Printf("resource: %s", resp.GetResource().String())
	t.Logf("datapoints: %d", len(resp))
	// log.Println(resp.Unit)
	// log.Printf("Point 1 : %.0f %s", resp.GetPoints()[0].GetValue().GetDoubleValue(), resp.GetUnit())

	// for _, point := range resp.Points {
	// 	log.Printf("Point : %.0f", point.GetValue().GetDoubleValue())
	// }

	metric.CloseClient()
}
