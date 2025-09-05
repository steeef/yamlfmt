package basic

import (
	"github.com/google/yamlfmt"
	"github.com/mitchellh/mapstructure"
)

type PreserveBackslashFormatterFactory struct{}

func (f *PreserveBackslashFormatterFactory) Type() string {
	return PreserveBackslashFormatterType
}

func (f *PreserveBackslashFormatterFactory) NewFormatter(configData map[string]interface{}) (yamlfmt.Formatter, error) {
	config := DefaultConfig()
	if configData != nil {
		err := mapstructure.Decode(configData, &config)
		if err != nil {
			return nil, err
		}
	}

	basicFormatter := &BasicFormatter{
		Config:       config,
		Features:     ConfigureFeaturesFromConfig(config),
		YAMLFeatures: ConfigureYAMLFeaturesFromConfig(config),
	}

	return &PreserveBackslashFormatter{
		BasicFormatter: basicFormatter,
	}, nil
}
