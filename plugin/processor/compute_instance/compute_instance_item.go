package compute_instance

import (
	"cloud.google.com/go/compute/apiv1/computepb"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	golang2 "github.com/kaytu-io/plugin-gcp/plugin/proto/src/golang"
	"google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"maps"
	"strconv"
)

type ComputeInstanceItem struct {
	ProjectId           string
	Name                string
	Id                  string
	MachineType         string
	Region              string
	Platform            string
	OptimizationLoading bool
	Preferences         []*golang.PreferenceItem
	Skipped             bool
	LazyLoadingEnabled  bool
	SkipReason          string
	Instance            *computepb.Instance
	Disks               []compute.Disk
	Metrics             map[string][]*golang2.DataPoint
	DisksMetrics        map[string]map[string][]*golang2.DataPoint
	Wastage             golang2.GCPComputeOptimizationResponse
}

func (i ComputeInstanceItem) ComputeInstanceDevice() (*golang.ChartRow, map[string]*golang.Properties) {
	row := golang.ChartRow{
		RowId:  i.Id,
		Values: make(map[string]*golang.ChartRowItem),
	}
	row.RowId = i.Id

	row.Values["resource_id"] = &golang.ChartRowItem{
		Value: i.Id,
	}
	row.Values["resource_name"] = &golang.ChartRowItem{
		Value: i.Name,
	}
	row.Values["resource_type"] = &golang.ChartRowItem{
		Value: "Compute Instance",
	}

	row.Values["current_cost"] = &golang.ChartRowItem{
		Value: utils.FormatPriceFloat(i.Wastage.Rightsizing.Current.Cost),
	}

	RegionProperty := &golang.Property{
		Key:     "Region",
		Current: i.Wastage.Rightsizing.Current.Region,
	}

	MachineTypeProperty := &golang.Property{
		Key:     "Machine Type",
		Current: i.Wastage.Rightsizing.Current.MachineType,
	}
	MachineFamilyProperty := &golang.Property{
		Key:     "Machine Family",
		Current: i.Wastage.Rightsizing.Current.MachineFamily,
	}
	CPUProperty := &golang.Property{
		Key:     "  CPU",
		Current: fmt.Sprintf("%d", i.Wastage.Rightsizing.Current.Cpu),
		Average: utils.Percentage(PWrapperDouble(i.Wastage.Rightsizing.Cpu.Avg)),
		Max:     utils.Percentage(PWrapperDouble(i.Wastage.Rightsizing.Cpu.Max)),
	}

	memoryProperty := &golang.Property{
		Key:     "  MemoryMB",
		Current: fmt.Sprintf("%d MB", i.Wastage.Rightsizing.Current.MemoryMb),
		Average: utils.Percentage(PWrapperDouble(i.Wastage.Rightsizing.Memory.Avg)),
		Max:     utils.Percentage(PWrapperDouble(i.Wastage.Rightsizing.Memory.Max)),
	}

	row.Values["project_id"] = &golang.ChartRowItem{
		Value: i.ProjectId,
	}

	row.Values["current_cost"] = &golang.ChartRowItem{
		Value: utils.FormatPriceFloat(i.Wastage.Rightsizing.Current.Cost),
	}

	if i.Wastage.Rightsizing.Recommended != nil {
		row.Values["right_sized_cost"] = &golang.ChartRowItem{
			Value: utils.FormatPriceFloat(i.Wastage.Rightsizing.Recommended.Cost),
		}
		row.Values["savings"] = &golang.ChartRowItem{
			Value: utils.FormatPriceFloat(i.Wastage.Rightsizing.Current.Cost - i.Wastage.Rightsizing.Recommended.Cost),
		}
		RegionProperty.Recommended = i.Wastage.Rightsizing.Recommended.Region
		MachineTypeProperty.Recommended = i.Wastage.Rightsizing.Recommended.MachineType
		CPUProperty.Recommended = fmt.Sprintf("%d", i.Wastage.Rightsizing.Recommended.Cpu)
		memoryProperty.Recommended = fmt.Sprintf("%d MB", i.Wastage.Rightsizing.Recommended.MemoryMb)
	}

	props := make(map[string]*golang.Properties)
	properties := &golang.Properties{}

	properties.Properties = append(properties.Properties, RegionProperty)
	properties.Properties = append(properties.Properties, MachineTypeProperty)
	properties.Properties = append(properties.Properties, MachineFamilyProperty)
	properties.Properties = append(properties.Properties, &golang.Property{
		Key: "Compute",
	})
	properties.Properties = append(properties.Properties, CPUProperty)
	properties.Properties = append(properties.Properties, memoryProperty)

	props[i.Id] = properties

	return &row, props
}

