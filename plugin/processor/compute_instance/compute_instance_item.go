package compute_instance

import (
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
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
	Metrics             map[string][]*monitoringpb.Point
	// Wastage             kaytu.EC2InstanceWastageResponse
}

func (i ComputeInstanceItem) ToOptimizationItem() *golang.ChartOptimizationItem {

	chartrow := &golang.ChartRow{
		RowId: i.Id,
		Values: map[string]*golang.ChartRowItem{
			"instance_id": {
				Value: i.Id,
			},
			"instance_name": {
				Value: i.Name,
			},
		},
	}

	oi := &golang.ChartOptimizationItem{
		OverviewChartRow: chartrow,
		// Id:                 i.Id,
		// Name:               i.Name,
		// ResourceType:       i.MachineType,
		// Region:             i.Region,
		// Devices:            nil,
		Preferences:        i.Preferences,
		Description:        "description placeholder",
		Loading:            i.OptimizationLoading,
		Skipped:            i.Skipped,
		SkipReason:         wrapperspb.String(i.SkipReason),
		LazyLoadingEnabled: i.LazyLoadingEnabled,
	}

	// if i.Instance.PlatformDetails != nil {
	// 	oi.Platform = *i.Instance.PlatformDetails
	// }
	// if oi.Name == "" {
	// 	oi.Name = string(i.Name)
	// }

	return oi
}
