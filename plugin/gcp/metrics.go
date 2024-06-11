package gcp

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CloudMonitoring struct {
	client *monitoring.MetricClient
	GCP
}

func NewCloudMonitoring(scopes []string) *CloudMonitoring {
	return &CloudMonitoring{
		GCP: NewGCP(scopes),
	}
}

func (c *CloudMonitoring) InitializeClient(ctx context.Context) error {

	c.GCP.GetCredentials(ctx)

	metricClient, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}

	c.client = metricClient

	return nil
}

func (c *CloudMonitoring) CloseClient() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *CloudMonitoring) NewMetricRequest(metricName string, instanceID string) *monitoringpb.ListTimeSeriesRequest {

	endtime := time.Now()
	starttime := endtime.Add(-24 * 1 * time.Hour) // 24 hours before current time

	request := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", c.ProjectID),
		Filter: fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			metricName,
			instanceID,
		),
		Interval: &monitoringpb.TimeInterval{
			EndTime:   timestamppb.New(endtime),
			StartTime: timestamppb.New(starttime),
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod: durationpb.New(time.Minute),
		},
		View: monitoringpb.ListTimeSeriesRequest_FULL,
	}

	return request
}

func (c *CloudMonitoring) GetMetric(request *monitoringpb.ListTimeSeriesRequest) *monitoringpb.TimeSeries {

	it := c.client.ListTimeSeries(context.Background(), request)

	resp, err := it.Next()
	if err != nil {
		panic(err)
	}

	return resp

}
