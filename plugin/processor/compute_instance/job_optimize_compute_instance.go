package compute_instance

import (
	"context"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"github.com/opengovern/plugin-gcp/plugin/processor/shared"
	golang2 "github.com/opengovern/plugin-gcp/plugin/proto/src/golang/gcp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/preferences"
	"github.com/opengovern/plugin-gcp/plugin/version"
)

type OptimizeComputeInstancesJob struct {
	processor *ComputeInstanceProcessor
	itemId    string
}

func NewOptimizeComputeInstancesJob(processor *ComputeInstanceProcessor, itemId string) *OptimizeComputeInstancesJob {
	return &OptimizeComputeInstancesJob{
		processor: processor,
		itemId:    itemId,
	}
}

func (job *OptimizeComputeInstancesJob) Properties() sdk.JobProperties {
	return sdk.JobProperties{
		ID:          fmt.Sprintf("optimize_compute_isntance_%s", job.itemId),
		Description: fmt.Sprintf("Optimizing %s", job.itemId),
		MaxRetry:    3,
	}
}

func (job *OptimizeComputeInstancesJob) Run(ctx context.Context) error {
	item, ok := job.processor.items.Get(job.itemId)
	if !ok {
		return fmt.Errorf("item not found %s", job.itemId)
	}

	if item.LazyLoadingEnabled {
		job.processor.jobQueue.Push(NewGetComputeInstanceMetricsJob(job.processor, job.itemId))
		return nil
	}

	requestId := uuid.NewString()

	var disks []*golang2.GcpComputeDisk
	diskFilled := make(map[string]float64)
	for _, disk := range item.Disks {
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
	for k, v := range preferences.Export(item.Preferences) {
		preferencesMap[k] = nil
		if v != nil {
			preferencesMap[k] = wrapperspb.String(*v)
		}
	}

	metrics := make(map[string]*golang2.Metric)
	for k, v := range item.Metrics {
		metrics[k] = &golang2.Metric{
			Data: v,
		}
	}

	diskMetrics := make(map[string]*golang2.DiskMetrics)
	for disk, m := range item.DisksMetrics {
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

	grpcCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("workspace-name", "kaytu"))
	grpcCtx, cancel := context.WithTimeout(grpcCtx, shared.GrpcOptimizeRequestTimeout)
	defer cancel()
	response, err := job.processor.client.GCPComputeOptimization(grpcCtx, &golang2.GCPComputeOptimizationRequest{
		RequestId:      wrapperspb.String(requestId),
		CliVersion:     wrapperspb.String(version.VERSION),
		Identification: job.processor.provider.Identify(),
		Instance: &golang2.GcpComputeInstance{
			Id:                utils.HashString(item.Id),
			Zone:              item.Region,
			MachineType:       item.MachineType,
			Preemptible:       item.Preemptible,
			InstanceOsLicense: item.InstanceOsLicense,
		},
		Disks:        disks,
		Preferences:  preferencesMap,
		Metrics:      metrics,
		DisksMetrics: diskMetrics,
		Loading:      false,
		Region:       item.Region,
	})
	if err != nil {
		return err
	}

	item.OptimizationLoading = false
	item.Skipped = false
	item.SkipReason = "N/A"
	item.LazyLoadingEnabled = false
	item.Wastage = response

	job.processor.items.Set(job.itemId, item)
	job.processor.publishOptimizationItem(item.ToOptimizationItem())
	job.processor.UpdateSummary(item.Id)

	return nil
}
