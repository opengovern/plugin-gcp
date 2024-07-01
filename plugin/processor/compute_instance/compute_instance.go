package compute_instance

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/plugin-gcp/plugin/gcp"
	golang2 "github.com/kaytu-io/plugin-gcp/plugin/proto/src/golang"
	util "github.com/kaytu-io/plugin-gcp/utils"
	"strconv"
	"strings"
	"sync/atomic"
)

type ComputeInstanceProcessor struct {
	provider                *gcp.Compute
	metricProvider          *gcp.CloudMonitoring
	items                   util.ConcurrentMap[string, ComputeInstanceItem]
	publishOptimizationItem func(item *golang.ChartOptimizationItem)
	publishResultSummary    func(summary *golang.ResultSummary)
	kaytuAcccessToken       string
	jobQueue                *sdk.JobQueue
	lazyloadCounter         atomic.Uint32
	client                  golang2.OptimizationClient

	summary util.ConcurrentMap[string, ComputeInstanceSummary]
}

func NewComputeInstanceProcessor(
	prv *gcp.Compute,
	metricPrv *gcp.CloudMonitoring,
	publishOptimizationItem func(item *golang.ChartOptimizationItem),
	publishResultSummary func(summary *golang.ResultSummary),
	kaytuAcccessToken string,
	jobQueue *sdk.JobQueue,
	client golang2.OptimizationClient,
) *ComputeInstanceProcessor {
	r := &ComputeInstanceProcessor{
		provider:                prv,
		metricProvider:          metricPrv,
		items:                   util.NewMap[string, ComputeInstanceItem](),
		publishOptimizationItem: publishOptimizationItem,
		publishResultSummary:    publishResultSummary,
		kaytuAcccessToken:       kaytuAcccessToken,
		jobQueue:                jobQueue,
		lazyloadCounter:         atomic.Uint32{},
		client:                  client,
	}

	jobQueue.Push(NewListComputeInstancesJob(r))
	return r
}

func (m *ComputeInstanceProcessor) ReEvaluate(id string, items []*golang.PreferenceItem) {
	v, _ := m.items.Get(id)
	v.Preferences = items
	m.items.Set(id, v)
	v.OptimizationLoading = true
	m.publishOptimizationItem(v.ToOptimizationItem())
	m.jobQueue.Push(NewOptimizeComputeInstancesJob(m, v))
}

func (m *ComputeInstanceProcessor) ExportNonInteractive() *golang.NonInteractiveExport {
	return &golang.NonInteractiveExport{
		Csv: m.exportCsv(),
	}
}

