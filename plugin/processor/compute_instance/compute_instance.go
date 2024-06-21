package compute_instance

import (
	"log"

	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"github.com/kaytu-io/plugin-gcp/plugin/gcp"
	util "github.com/kaytu-io/plugin-gcp/utils"
)

type ComputeInstanceProcessor struct {
	provider                *gcp.Compute
	metricProvider          *gcp.CloudMonitoring
	items                   util.ConcurrentMap[string, ComputeInstanceItem]
	publishOptimizationItem func(item *golang.ChartOptimizationItem)
	publishResultSummary    func(summary *golang.ResultSummary)
	kaytuAcccessToken       string
	jobQueue                *sdk.JobQueue
}

func NewComputeInstanceProcessor(
	prv *gcp.Compute,
	metricPrv *gcp.CloudMonitoring,
	publishOptimizationItem func(item *golang.ChartOptimizationItem),
	publishResultSummary func(summary *golang.ResultSummary),
	kaytuAcccessToken string,
	jobQueue *sdk.JobQueue,
) *ComputeInstanceProcessor {
	log.Println("creating processor")
	r := &ComputeInstanceProcessor{
		provider:                prv,
		metricProvider:          metricPrv,
		items:                   util.NewMap[string, ComputeInstanceItem](),
		publishOptimizationItem: publishOptimizationItem,
		publishResultSummary:    publishResultSummary,
		kaytuAcccessToken:       kaytuAcccessToken,
		jobQueue:                jobQueue,
		// configuration:           configurations,
		// lazyloadCounter:         lazyloadCounter,
	}

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

func (p *ComputeInstanceProcessor) ExportNonInteractive() *golang.NonInteractiveExport {
	return nil
}
