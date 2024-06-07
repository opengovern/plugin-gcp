package compute_instance

import (
	"log"

	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"github.com/kaytu-io/plugin-gcp/plugin/gcp"
	util "github.com/kaytu-io/plugin-gcp/utils"
)

type ComputeInstanceProcessor struct {
	provider *gcp.Compute
	// metricProvider          *gcp.CloudMonitoring
	// identification          map[string]string
	items                   util.ConcurrentMap[string, ComputeInstanceItem]
	publishOptimizationItem func(item *golang.OptimizationItem)
	kaytuAcccessToken       string
	jobQueue                *sdk.JobQueue
}

func NewComputeInstanceProcessor(
	prv *gcp.Compute,
	// metric *gcp.CloudMonitoring,
	publishOptimizationItem func(item *golang.OptimizationItem),
	kaytuAcccessToken string,
	jobQueue *sdk.JobQueue,
) *ComputeInstanceProcessor {
	r := &ComputeInstanceProcessor{
		provider: prv,
		// metricProvider: metric,
		// identification: identification,
		items:                   util.NewMap[string, ComputeInstanceItem](),
		publishOptimizationItem: publishOptimizationItem,
		kaytuAcccessToken:       kaytuAcccessToken,
		jobQueue:                jobQueue,
		// configuration:           configurations,
		// lazyloadCounter:         lazyloadCounter,
	}
	log.Println("creating processor")

	jobQueue.Push(NewListComputeInstancesJob(r))
	return r
}

func (m *ComputeInstanceProcessor) ReEvaluate(id string, items []*golang.PreferenceItem) {
	log.Println("Reevaluate unimplemented")
	// v, _ := m.items.Get(id)
	// v.Preferences = items
	// m.items.Set(id, v)
	// m.jobQueue.Push(NewOptimizeEC2InstanceJob(m, v))
}
