package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/bpfoster/nutbeat/config"

	nut "github.com/robbiet480/go.nut"
)

// nutbeat configuration.
type nutbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of nutbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &nutbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts nutbeat.
func (bt *nutbeat) Run(b *beat.Beat) error {
	logp.Info("nutbeat is running! Hit CTRL-C to stop it.")

	client, connectErr := nut.Connect(bt.config.Host)
	if connectErr != nil {
		fmt.Print(connectErr)
		return connectErr
	}

	if bt.config.Authentication.Username != nil && bt.config.Authentication.Password != nil {
		var authenticationError error
		_, authenticationError = client.Authenticate(*bt.config.Authentication.Username, *bt.config.Authentication.Password)
		if authenticationError != nil {
			fmt.Print(authenticationError)
			return authenticationError
		}
	}

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}


		upsList, listErr := client.GetUPSList()
		if listErr != nil {
			fmt.Print(listErr)
			return listErr
		}

		for _, ups := range upsList {
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"type":    b.Info.Name,
					"ups": ups.Name,
					"variables": bt.TransformVariableData(ups.Variables),
				},
			}
			bt.client.Publish(event)
			logp.Info("Event sent")

		}
	}
}

func (bt *nutbeat) TransformVariableData(variables []nut.Variable) common.MapStr {
	transformedData := common.MapStr{}

	for _, variable := range variables {
		transformedData.Put(variable.Name, variable.Value)
	}

	return transformedData
}

// Stop stops nutbeat.
func (bt *nutbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
