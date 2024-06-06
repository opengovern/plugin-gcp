package processor

import "github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"

type PluginProcessor interface {
	ReEvaluate(id string, items []*golang.PreferenceItem)
}