func (m *ComputeInstanceProcessor) exportCsv() []*golang.CSVRow {
	headers := []string{
		"Project ID", "Region", "Resource Type", "Resource ID", "Resource Name", "Platform",
		"Device Runtime (Hrs)", "Current Cost", "Recommendation Cost", "Net Savings",
		"Current Spec", "Suggested Spec", "Parent Device", "Justification", "Additional Details",
	}
	var rows []*golang.CSVRow
	rows = append(rows, &golang.CSVRow{Row: headers})

	m.items.Range(func(key string, value ComputeInstanceItem) bool {
		var additionalDetails []string
		var rightSizingCost, saving, recSpec string
		if value.Wastage.RightSizing.Recommended != nil {
			rightSizingCost = utils.FormatPriceFloat(value.Wastage.RightSizing.Recommended.Cost)
			saving = utils.FormatPriceFloat(value.Wastage.RightSizing.Current.Cost - value.Wastage.RightSizing.Recommended.Cost)
			recSpec = value.Wastage.RightSizing.Recommended.MachineType

			additionalDetails = append(additionalDetails,
				fmt.Sprintf("Machine Type:: Current: %s - Recommended: %s", value.Wastage.RightSizing.Current.MachineType,
					value.Wastage.RightSizing.Recommended.MachineType))
			additionalDetails = append(additionalDetails,
				fmt.Sprintf("Region:: Current: %s - Recommended: %s", value.Wastage.RightSizing.Current.Region,
					value.Wastage.RightSizing.Recommended.Region))
			additionalDetails = append(additionalDetails,
				fmt.Sprintf("CPU:: Current: %d - Recommended: %d", value.Wastage.RightSizing.Current.CPU,
					value.Wastage.RightSizing.Recommended.CPU))
			additionalDetails = append(additionalDetails,
				fmt.Sprintf("Memory:: Current: %d - Recommended: %d", value.Wastage.RightSizing.Current.MemoryMb,
					value.Wastage.RightSizing.Recommended.MemoryMb))
		}
		computeRow := []string{
			value.ProjectId, value.Region, "Compute Instance", value.Id, value.Name, value.Platform,
			"730 Hrs", utils.FormatPriceFloat(value.Wastage.RightSizing.Current.Cost), rightSizingCost, saving,
			value.Wastage.RightSizing.Current.MachineType, recSpec, "None", value.Wastage.RightSizing.Description, strings.Join(additionalDetails, "---")}

		rows = append(rows, &golang.CSVRow{Row: computeRow})

		for _, d := range value.Disks {
			dKey := strconv.FormatUint(d.Id, 10)
			disk := value.Wastage.VolumeRightSizing[dKey]
			var diskAdditionalDetails []string
			var diskRightSizingCost, diskSaving, diskRecSpec string
			if disk.Recommended != nil {
				diskRightSizingCost = utils.FormatPriceFloat(disk.Recommended.Cost)
				diskSaving = utils.FormatPriceFloat(disk.Current.Cost - disk.Recommended.Cost)
				diskRecSpec = fmt.Sprintf("%s / %d GB", disk.Recommended.DiskType, disk.Recommended.DiskSize)

				diskAdditionalDetails = append(diskAdditionalDetails,
					fmt.Sprintf("Region:: Current: %s - Recommended: %s", disk.Current.Region,
						disk.Recommended.Region))
				diskAdditionalDetails = append(diskAdditionalDetails,
					fmt.Sprintf("ReadIopsExpectation:: Current: %d - Recommended: %d", disk.Current.ReadIopsLimit,
						disk.Recommended.ReadIopsLimit))
				diskAdditionalDetails = append(diskAdditionalDetails,
					fmt.Sprintf("WriteIopsExpectation:: Current: %d - Recommended: %d", disk.Current.WriteIopsLimit,
						disk.Recommended.WriteIopsLimit))
				diskAdditionalDetails = append(diskAdditionalDetails,
					fmt.Sprintf("ReadThroughputExpectation:: Current: %.2f - Recommended: %.2f", disk.Current.ReadThroughputLimit,
						disk.Recommended.ReadThroughputLimit))
				diskAdditionalDetails = append(diskAdditionalDetails,
					fmt.Sprintf("WriteThroughputExpectation:: Current: %.2f - Recommended: %.2f", disk.Current.WriteThroughputLimit,
						disk.Recommended.WriteThroughputLimit))
			}
			diskRow := []string{
				value.ProjectId, value.Region, "Compute Disk", dKey, d.Name, "N/A",
				"730 Hrs", utils.FormatPriceFloat(disk.Current.Cost), diskRightSizingCost, diskSaving,
				fmt.Sprintf("%s / %d GB", disk.Current.DiskType, disk.Current.DiskSize), diskRecSpec,
				"None", value.Wastage.RightSizing.Description, strings.Join(diskAdditionalDetails, "---")}

			rows = append(rows, &golang.CSVRow{Row: diskRow})
		}

		return true
	})
	return rows
}

func (m *ComputeInstanceProcessor) ResultsSummary() *golang.ResultSummary {
	summary := &golang.ResultSummary{}
	var totalCost, savings float64
	m.summary.Range(func(_ string, item ComputeInstanceSummary) bool {
		totalCost += item.CurrentRuntimeCost
		savings += item.Savings
		return true
	})

	summary.Message = fmt.Sprintf("Current runtime cost: %s, Savings: %s",
		style.CostStyle.Render(fmt.Sprintf("%s", utils.FormatPriceFloat(totalCost))), style.SavingStyle.Render(fmt.Sprintf("%s", utils.FormatPriceFloat(savings))))
	return summary
}

func (m *ComputeInstanceProcessor) UpdateSummary(itemId string) {
	i, ok := m.items.Get(itemId)
	if ok && i.Wastage.RightSizing.Recommended != nil {
		totalSaving := 0.0
		totalCurrentCost := 0.0
		for _, v := range i.Wastage.VolumeRightSizing {
			totalSaving += v.Current.Cost - v.Recommended.Cost
			totalCurrentCost += v.Current.Cost
		}
		totalSaving += i.Wastage.RightSizing.Current.Cost - i.Wastage.RightSizing.Recommended.Cost
		totalCurrentCost += i.Wastage.RightSizing.Current.Cost

		m.summary.Set(itemId, ComputeInstanceSummary{
			CurrentRuntimeCost: totalCurrentCost,
			Savings:            totalSaving,
		})
	}
	m.publishResultSummary(m.ResultsSummary())
}
