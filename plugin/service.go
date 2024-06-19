package plugin

import (
	"context"
	"fmt"
	"log"

	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
	"github.com/kaytu-io/plugin-gcp/plugin/gcp"
	"github.com/kaytu-io/plugin-gcp/plugin/preferences"
	"github.com/kaytu-io/plugin-gcp/plugin/processor"
	"github.com/kaytu-io/plugin-gcp/plugin/processor/compute_instance"
	"github.com/kaytu-io/plugin-gcp/plugin/version"
)

type GCPPlugin struct {
	stream    *sdk.StreamController
	processor processor.PluginProcessor
}

func NewPlugin() *GCPPlugin {
	return &GCPPlugin{}
}

func (p *GCPPlugin) GetConfig() golang.RegisterConfig {
	return golang.RegisterConfig{
		Name:     "kaytu-io/plugin-gcp",
		Version:  version.VERSION,
		Provider: "gcp",
		Commands: []*golang.Command{
			{
				Name:        "compute-instance",
				Description: "Get optimization suggestions for your Compute Engine Instances",
				Flags: []*golang.Flag{
					{
						Name:        "profile",
						Default:     "",
						Description: "GCP profile for authentication",
						Required:    false,
					},
				},
				DefaultPreferences: preferences.DefaultComputeEnginePreferences,
				LoginRequired:      true,
			},
		},
		OverviewChart: &golang.ChartDefinition{

			Columns: []*golang.ChartColumnItem{
				{
					Id:       "instance_id",
					Name:     "Instance ID",
					Width:    uint32(10),
					Sortable: true,
				},
				{
					Id:       "instance_name",
					Name:     "Instance Name",
					Width:    uint32(10),
					Sortable: true,
				},
			},
		},
		DevicesChart: &golang.ChartDefinition{
			Columns: []*golang.ChartColumnItem{
				{
					Id:       "instance_name",
					Name:     "Instance Name",
					Width:    uint32(10),
					Sortable: true,
				},
				{
					Id:       "project_id",
					Name:     "Project ID",
					Width:    uint32(10),
					Sortable: true,
				},
			},
		},
	}
}

func (p *GCPPlugin) SetStream(stream *sdk.StreamController) {
	p.stream = stream
}

// StartProcess implements sdk.Processor.
func (p *GCPPlugin) StartProcess(cmd string, flags map[string]string, kaytuAccessToken string, jobQueue *sdk.JobQueue) error {

	// scope used from https://developers.google.com/identity/protocols/oauth2/scopes#compute
	gcpProvider := gcp.NewCompute(
		[]string{
			"https://www.googleapis.com/auth/compute.readonly",
		},
	)

	metricClient := gcp.NewCloudMonitoring(
		[]string{
			"https://www.googleapis.com/auth/monitoring.read",
		},
	)

	log.Println("Initializing clients")

	err := gcpProvider.InitializeClient(context.Background())
	if err != nil {
		return err
	}

	err = metricClient.InitializeClient(context.Background())
	if err != nil {
		return err
	}

	publishOptimizationItem := func(item *golang.ChartOptimizationItem) {
		p.stream.Send(&golang.PluginMessage{
			PluginMessage: &golang.PluginMessage_Coi{
				Coi: item,
			},
		})
	}

	publishResultsReady := func(b bool) {
		p.stream.Send(&golang.PluginMessage{
			PluginMessage: &golang.PluginMessage_Ready{
				Ready: &golang.ResultsReady{
					Ready: b,
				},
			},
		})
	}

	publishResultSummary := func(summary *golang.ResultSummary) {
		p.stream.Send(&golang.PluginMessage{
			PluginMessage: &golang.PluginMessage_Summary{
				Summary: summary,
			},
		})
	}

	publishResultsReady(false)

	if cmd == "compute-instance" {
		p.processor = compute_instance.NewComputeInstanceProcessor(
			gcpProvider,
			metricClient,
			publishOptimizationItem,
			publishResultSummary,
			kaytuAccessToken,
			jobQueue,
		)
	} else {
		return fmt.Errorf("invalid command: %s", cmd)
	}
	jobQueue.SetOnFinish(func(ctx context.Context) {
		publishResultsReady(true)
	})

	return nil
}

func (p *GCPPlugin) ReEvaluate(evaluate *golang.ReEvaluate) {
	p.processor.ReEvaluate(evaluate.Id, evaluate.Preferences)
}

func (p *GCPPlugin) ExportNonInteractive() *golang.NonInteractiveExport {
	return nil
}
