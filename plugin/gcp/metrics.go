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

func (c *CloudMonitoring) NewInstanceMetricRequest(
	metricName string, // fully qualified name of the metric
	instanceID string, // compute instance ID
	startTime time.Time, // start time of requested time series
	endTime time.Time, // end time of requested time series
	periodInSeconds int64, // period, for which the datapoints will be aggregated into one, in seconds
) *monitoringpb.ListTimeSeriesRequest {

	request := &monitoringpb.ListTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", c.ProjectID),
		Filter: fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			metricName,
			instanceID,
		),
		Interval: &monitoringpb.TimeInterval{
			EndTime:   timestamppb.New(endTime),
			StartTime: timestamppb.New(startTime),
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod: &durationpb.Duration{
				Seconds: periodInSeconds,
			},
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_MEAN, // will represent all the datapoints in the above period, with a mean
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
