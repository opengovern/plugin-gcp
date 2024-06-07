package gcp

import (
	"context"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
)

type CloudMonitoring struct {
	client *monitoring.MetricClient
	GCP
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
