package compute_instance

import (
	"context"
	"google.golang.org/api/compute/v1"
	"log"
	"strconv"
	"strings"

	"github.com/kaytu-io/plugin-gcp/plugin/preferences"
	util "github.com/kaytu-io/plugin-gcp/utils"
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

func (job *ListComputeInstancesJob) Run(ctx context.Context) error {

	log.Println("Running list compute instance job")

	instances, err := job.processor.provider.GetAllInstances()
	if err != nil {
		return err
	}

	log.Printf("# of instances: %d", len(instances))

	for _, instance := range instances {
		var disks []compute.Disk
		for _, attachedDisk := range instance.Disks {
			diskURLParts := strings.Split(*attachedDisk.Source, "/")
			diskName := diskURLParts[len(diskURLParts)-1]

			zoneURLParts := strings.Split(*instance.Zone, "/")
			instanceZone := zoneURLParts[len(zoneURLParts)-1]

			diskDetails, err := job.processor.provider.GetDiskDetails(instanceZone, diskName)
			if err != nil {
				return err
			}
			disks = append(disks, *diskDetails)
		}

		oi := ComputeInstanceItem{
			ProjectId:           job.processor.provider.ProjectID,
			Name:                *instance.Name,
			Id:                  strconv.FormatUint(instance.GetId(), 10),
			MachineType:         util.TrimmedString(*instance.MachineType, "/"),
			Region:              util.TrimmedString(*instance.Zone, "/"),
			Platform:            instance.GetCpuPlatform(),
			OptimizationLoading: false,
			Preferences:         preferences.DefaultComputeEnginePreferences,
			Skipped:             false,
			LazyLoadingEnabled:  false,
			SkipReason:          "NA",
			Disks:               disks,
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
