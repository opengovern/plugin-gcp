package compute_instance

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/plugin-gcp/plugin/kaytu"
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
	Disks               []compute.Disk
	Metrics             map[string][]kaytu.Datapoint
	DisksMetrics        map[string]map[string][]kaytu.Datapoint
	Wastage             kaytu.GcpComputeInstanceWastageResponse
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
		Value: utils.FormatPriceFloat(i.Wastage.RightSizing.Current.Cost),
	}

	ZoneProperty := &golang.Property{
		Key:     "Zone",
		Current: i.Wastage.RightSizing.Current.Zone,
	}

	MachineTypeProperty := &golang.Property{
		Key:     "Machine Type",
		Current: i.Wastage.RightSizing.Current.MachineType,
	}
	MachineFamilyProperty := &golang.Property{
		Key:     "Machine Family",
		Current: i.Wastage.RightSizing.Current.MachineFamily,
	}
	CPUProperty := &golang.Property{
		Key:     "  CPU",
		Current: fmt.Sprintf("%d", i.Wastage.RightSizing.Current.CPU),
		Average: utils.Percentage(i.Wastage.RightSizing.CPU.Avg),
		Max:     utils.Percentage(i.Wastage.RightSizing.CPU.Max),
	}

	memoryProperty := &golang.Property{
		Key:     "  MemoryMB",
		Current: fmt.Sprintf("%d MB", i.Wastage.RightSizing.Current.MemoryMb),
		Average: utils.Percentage(i.Wastage.RightSizing.Memory.Avg),
		Max:     utils.Percentage(i.Wastage.RightSizing.Memory.Max),
	}

	row.Values["project_id"] = &golang.ChartRowItem{
		Value: i.ProjectId,
	}

	row.Values["current_cost"] = &golang.ChartRowItem{
		Value: utils.FormatPriceFloat(i.Wastage.RightSizing.Current.Cost),
	}

	if i.Wastage.RightSizing.Recommended != nil {
		row.Values["right_sized_cost"] = &golang.ChartRowItem{
			Value: utils.FormatPriceFloat(i.Wastage.RightSizing.Recommended.Cost),
		}
		row.Values["savings"] = &golang.ChartRowItem{
			Value: utils.FormatPriceFloat(i.Wastage.RightSizing.Current.Cost - i.Wastage.RightSizing.Recommended.Cost),
		}
		ZoneProperty.Recommended = i.Wastage.RightSizing.Recommended.Zone
		MachineTypeProperty.Recommended = i.Wastage.RightSizing.Recommended.MachineType
		CPUProperty.Recommended = fmt.Sprintf("%d", i.Wastage.RightSizing.Recommended.CPU)
		memoryProperty.Recommended = fmt.Sprintf("%d MB", i.Wastage.RightSizing.Recommended.MemoryMb)
	}

	props := make(map[string]*golang.Properties)
	properties := &golang.Properties{}

	properties.Properties = append(properties.Properties, ZoneProperty)
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
		disk := i.Wastage.VolumeRightSizing[key]

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
			Key:     "  Read IOPS Limit",
			Current: fmt.Sprintf("%d", disk.Current.ReadIopsLimit),
			Average: utils.PFloat64ToString(disk.ReadIops.Avg),
			Max:     utils.PFloat64ToString(disk.ReadIops.Max),
		}
		DiskWriteIopsProperty := &golang.Property{
			Key:     "  Write IOPS Limit",
			Current: fmt.Sprintf("%d", disk.Current.WriteIopsLimit),
			Average: utils.PFloat64ToString(disk.WriteIops.Avg),
			Max:     utils.PFloat64ToString(disk.WriteIops.Max),
		}
		DiskReadThroughputProperty := &golang.Property{
			Key:     "  Read Throughput Limit",
			Current: fmt.Sprintf("%d Mb", disk.Current.ReadThroughputLimit),
			Average: fmt.Sprintf("%s Mb", utils.PFloat64ToString(disk.ReadThroughput.Avg)),
			Max:     fmt.Sprintf("%s Mb", utils.PFloat64ToString(disk.ReadThroughput.Max)),
		}
		DiskWriteThroughputProperty := &golang.Property{
			Key:     "  Write Throughput Limit",
			Current: fmt.Sprintf("%d Mb", disk.Current.WriteThroughputLimit),
			Average: fmt.Sprintf("%s Mb", utils.PFloat64ToString(disk.WriteThroughput.Avg)),
			Max:     fmt.Sprintf("%s Mb", utils.PFloat64ToString(disk.WriteThroughput.Max)),
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
			DiskReadThroughputProperty.Recommended = fmt.Sprintf("%d Mb", disk.Recommended.ReadThroughputLimit)
			DiskWriteThroughputProperty.Recommended = fmt.Sprintf("%d Mb", disk.Recommended.WriteThroughputLimit)
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
	if i.Wastage.RightSizing.Recommended != nil {
		totalSaving := 0.0
		totalCurrentCost := 0.0
		totalSaving += i.Wastage.RightSizing.Current.Cost - i.Wastage.RightSizing.Recommended.Cost
		totalCurrentCost += i.Wastage.RightSizing.Current.Cost
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
		Description:        "description placeholder",
		Loading:            i.OptimizationLoading,
		Skipped:            i.Skipped,
		SkipReason:         wrapperspb.String(i.SkipReason),
		LazyLoadingEnabled: i.LazyLoadingEnabled,
	}

	return coi
}
