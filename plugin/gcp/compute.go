// Google Cloud Compute Instances

package gcp

import (
	"context"
	"log"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Compute struct {
	client *compute.InstancesClient
	GCP
}

func NewCompute(scopes []string) *Compute {
	return &Compute{
		GCP: NewGCP(scopes),
	}
}

func (c *Compute) InitializeClient(ctx context.Context) error {

	c.GCP.GetCredentials(ctx)

	// log.Println(string(c.GCP.credentials.JSON))
	// log.Println(c.GCP.ProjectID)

	instancesClient, err := compute.NewInstancesRESTClient(
		ctx,
		option.WithCredentials(c.GCP.credentials),
	)
	if err != nil {
		return err
	}

	// log.Println(instancesClient)

	c.client = instancesClient

	return nil
}

func (c *Compute) CloseClient() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Compute) ListAllInstances() error {

	req := &computepb.AggregatedListInstancesRequest{
		Project: c.ProjectID,
	}

	it := c.client.AggregatedList(context.Background(), req)

	log.Println("instances found: ")

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		instances := pair.Value.Instances

		if len(instances) > 0 {
			// log.Printf("%s\n", pair.Key)
			for _, instance := range instances {
				log.Printf("%s", instance.GetName())
			}
		}
	}
	return nil
}

func (c *Compute) GetAllInstances() ([]*computepb.Instance, error) {

	var allInstances []*computepb.Instance

	req := &computepb.AggregatedListInstancesRequest{
		Project: c.ProjectID,
	}

	it := c.client.AggregatedList(context.Background(), req)

	log.Println("instances found: ")

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		instances := pair.Value.Instances

		// allInstances = append(allInstances, *instances)

		if len(instances) > 0 {
			allInstances = append(allInstances, instances...)

			// log.Printf("%s\n", pair.Key) = append(allInstances, instances...)
			// for _, instance := range instances {
			// 	log.Printf("%s", instance.GetName())
			// 	allInstances = append(allInstances, *instance)
			// }
		}
	}
	return allInstances, nil
}
