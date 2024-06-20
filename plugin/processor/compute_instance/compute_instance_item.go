package compute_instance

import (
	"fmt"
	"maps"

	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/plugin-gcp/plugin/kaytu"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ComputeInstanceItem struct {
	Name                string
	Id                  string
	MachineType         string
	Region              string
	OptimizationLoading bool
	Preferences         []*golang.PreferenceItem
	Skipped             bool
	LazyLoadingEnabled  bool
	SkipReason          string
	Metrics             map[string][]kaytu.Datapoint
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
		Current: i.Wastage.RightSizing.Current.MachineType,
	}

	memoryProperty := &golang.Property{
		Key:     "  MemoryMB",
		Current: fmt.Sprintf("%d MB", i.Wastage.RightSizing.Current.MemoryMb),
		Average: utils.Percentage(i.Wastage.RightSizing.Memory.Avg),
		Max:     utils.Percentage(i.Wastage.RightSizing.Memory.Max),
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
		memoryProperty.Recommended = fmt.Sprintf("%d GiB", i.Wastage.RightSizing.Recommended.MemoryMb)
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

func (i ComputeInstanceItem) Devices() ([]*golang.ChartRow, map[string]*golang.Properties) {

	var deviceRows []*golang.ChartRow
	deviceProps := make(map[string]*golang.Properties)

	instanceRows, instanceProps := i.ComputeInstanceDevice()

	deviceRows = append(deviceRows, instanceRows)
	maps.Copy(deviceProps, instanceProps)

	return deviceRows, deviceProps
}

func (i ComputeInstanceItem) ToOptimizationItem() *golang.ChartOptimizationItem {

	deviceRows, deviceProps := i.Devices()

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
			"Region": {
				Value: i.Region,
			},
			"Total Savings": {
				Value: i.Region,
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