func (i ComputeInstanceItem) ComputeDiskDevice() ([]*golang.ChartRow, map[string]*golang.Properties) {
	var rows []*golang.ChartRow
	props := make(map[string]*golang.Properties)

	for _, d := range i.Disks {
		key := strconv.FormatUint(d.Id, 10)
		disk := i.Wastage.VolumesRightsizing[key]

		row := golang.ChartRow{
			RowId:  key,
			Values: make(map[string]*golang.ChartRowItem),
		}
		row.RowId = key

		row.Values["project_id"] = &golang.ChartRowItem{
			Value: i.ProjectId,
		}
		row.Values["resource_id"] = &golang.ChartRowItem{
			Value: key,
		}
		row.Values["resource_name"] = &golang.ChartRowItem{
			Value: d.Name,
		}
		row.Values["resource_type"] = &golang.ChartRowItem{
			Value: "Compute Disk",
		}

		row.Values["current_cost"] = &golang.ChartRowItem{
			Value: utils.FormatPriceFloat(disk.Current.Cost),
		}

		RegionProperty := &golang.Property{
			Key:     "Region",
			Current: disk.Current.Region,
		}

		DiskTypeProperty := &golang.Property{
			Key:     "Disk Type",
			Current: disk.Current.DiskType,
		}
		DiskSizeProperty := &golang.Property{
			Key:     "Disk Size",
			Current: fmt.Sprintf("%d GB", disk.Current.DiskSize),
		}
		DiskReadIopsProperty := &golang.Property{
			Key:     "  Read IOPS Expectation",
			Current: fmt.Sprintf("%d", disk.Current.ReadIopsLimit),
			Average: utils.PFloat64ToString(PWrapperDouble(disk.ReadIops.Avg)),
			Max:     utils.PFloat64ToString(PWrapperDouble(disk.ReadIops.Max)),
		}
		DiskWriteIopsProperty := &golang.Property{
			Key:     "  Write IOPS Expectation",
			Current: fmt.Sprintf("%d", disk.Current.WriteIopsLimit),
			Average: utils.PFloat64ToString(PWrapperDouble(disk.WriteIops.Avg)),
			Max:     utils.PFloat64ToString(PWrapperDouble(disk.WriteIops.Max)),
		}
		DiskReadThroughputProperty := &golang.Property{
			Key:     "  Read Throughput Expectation",
			Current: fmt.Sprintf("%.2f Mb", disk.Current.ReadThroughputLimit),
			Average: fmt.Sprintf("%s Mb", utils.PFloat64ToString(PWrapperDouble(disk.ReadThroughput.Avg))),
			Max:     fmt.Sprintf("%s Mb", utils.PFloat64ToString(PWrapperDouble(disk.ReadThroughput.Max))),
		}
		DiskWriteThroughputProperty := &golang.Property{
			Key:     "  Write Throughput Expectation",
			Current: fmt.Sprintf("%.2f Mb", disk.Current.WriteThroughputLimit),
			Average: fmt.Sprintf("%s Mb", utils.PFloat64ToString(PWrapperDouble(disk.WriteThroughput.Avg))),
			Max:     fmt.Sprintf("%s Mb", utils.PFloat64ToString(PWrapperDouble(disk.WriteThroughput.Max))),
		}

		if disk.Recommended != nil {
			row.Values["right_sized_cost"] = &golang.ChartRowItem{
				Value: utils.FormatPriceFloat(disk.Recommended.Cost),
			}
			row.Values["savings"] = &golang.ChartRowItem{
				Value: utils.FormatPriceFloat(disk.Current.Cost - disk.Recommended.Cost),
			}
			RegionProperty.Recommended = disk.Recommended.Region
			DiskTypeProperty.Recommended = disk.Recommended.DiskType
			DiskReadIopsProperty.Recommended = fmt.Sprintf("%d", disk.Recommended.ReadIopsLimit)
			DiskWriteIopsProperty.Recommended = fmt.Sprintf("%d", disk.Recommended.WriteIopsLimit)
			DiskReadThroughputProperty.Recommended = fmt.Sprintf("%.2f Mb", disk.Recommended.ReadThroughputLimit)
			DiskWriteThroughputProperty.Recommended = fmt.Sprintf("%.2f Mb", disk.Recommended.WriteThroughputLimit)
			DiskSizeProperty.Recommended = fmt.Sprintf("%d GB", disk.Recommended.DiskSize)
		}

		properties := &golang.Properties{}

		properties.Properties = append(properties.Properties, RegionProperty)
		properties.Properties = append(properties.Properties, DiskTypeProperty)
		properties.Properties = append(properties.Properties, DiskSizeProperty)
		properties.Properties = append(properties.Properties, &golang.Property{
			Key: "IOPS",
		})
		properties.Properties = append(properties.Properties, DiskReadIopsProperty)
		properties.Properties = append(properties.Properties, DiskWriteIopsProperty)
		properties.Properties = append(properties.Properties, &golang.Property{
			Key: "Throughput",
		})
		properties.Properties = append(properties.Properties, DiskReadThroughputProperty)
		properties.Properties = append(properties.Properties, DiskWriteThroughputProperty)

		props[key] = properties
		rows = append(rows, &row)
	}

	return rows, props
}

