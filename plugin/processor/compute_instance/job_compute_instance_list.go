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
			Metrics:             nil,
		}

		log.Printf("OI instance: %s", oi.Name)

		job.processor.items.Set(oi.Id, oi)
		job.processor.publishOptimizationItem(oi.ToOptimizationItem())
	}

	for _, instance := range instances {

		job.processor.jobQueue.Push(NewGetComputeInstanceMetricsJob(job.processor, instance))
	}

	return nil

}
