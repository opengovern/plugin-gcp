package compute_instance

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"github.com/kaytu-io/kaytu/pkg/style"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/opengovern/plugin-gcp/plugin/gcp"
	golang2 "github.com/opengovern/plugin-gcp/plugin/proto/src/golang/gcp"
	"strconv"
	"strings"
	"sync/atomic"
)

type ComputeInstanceProcessor struct {
	provider                *gcp.Compute
	metricProvider          *gcp.CloudMonitoring
	items                   utils.ConcurrentMap[string, ComputeInstanceItem]
	publishOptimizationItem func(item *golang.ChartOptimizationItem)
	publishResultSummary    func(summary *golang.ResultSummary)
	kaytuAcccessToken       string
	jobQueue                *sdk.JobQueue
	lazyloadCounter         atomic.Uint32
	client                  golang2.OptimizationClient

	defaultPreferences []*golang.PreferenceItem

	summary utils.ConcurrentMap[string, ComputeInstanceSummary]
}

func NewComputeInstanceProcessor(
	prv *gcp.Compute,
	metricPrv *gcp.CloudMonitoring,
	publishOptimizationItem func(item *golang.ChartOptimizationItem),
	publishResultSummary func(summary *golang.ResultSummary),
	kaytuAcccessToken string,
	jobQueue *sdk.JobQueue,
	client golang2.OptimizationClient,
	defaultPreferences []*golang.PreferenceItem,
) *ComputeInstanceProcessor {
	r := &ComputeInstanceProcessor{
		provider:                prv,
		metricProvider:          metricPrv,
		items:                   utils.NewConcurrentMap[string, ComputeInstanceItem](),
		publishOptimizationItem: publishOptimizationItem,
		publishResultSummary:    publishResultSummary,
		kaytuAcccessToken:       kaytuAcccessToken,
		jobQueue:                jobQueue,
		lazyloadCounter:         atomic.Uint32{},
		client:                  client,
		defaultPreferences:      defaultPreferences,
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
	m.jobQueue.Push(NewOptimizeComputeInstancesJob(m, id))
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
		if value.Wastage.Rightsizing != nil && value.Wastage.Rightsizing.Recommended != nil {
			rightSizingCost = utils.FormatPriceFloat(value.Wastage.Rightsizing.Recommended.Cost)
			saving = utils.FormatPriceFloat(value.Wastage.Rightsizing.Current.Cost - value.Wastage.Rightsizing.Recommended.Cost)
			recSpec = value.Wastage.Rightsizing.Recommended.MachineType

			additionalDetails = append(additionalDetails,
				fmt.Sprintf("Machine Type:: Current: %s - Recommended: %s", value.Wastage.Rightsizing.Current.MachineType,
					value.Wastage.Rightsizing.Recommended.MachineType))
			additionalDetails = append(additionalDetails,
				fmt.Sprintf("Region:: Current: %s - Recommended: %s", value.Wastage.Rightsizing.Current.Region,
					value.Wastage.Rightsizing.Recommended.Region))
			additionalDetails = append(additionalDetails,
				fmt.Sprintf("CPU:: Current: %d - Recommended: %d", value.Wastage.Rightsizing.Current.Cpu,
					value.Wastage.Rightsizing.Recommended.Cpu))
			additionalDetails = append(additionalDetails,
				fmt.Sprintf("Memory:: Current: %d - Recommended: %d", value.Wastage.Rightsizing.Current.MemoryMb,
					value.Wastage.Rightsizing.Recommended.MemoryMb))
		}
		computeRow := []string{
			value.ProjectId, value.Region, "Compute Instance", value.Id, value.Name, value.Platform,
			"730 Hrs", utils.FormatPriceFloat(value.Wastage.Rightsizing.Current.Cost), rightSizingCost, saving,
			value.Wastage.Rightsizing.Current.MachineType, recSpec, "None", value.Wastage.Rightsizing.Description, strings.Join(additionalDetails, "---")}

		rows = append(rows, &golang.CSVRow{Row: computeRow})

		for _, d := range value.Disks {
			dKey := strconv.FormatUint(d.Id, 10)
			disk := value.Wastage.VolumesRightsizing[dKey]
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
				"None", value.Wastage.Rightsizing.Description, strings.Join(diskAdditionalDetails, "---")}

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
	if ok && i.Wastage != nil && i.Wastage.Rightsizing.Recommended != nil {
		totalSaving := 0.0
		totalCurrentCost := 0.0
		for _, v := range i.Wastage.VolumesRightsizing {
			totalSaving += v.Current.Cost - v.Recommended.Cost
			totalCurrentCost += v.Current.Cost
		}
		totalSaving += i.Wastage.Rightsizing.Current.Cost - i.Wastage.Rightsizing.Recommended.Cost
		totalCurrentCost += i.Wastage.Rightsizing.Current.Cost

		m.summary.Set(itemId, ComputeInstanceSummary{
			CurrentRuntimeCost: totalCurrentCost,
			Savings:            totalSaving,
		})
	}
	m.publishResultSummary(m.ResultsSummary())
}
