package simulator

import (
	"os"

	"go.yaml.in/yaml/v3"
)

/**/
type Config struct {
	ActionPerTurn        int                     `yaml:"action_per_turn"`
	TimeBetweenTurn      int                     `yaml:"time_between_turn"` // in ms
	OTLPReceiverEndpoint string                  `yaml:"otlp_receiver_endpoint"`
	Actions              map[string]ActionConfig `yaml:"actions"`
	Players              PlayersConfig           `yaml:"players"`
}

type ActionConfig struct {
	Weight int `yaml:"weight"`
}

type PlayersConfig struct {
	Count      int           `yaml:"count"`
	InitStates []StateConfig `yaml:"init_states"`
}

type StateConfig struct {
	Name   string         `yaml:"name"`
	Weight int            `yaml:"weight"`
	State  map[string]any `yaml:"state"`
}

func parseConfig(cfgPath string) (*Config, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	c := Config{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
