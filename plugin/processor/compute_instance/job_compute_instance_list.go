package compute_instance

import (
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

	instances, err := job.processor.provider.GetAllInstances()
	if err != nil {
		return err
	}

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

		job.processor.items.Set(oi.Id, oi)
		job.processor.publishOptimizationItem(oi.ToOptimizationItem())
	}

	return job.processor.provider.ListAllInstances()

}
