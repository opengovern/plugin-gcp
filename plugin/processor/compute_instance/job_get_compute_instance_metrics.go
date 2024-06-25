package compute_instance

import (
	"context"
	"fmt"
	"google.golang.org/api/compute/v1"
	"log"
	"strconv"
	"time"

	"github.com/kaytu-io/plugin-gcp/plugin/kaytu"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/kaytu-io/plugin-gcp/plugin/preferences"
	util "github.com/kaytu-io/plugin-gcp/utils"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetComputeInstanceMetricsJob struct {
	processor *ComputeInstanceProcessor
	instance  *computepb.Instance
	disks     []compute.Disk
}

func NewGetComputeInstanceMetricsJob(processor *ComputeInstanceProcessor, instance *computepb.Instance, disks []compute.Disk) *GetComputeInstanceMetricsJob {
	return &GetComputeInstanceMetricsJob{
		processor: processor,
		instance:  instance,
		disks:     disks,
	}
}

func (job *GetComputeInstanceMetricsJob) Id() string {
	return fmt.Sprintf("get_compute_instance_metrics_%d", job.instance.GetId())
}

func (job *GetComputeInstanceMetricsJob) Description() string {
	return fmt.Sprintf("Get metrics for compute instance: %d", job.instance.GetId())

}

func (job *GetComputeInstanceMetricsJob) Run(ctx context.Context) error {

	endTime := time.Now()                         // end time of requested time series
	startTime := endTime.Add(-24 * 1 * time.Hour) // start time of requested time series

	// metricName string,
	// instanceID string,
	// startTime time.Time,
	// endTime time.Time,
	// periodInSeconds int64,
	cpuRequest := job.processor.metricProvider.NewTimeSeriesRequest(
		fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			"compute.googleapis.com/instance/cpu/utilization", // fully qualified name of the metric
			fmt.Sprint(job.instance.GetId()),                  // compute instance ID
		),
		&monitoringpb.TimeInterval{
			EndTime:   timestamppb.New(endTime),
			StartTime: timestamppb.New(startTime),
		},
		&monitoringpb.Aggregation{
			AlignmentPeriod: &durationpb.Duration{ // period, for which the datapoints will be aggregated into one, in seconds
				Seconds: 60,
			},
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_MEAN, // will represent all the datapoints in the above period, with a mean
		},
	)

	cpumetric, err := job.processor.metricProvider.GetMetric(cpuRequest)
	if err != nil {
		return err
	}

	memoryRequest := job.processor.metricProvider.NewTimeSeriesRequest(
		fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			"compute.googleapis.com/instance/memory/balloon/ram_used",
			fmt.Sprint(job.instance.GetId()),
		),
		&monitoringpb.TimeInterval{
			EndTime:   timestamppb.New(endTime),
			StartTime: timestamppb.New(startTime),
		},
		&monitoringpb.Aggregation{
			AlignmentPeriod: &durationpb.Duration{
				Seconds: 60,
			},
			PerSeriesAligner: monitoringpb.Aggregation_ALIGN_MEAN, // will represent all the datapoints in the above period, with a mean
		},
	)

	memoryMetric, err := job.processor.metricProvider.GetMetric(memoryRequest)
	if err != nil {
		return err
	}

	instanceMetrics := make(map[string][]kaytu.Datapoint)

	instanceMetrics["cpuUtilization"] = cpumetric
	instanceMetrics["memoryUtilization"] = memoryMetric

	oi := ComputeInstanceItem{
		ProjectId:           job.processor.provider.ProjectID,
		Name:                *job.instance.Name,
		Id:                  strconv.FormatUint(job.instance.GetId(), 10),
		MachineType:         util.TrimmedString(*job.instance.MachineType, "/"),
		Region:              util.TrimmedString(*job.instance.Zone, "/"),
		Platform:            job.instance.GetCpuPlatform(),
		OptimizationLoading: false,
		Preferences:         preferences.DefaultComputeEnginePreferences,
		Skipped:             false,
		LazyLoadingEnabled:  false,
		SkipReason:          "NA",
		Disks:               job.disks,
		Metrics:             instanceMetrics,
	}

	for k, v := range oi.Metrics {
		log.Printf("%s : %d", k, len(v))
	}

	job.processor.items.Set(oi.Id, oi)
	job.processor.publishOptimizationItem(oi.ToOptimizationItem())

	job.processor.jobQueue.Push(NewOptimizeComputeInstancesJob(job.processor, oi))

	return nil

}
