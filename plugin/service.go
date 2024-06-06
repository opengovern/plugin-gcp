package plugin

import (
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
	stream    golang.Plugin_RegisterClient
	processor processor.PluginProcessor
}

func NewPlugin() *GCPPlugin {
	return &GCPPlugin{}
}

func (p *GCPPlugin) GetConfig() golang.RegisterConfig {
	log.Println("Get config")
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
	}
}

func (p *GCPPlugin) SetStream(stream golang.Plugin_RegisterClient) {
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

	publishOptimizationItem := func(item *golang.OptimizationItem) {
		p.stream.Send(&golang.PluginMessage{
			PluginMessage: &golang.PluginMessage_Oi{
				Oi: item,
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
	publishResultsReady(false)

	if cmd == "compute-instance" {
		p.processor = compute_instance.NewComputeInstanceProcessor(
			gcpProvider,
			publishOptimizationItem,
			kaytuAccessToken,
			jobQueue,
		)
	} else {
		return fmt.Errorf("invalid command: %s", cmd)
	}
	jobQueue.SetOnFinish(func() {
		publishResultsReady(true)
	})

	return nil
}

func (p *GCPPlugin) ReEvaluate(evaluate *golang.ReEvaluate) {
	p.processor.ReEvaluate(evaluate.Id, evaluate.Preferences)
}