func (i ComputeInstanceItem) Devices() ([]*golang.ChartRow, map[string]*golang.Properties) {

	var deviceRows []*golang.ChartRow
	deviceProps := make(map[string]*golang.Properties)

	instanceRows, instanceProps := i.ComputeInstanceDevice()
	diskRows, diskProps := i.ComputeDiskDevice()

	deviceRows = append(deviceRows, instanceRows)
	deviceRows = append(deviceRows, diskRows...)
	maps.Copy(deviceProps, instanceProps)
	maps.Copy(deviceProps, diskProps)

	return deviceRows, deviceProps
}

func (i ComputeInstanceItem) ToOptimizationItem() *golang.ChartOptimizationItem {

	deviceRows, deviceProps := i.Devices()

	status := ""
	if i.Skipped {
		status = fmt.Sprintf("skipped - %s", i.SkipReason)
	} else if i.LazyLoadingEnabled && !i.OptimizationLoading {
		status = "press enter to load"
	} else if i.OptimizationLoading {
		status = "loading"
	} else if i.Wastage.Rightsizing.Recommended != nil {
		totalSaving := 0.0
		totalCurrentCost := 0.0
		totalSaving += i.Wastage.Rightsizing.Current.Cost - i.Wastage.Rightsizing.Recommended.Cost
		totalCurrentCost += i.Wastage.Rightsizing.Current.Cost
		for _, d := range i.Wastage.VolumesRightsizing {
			totalSaving += d.Current.Cost - d.Recommended.Cost
			totalCurrentCost += d.Current.Cost
		}
		status = fmt.Sprintf("%s (%.2f%%)", utils.FormatPriceFloat(totalSaving), (totalSaving/totalCurrentCost)*100)
	}

	chartrow := &golang.ChartRow{
		RowId: i.Id,
		Values: map[string]*golang.ChartRowItem{
			"x_kaytu_right_arrow": {
				Value: "→",
			},
			"resource_id": {
				Value: i.Id,
			},
			"resource_name": {
				Value: i.Name,
			},
			"resource_type": {
				Value: i.MachineType,
			},
			"region": {
				Value: i.Region,
			},
			"platform": {
				Value: i.Platform,
			},
			"total_saving": {
				Value: status,
			},
		},
	}

	coi := &golang.ChartOptimizationItem{
		OverviewChartRow:   chartrow,
		DevicesChartRows:   deviceRows,
		DevicesProperties:  deviceProps,
		Preferences:        i.Preferences,
		Description:        i.Wastage.Rightsizing.Description,
		Loading:            i.OptimizationLoading,
		Skipped:            i.Skipped,
		SkipReason:         wrapperspb.String(i.SkipReason),
		LazyLoadingEnabled: i.LazyLoadingEnabled,
	}

	return coi
}

func PWrapperDouble(v *wrapperspb.DoubleValue) *float64 {
	if v == nil {
		return nil
	}
	value := v.GetValue()
	return &value
}
