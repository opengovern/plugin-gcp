package compute_instance

import (
	"context"
	"fmt"
	golang2 "github.com/kaytu-io/plugin-gcp/plugin/proto/src/golang"
	"google.golang.org/api/compute/v1"
	"log"
	"strconv"
	"time"

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

	disksMetrics := make(map[string]map[string][]*golang2.DataPoint)
	for _, disk := range job.disks {
		id := strconv.FormatUint(disk.Id, 10)
		disksMetrics[id] = make(map[string][]*golang2.DataPoint)

		diskReadIopsRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/read_ops_count",
				fmt.Sprint(job.instance.GetId()), disk.Name),
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

		diskReadIopsMetrics, err := job.processor.metricProvider.GetMetric(diskReadIopsRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskReadIOPS"] = diskReadIopsMetrics

		diskWriteIopsRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/write_ops_count",
				fmt.Sprint(job.instance.GetId()), disk.Name),
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

		diskWriteIopsMetrics, err := job.processor.metricProvider.GetMetric(diskWriteIopsRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskWriteIOPS"] = diskWriteIopsMetrics

		diskReadThroughputRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/read_bytes_count",
				fmt.Sprint(job.instance.GetId()), disk.Name),
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

		diskReadThroughputMetrics, err := job.processor.metricProvider.GetMetric(diskReadThroughputRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskReadThroughput"] = diskReadThroughputMetrics

		diskWriteThroughputRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/write_bytes_count",
				fmt.Sprint(job.instance.GetId()), disk.Name),
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

		diskWriteThroughputMetrics, err := job.processor.metricProvider.GetMetric(diskWriteThroughputRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskWriteThroughput"] = diskWriteThroughputMetrics
	}

	instanceMetrics := make(map[string][]*golang2.DataPoint)

	instanceMetrics["cpuUtilization"] = cpumetric
	instanceMetrics["memoryUtilization"] = memoryMetric

	oi := ComputeInstanceItem{
		ProjectId:           job.processor.provider.ProjectID,
		Name:                *job.instance.Name,
		Id:                  strconv.FormatUint(job.instance.GetId(), 10),
		MachineType:         util.TrimmedString(*job.instance.MachineType, "/"),
		Region:              util.TrimmedString(*job.instance.Zone, "/"),
		Platform:            job.instance.GetCpuPlatform(),
		OptimizationLoading: true,
		Preferences:         preferences.DefaultComputeEnginePreferences,
		Skipped:             false,
		LazyLoadingEnabled:  false,
		SkipReason:          "NA",
		Instance:            job.instance,
		Disks:               job.disks,
		Metrics:             instanceMetrics,
		DisksMetrics:        disksMetrics,
	}

	for d, v := range oi.DisksMetrics {
		for k, v := range v {
			log.Printf("%s %s : %d", d, k, len(v))
		}
	}

	job.processor.items.Set(oi.Id, oi)
	job.processor.publishOptimizationItem(oi.ToOptimizationItem())

	job.processor.jobQueue.Push(NewOptimizeComputeInstancesJob(job.processor, oi))
	job.processor.UpdateSummary(oi.Id)

	return nil

}
