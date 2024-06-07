package gcp

import (
	"context"
	"log"
	"testing"
)

func TestListAllInstances(t *testing.T) {
	log.Printf("running %s", t.Name())
	compute := NewCompute(
		[]string{
			"https://www.googleapis.com/auth/compute.readonly",
		},
	)
	err := compute.InitializeClient(context.Background())
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}

	log.Printf("[%s]: %s", t.Name(), compute.ProjectID)

	err = compute.ListAllInstances()
	if err != nil {
		compute.CloseClient()
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}
	compute.CloseClient()

}

func TestGetAllInstances(t *testing.T) {

	log.Printf("running %s", t.Name())
	compute := NewCompute(
		[]string{
			"https://www.googleapis.com/auth/compute.readonly",
		},
	)
	err := compute.InitializeClient(context.Background())
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}

	log.Printf("[%s]: %s", t.Name(), compute.ProjectID)

	instances, err := compute.GetAllInstances()
	if err != nil {
		compute.CloseClient()
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}

	for _, instance := range instances {
		log.Println(instance.GetMachineType())
	}

	// log.Println(instances)

	compute.CloseClient()

}
