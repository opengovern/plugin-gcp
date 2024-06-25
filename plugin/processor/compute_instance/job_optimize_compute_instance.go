package compute_instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/preferences"
	"github.com/kaytu-io/plugin-gcp/plugin/kaytu"
	"github.com/kaytu-io/plugin-gcp/plugin/version"
)

type OptimizeComputeInstancesJob struct {
	processor *ComputeInstanceProcessor
	item      ComputeInstanceItem
}

func NewOptimizeComputeInstancesJob(processor *ComputeInstanceProcessor, item ComputeInstanceItem) *OptimizeComputeInstancesJob {
	return &OptimizeComputeInstancesJob{
		processor: processor,
		item:      item,
	}
}

func (job *OptimizeComputeInstancesJob) Id() string {
	return fmt.Sprintf("optimize_compute_isntance_%s", job.item.Id)
}

func (job *OptimizeComputeInstancesJob) Description() string {
	return fmt.Sprintf("Optimizing %s", job.item.Id)

}

func (job *OptimizeComputeInstancesJob) Run(ctx context.Context) error {

	requestId := uuid.NewString()

	var disks []kaytu.GcpComputeDisk
	diskFilled := make(map[string]float64)
	for _, disk := range job.item.Disks {
		id := strconv.FormatUint(disk.Id, 10)
		disks = append(disks, kaytu.GcpComputeDisk{
			HashedDiskId:    id,
			DiskSize:        &disk.SizeGb,
			DiskType:        disk.Type,
			Region:          disk.Region,
			ProvisionedIops: &disk.ProvisionedIops,
			Zone:            disk.Zone,
		})
		diskFilled[id] = 0
	}

	request := kaytu.GcpComputeInstanceWastageRequest{
		RequestId:      &requestId,
		CliVersion:     &version.VERSION,
		Identification: job.processor.provider.Identify(),
		Instance: kaytu.GcpComputeInstance{
			HashedInstanceId: utils.HashString(job.item.Id),
			Zone:             job.item.Region,
			MachineType:      job.item.MachineType,
		},
		Disks:       disks,
		Metrics:     job.item.Metrics,
		Region:      job.item.Region,
		Preferences: preferences.Export(job.item.Preferences),
		Loading:     false,
	}

	response, err := kaytu.Ec2InstanceWastageRequest(request, job.processor.kaytuAcccessToken)
	if err != nil {
		return err
	}

	job.item = ComputeInstanceItem{
		ProjectId:           job.item.ProjectId,
		Name:                job.item.Name,
		Id:                  job.item.Id,
		MachineType:         job.item.MachineType,
		Region:              job.item.Region,
		Platform:            job.item.Platform,
		OptimizationLoading: false,
		Preferences:         job.item.Preferences,
		Skipped:             false,
		LazyLoadingEnabled:  false,
		SkipReason:          "NA",
		Metrics:             job.item.Metrics,
		Disks:               job.item.Disks,
		Wastage:             *response,
	}

	job.processor.items.Set(job.item.Id, job.item)
	job.processor.publishOptimizationItem(job.item.ToOptimizationItem())

	return nil
}
