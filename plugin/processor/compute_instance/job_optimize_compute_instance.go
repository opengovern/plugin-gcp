package compute_instance

import (
	"context"
	"fmt"
	"github.com/kaytu-io/plugin-gcp/plugin/processor/shared"
	golang2 "github.com/kaytu-io/plugin-gcp/plugin/proto/src/golang/gcp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/preferences"
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
	if job.item.LazyLoadingEnabled {
		job.processor.jobQueue.Push(NewGetComputeInstanceMetricsJob(job.processor, job.item.Instance, job.item.Disks))
		return nil
	}

	requestId := uuid.NewString()

	var disks []*golang2.GcpComputeDisk
	diskFilled := make(map[string]float64)
	for _, disk := range job.item.Disks {
		id := strconv.FormatUint(disk.Id, 10)
		typeURLParts := strings.Split(disk.Type, "/")
		diskType := typeURLParts[len(typeURLParts)-1]

		zoneURLParts := strings.Split(disk.Zone, "/")
		diskZone := zoneURLParts[len(zoneURLParts)-1]
		region := strings.Join([]string{strings.Split(diskZone, "-")[0], strings.Split(diskZone, "-")[1]}, "-")

		disks = append(disks, &golang2.GcpComputeDisk{
			Id:              id,
			DiskSize:        wrapperspb.Int64(disk.SizeGb),
			DiskType:        diskType,
			Region:          region,
			ProvisionedIops: wrapperspb.Int64(disk.ProvisionedIops),
			Zone:            diskZone,
		})
		diskFilled[id] = 0
	}

	preferencesMap := map[string]*wrapperspb.StringValue{}
	for k, v := range preferences.Export(job.item.Preferences) {
		preferencesMap[k] = nil
		if v != nil {
			preferencesMap[k] = wrapperspb.String(*v)
		}
	}

	grpcCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("workspace-name", "kaytu"))
	grpcCtx, cancel := context.WithTimeout(grpcCtx, shared.GrpcOptimizeRequestTimeout)
	defer cancel()

	metrics := make(map[string]*golang2.Metric)
	for k, v := range job.item.Metrics {
		metrics[k] = &golang2.Metric{
			Data: v,
		}
	}

	diskMetrics := make(map[string]*golang2.DiskMetrics)
	for disk, m := range job.item.DisksMetrics {
		diskM := make(map[string]*golang2.Metric)
		for k, v := range m {
			diskM[k] = &golang2.Metric{
				Data: v,
			}
		}
		diskMetrics[disk] = &golang2.DiskMetrics{
			Metrics: diskM,
		}
	}

	response, err := job.processor.client.GCPComputeOptimization(ctx, &golang2.GCPComputeOptimizationRequest{
		RequestId:      wrapperspb.String(requestId),
		CliVersion:     wrapperspb.String(version.VERSION),
		Identification: job.processor.provider.Identify(),
		Instance: &golang2.GcpComputeInstance{
			Id:          utils.HashString(job.item.Id),
			Zone:        job.item.Region,
			MachineType: job.item.MachineType,
		},
		Disks:        disks,
		Preferences:  preferencesMap,
		Metrics:      metrics,
		DisksMetrics: diskMetrics,
		Loading:      false,
		Region:       job.item.Region,
	})
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
		DisksMetrics:        job.item.DisksMetrics,
		Instance:            job.item.Instance,
		Disks:               job.item.Disks,
		Wastage:             *response,
	}

	job.processor.items.Set(job.item.Id, job.item)
	job.processor.publishOptimizationItem(job.item.ToOptimizationItem())
	job.processor.UpdateSummary(job.item.Id)

	return nil
}
