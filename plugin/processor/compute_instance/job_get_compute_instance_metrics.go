package compute_instance

import (
	"context"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	golang2 "github.com/kaytu-io/plugin-gcp/plugin/proto/src/golang/gcp"
	"log"
	"strconv"
	"time"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetComputeInstanceMetricsJob struct {
	processor *ComputeInstanceProcessor
	itemId    string
}

func NewGetComputeInstanceMetricsJob(processor *ComputeInstanceProcessor, itemId string) *GetComputeInstanceMetricsJob {
	return &GetComputeInstanceMetricsJob{
		processor: processor,
		itemId:    itemId,
	}
}

func (job *GetComputeInstanceMetricsJob) Properties() sdk.JobProperties {
	return sdk.JobProperties{
		ID:          fmt.Sprintf("get_compute_instance_metrics_%s", job.itemId),
		Description: fmt.Sprintf("Get metrics for compute instance: %s", job.itemId),
		MaxRetry:    0,
	}
}

func (job *GetComputeInstanceMetricsJob) Run(ctx context.Context) error {
	item, ok := job.processor.items.Get(job.itemId)
	if !ok {
		return fmt.Errorf("item not found %s", job.itemId)
	}

	endTime := time.Now()                         // end time of requested time series
	startTime := endTime.Add(-24 * 1 * time.Hour) // start time of requested time series

	cpuRequest := job.processor.metricProvider.NewTimeSeriesRequest(
		fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			"compute.googleapis.com/instance/cpu/utilization", // fully qualified name of the metric
			fmt.Sprint(item.Instance.GetId()),                 // compute instance ID
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

	cpumetric, err := job.processor.metricProvider.GetMetric(ctx, cpuRequest)
	if err != nil {
		return err
	}

	memoryRequest := job.processor.metricProvider.NewTimeSeriesRequest(
		fmt.Sprintf(
			`metric.type="%s" AND resource.labels.instance_id="%s"`,
			"compute.googleapis.com/instance/memory/balloon/ram_used",
			fmt.Sprint(item.Instance.GetId()),
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

	memoryMetric, err := job.processor.metricProvider.GetMetric(ctx, memoryRequest)
	if err != nil {
		return err
	}

	disksMetrics := make(map[string]map[string][]*golang2.DataPoint)
	for _, disk := range item.Disks {
		id := strconv.FormatUint(disk.Id, 10)
		disksMetrics[id] = make(map[string][]*golang2.DataPoint)

		diskReadIopsRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/read_ops_count",
				fmt.Sprint(item.Instance.GetId()), disk.Name),
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

		diskReadIopsMetrics, err := job.processor.metricProvider.GetMetric(ctx, diskReadIopsRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskReadIOPS"] = diskReadIopsMetrics

		diskWriteIopsRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/write_ops_count",
				fmt.Sprint(item.Instance.GetId()), disk.Name),
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

		diskWriteIopsMetrics, err := job.processor.metricProvider.GetMetric(ctx, diskWriteIopsRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskWriteIOPS"] = diskWriteIopsMetrics

		diskReadThroughputRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/read_bytes_count",
				fmt.Sprint(item.Instance.GetId()), disk.Name),
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

		diskReadThroughputMetrics, err := job.processor.metricProvider.GetMetric(ctx, diskReadThroughputRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskReadThroughput"] = diskReadThroughputMetrics

		diskWriteThroughputRequest := job.processor.metricProvider.NewTimeSeriesRequest(
			fmt.Sprintf(
				`metric.type="%s" AND resource.labels.instance_id="%s" AND metric.labels.device_name="%s"`,
				"compute.googleapis.com/instance/disk/write_bytes_count",
				fmt.Sprint(item.Instance.GetId()), disk.Name),
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

		diskWriteThroughputMetrics, err := job.processor.metricProvider.GetMetric(ctx, diskWriteThroughputRequest)
		if err != nil {
			return err
		}

		disksMetrics[id]["DiskWriteThroughput"] = diskWriteThroughputMetrics
	}

	instanceMetrics := make(map[string][]*golang2.DataPoint)

	instanceMetrics["cpuUtilization"] = cpumetric
	instanceMetrics["memoryUtilization"] = memoryMetric

	item.OptimizationLoading = true
	item.Skipped = false
	item.SkipReason = "N/A"
	item.LazyLoadingEnabled = false
	item.Metrics = instanceMetrics
	item.DisksMetrics = disksMetrics

	for d, v := range item.DisksMetrics {
		for k, v := range v {
			log.Printf("%s %s : %d", d, k, len(v))
		}
	}

	job.processor.items.Set(item.Id, item)
	job.processor.publishOptimizationItem(item.ToOptimizationItem())
	job.processor.UpdateSummary(item.Id)

	job.processor.jobQueue.Push(NewOptimizeComputeInstancesJob(job.processor, item.Id))

	return nil

}
