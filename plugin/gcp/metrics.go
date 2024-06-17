package gcp

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
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

func (c *CloudMonitoring) NewTimeSeriesRequest(
	filter string, // filter for time series metric, containing metric name and resource label
	interval *monitoringpb.TimeInterval, // interval containing start and end time of the requested time series
	aggregation *monitoringpb.Aggregation, // operations to perform on time series data before returning
) *monitoringpb.ListTimeSeriesRequest {

	return &monitoringpb.ListTimeSeriesRequest{
		Name:        fmt.Sprintf("projects/%s", c.ProjectID),
		Filter:      filter,
		Interval:    interval,
		Aggregation: aggregation,
		View:        monitoringpb.ListTimeSeriesRequest_FULL,
	}

}

func (c *CloudMonitoring) GetMetric(request *monitoringpb.ListTimeSeriesRequest) (*monitoringpb.TimeSeries, error) {

	it := c.client.ListTimeSeries(context.Background(), request)

	resp, err := it.Next()
	if err != nil {
		return nil, err
	}

	return resp, err

}
