package compute_instance

import (
	"context"
	"log"
	"strconv"

	"github.com/kaytu-io/plugin-gcp/plugin/preferences"
)

type ListComputeInstancesJob struct {
	processor *ComputeInstanceProcessor
}

func NewListComputeInstancesJob(processor *ComputeInstanceProcessor) *ListComputeInstancesJob {
	return &ListComputeInstancesJob{
		processor: processor,
	}
}

func (job *ListComputeInstancesJob) Id() string {
	return "list_compute_instances"
}

func (job *ListComputeInstancesJob) Description() string {
	return "List all compute instances in current project"

}

func (job *ListComputeInstancesJob) Run() error {

	log.Println("Running list compute instance job")

	err := job.processor.provider.InitializeClient(context.Background())
	if err != nil {
		return err
	}

	instances, err := job.processor.provider.GetAllInstances()
	if err != nil {
		return err
	}

	log.Printf("# of instances: %d", len(instances))

	for _, instance := range instances {
		oi := ComputeInstanceItem{
			Name:                *instance.Name,
			Id:                  strconv.FormatUint(instance.GetId(), 10),
			MachineType:         instance.GetMachineType(),
			Region:              instance.GetZone(),
			OptimizationLoading: false,
			Preferences:         preferences.DefaultComputeEnginePreferences,
			Skipped:             false,
			LazyLoadingEnabled:  false,
			SkipReason:          "NA",
			// Instance:            *instance,
			// Region:              j.region,
			// OptimizationLoading: true,
			// LazyLoadingEnabled:  false,
			// Preferences:         preferences2.DefaultEC2Preferences,
		}

		log.Printf("OI instance: %s", oi.Name)

		job.processor.items.Set(oi.Id, oi)
		job.processor.publishOptimizationItem(oi.ToOptimizationItem())
	}

	if err = job.processor.provider.CloseClient(); err != nil {
		return err
	}

	return nil

}
