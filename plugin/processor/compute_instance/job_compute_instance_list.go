package compute_instance

import (
	"context"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"google.golang.org/api/compute/v1"
	"log"
	"strconv"
	"strings"

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

func (job *ListComputeInstancesJob) Properties() sdk.JobProperties {
	return sdk.JobProperties{
		ID:          "list_compute_instances",
		Description: "List all compute instances in current project",
		MaxRetry:    0,
	}
}

func (job *ListComputeInstancesJob) Run(ctx context.Context) error {
	log.Println("Running list compute instance job")

	instances, err := job.processor.provider.GetAllInstances(ctx)
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

			diskDetails, err := job.processor.provider.GetDiskDetails(ctx, instanceZone, diskName)
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
			Preemptible:         *instance.Scheduling.Preemptible,
			OptimizationLoading: true,
			Preferences:         job.processor.defaultPreferences,
			Skipped:             false,
			LazyLoadingEnabled:  false,
			SkipReason:          "NA",
			Instance:            instance,
			Disks:               disks,
			Metrics:             nil,
			DisksMetrics:        nil,
		}

		if !oi.Skipped {
			job.processor.lazyloadCounter.Add(1)
			if job.processor.lazyloadCounter.Load() > uint32(1) {
				oi.LazyLoadingEnabled = true
			}
		}

		job.processor.items.Set(oi.Id, oi)
		job.processor.publishOptimizationItem(oi.ToOptimizationItem())
		job.processor.UpdateSummary(oi.Id)

		job.processor.jobQueue.Push(NewGetComputeInstanceMetricsJob(job.processor, oi.Id))
	}

	return nil

}
