package gcp

import (
	"context"
	"fmt"
	golang2 "github.com/kaytu-io/plugin-gcp/plugin/proto/src/golang"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/wrapperspb"

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

func (c *CloudMonitoring) GetMetric(request *monitoringpb.ListTimeSeriesRequest) ([]*golang2.DataPoint, error) {
	var dps []*golang2.DataPoint

	it := c.client.ListTimeSeries(context.Background(), request)
	for {
		resp, err := it.Next()

		if err != nil {
			if err == iterator.Done {
				break
			} else {
				return nil, err
			}
		}
		dps = append(dps, convertDatapoints(resp)...)
	}

	return dps, nil

}

func convertDatapoints(resp *monitoringpb.TimeSeries) []*golang2.DataPoint {
	var dps []*golang2.DataPoint
	for _, dp := range resp.GetPoints() {
		dps = append(dps, &golang2.DataPoint{
			Value:     dp.GetValue().GetDoubleValue(),
			StartTime: wrapperspb.Int64(dp.GetInterval().GetStartTime().GetSeconds()),
			EndTime:   wrapperspb.Int64(dp.GetInterval().GetEndTime().GetSeconds()),
		})
	}
	return dps
}
