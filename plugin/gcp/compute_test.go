package gcp

import (
	"context"
	"log"
	"os"
	"testing"
)

var (
	ctx = context.Background()
)

func TestListAllInstances(t *testing.T) {
	log.Printf("running %s", t.Name())
	compute := NewCompute(
		[]string{
			"https://www.googleapis.com/auth/compute.readonly",
		},
	)
	err := compute.InitializeClient(ctx)
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
		return
	}
	defer compute.CloseClient()
	log.Printf("[%s]: %s", t.Name(), compute.ProjectID)

	err = compute.ListAllInstances(ctx)
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}
}

func TestGetAllInstances(t *testing.T) {

	log.Printf("running %s", t.Name())
	compute := NewCompute(
		[]string{
			"https://www.googleapis.com/auth/compute.readonly",
		},
	)
	err := compute.InitializeClient(ctx)
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
		return
	}
	defer compute.CloseClient()

	log.Printf("[%s]: %s", t.Name(), compute.ProjectID)

	instances, err := compute.GetAllInstances(ctx)
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
		return
	}

	for _, instance := range instances {
		log.Println(instance.GetMachineType())
	}
}

// TEST_INSTANCE_ZONE="us-east1-b" TEST_INSTANCE_MACHINE_TYPE="e2-micro" TEST_INSTANCE_ID="7828543314219019363" make testgcp
func TestGetMemory(t *testing.T) {

	machineType := os.Getenv("TEST_INSTANCE_MACHINE_TYPE")
	zone := os.Getenv("TEST_INSTANCE_ZONE")

	log.Printf("running %s", t.Name())
	compute := NewCompute(
		[]string{
			"https://www.googleapis.com/auth/compute.readonly",
		},
	)
	err := compute.InitializeClient(ctx)
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
		return
	}
	defer compute.CloseClient()

	memory, err := compute.GetMemory(ctx, machineType, zone)
	if err != nil {
		t.Errorf("[%s]: %s", t.Name(), err.Error())
	}

	log.Printf("Memory : %d", &memory)
}
