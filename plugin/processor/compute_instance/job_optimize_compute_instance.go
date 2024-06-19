package compute_instance

import (
	"context"
	"fmt"

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

	request := kaytu.GcpComputeInstanceWastageRequest{
		RequestId:      &requestId,
		CliVersion:     &version.VERSION,
		Identification: job.processor.provider.Identify(),
		Instance: kaytu.GcpComputeInstance{
			HashedInstanceId: utils.HashString(job.item.Id),
			Zone:             job.item.Region,
			MachineType:      job.item.MachineType,
		},
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
		Name:                job.item.Name,
		Id:                  job.item.Id,
		MachineType:         job.item.MachineType,
		Region:              job.item.Region,
		OptimizationLoading: false,
		Preferences:         job.item.Preferences,
		Skipped:             false,
		LazyLoadingEnabled:  false,
		SkipReason:          "NA",
		Metrics:             job.item.Metrics,
		Wastage:             *response,
	}

	job.processor.items.Set(job.item.Id, job.item)
	job.processor.publishOptimizationItem(job.item.ToOptimizationItem())

	return nil
}
