package compute_instance

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/kaytu-io/plugin-gcp/plugin/preferences"
)

type GetComputeInstanceMetricsJob struct {
	processor *ComputeInstanceProcessor
	instance  *computepb.Instance
}

func NewGetComputeInstanceMetricsJob(processor *ComputeInstanceProcessor, instance *computepb.Instance) *GetComputeInstanceMetricsJob {
	return &GetComputeInstanceMetricsJob{
		processor: processor,
		instance:  instance,
	}
}

func (job *GetComputeInstanceMetricsJob) Id() string {
	return "get_compute_instance_metrics"
}

func (job *GetComputeInstanceMetricsJob) Description() string {
	return "Get metrics for the specified compute instance"

}

func (job *GetComputeInstanceMetricsJob) Run() error {

	err := job.processor.metricProvider.InitializeClient(context.Background())
	if err != nil {
		return err
	}

	endtime := time.Now()
	starttime := endtime.Add(-24 * 1 * time.Hour)

	cpuRequest := job.processor.metricProvider.NewInstanceMetricRequest(
		"compute.googleapis.com/instance/cpu/utilization",
		fmt.Sprint(job.instance.GetId()),
		starttime,
		endtime,
		60, // 1 minute
	)

	cpumetric, err := job.processor.metricProvider.GetMetric(cpuRequest)
	if err != nil {
		return err
	}

	memoryRequest := job.processor.metricProvider.NewInstanceMetricRequest(
		"compute.googleapis.com/instance/cpu/utilization",
		fmt.Sprint(job.instance.GetId()),
		starttime,
		endtime,
		60, // 1 minute
	)

	memoryMetric, err := job.processor.metricProvider.GetMetric(memoryRequest)
	if err != nil {
		return err
	}

	instanceMetrics := make(map[string][]*monitoringpb.Point)

	instanceMetrics["cpuUtilization"] = cpumetric.GetPoints()
	instanceMetrics["memoryUtilization"] = memoryMetric.GetPoints()

	oi := ComputeInstanceItem{
		Name:                *job.instance.Name,
		Id:                  strconv.FormatUint(job.instance.GetId(), 10),
		MachineType:         job.instance.GetMachineType(),
		Region:              job.instance.GetZone(),
		OptimizationLoading: false,
		Preferences:         preferences.DefaultComputeEnginePreferences,
		Skipped:             false,
		LazyLoadingEnabled:  false,
		SkipReason:          "NA",
		Metrics:             instanceMetrics,
	}

	for k, v := range oi.Metrics {
		log.Printf("%s : %d", k, len(v))
	}

	job.processor.items.Set(oi.Id, oi)
	job.processor.publishOptimizationItem(oi.ToOptimizationItem())

	return nil

}
