package preferences

import (
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var DefaultComputeEnginePreferences = []*golang.PreferenceItem{
	{Service: "ComputeInstance", Key: "vCPU", IsNumber: true},
	{Service: "ComputeInstance", Key: "Region", Pinned: true},
	{Service: "ComputeInstance", Key: "ExcludeCustomInstances", Value: wrapperspb.String("No"), PreventPinning: true, PossibleValues: []string{"No", "Yes"}},
	{Service: "ComputeInstance", Key: "MachineFamily", Pinned: false},
	{Service: "ComputeInstance", Key: "MemoryGB", Alias: "Memory", IsNumber: true, Unit: "GiB"},
	{Service: "ComputeInstance", Key: "CPUBreathingRoom", IsNumber: true, Value: wrapperspb.String("10"), PreventPinning: true, Unit: "%"},
	{Service: "ComputeInstance", Key: "MemoryBreathingRoom", IsNumber: true, Value: wrapperspb.String("10"), PreventPinning: true, Unit: "%"},
	{Service: "ComputeInstance", Key: "ExcludeUpsizingFeature", Value: wrapperspb.String("Yes"), PreventPinning: true, PossibleValues: []string{"No", "Yes"}},

	{Service: "ComputeDisk", Key: "DiskType"},
	{Service: "ComputeDisk", Key: "DiskSizeGb", IsNumber: true, Unit: "GiB"},
}
