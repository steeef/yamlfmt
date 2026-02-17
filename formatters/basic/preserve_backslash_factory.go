// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package basic

import (
	"github.com/google/yamlfmt"
	"github.com/mitchellh/mapstructure"
)

type PreserveBackslashFormatterFactory struct{}

func (f *PreserveBackslashFormatterFactory) Type() string {
	return PreserveBackslashFormatterType
}

func (f *PreserveBackslashFormatterFactory) NewFormatter(configData map[string]any) (yamlfmt.Formatter, error) {
	config := DefaultConfig()
	if configData != nil {
		err := mapstructure.Decode(configData, &config)
		if err != nil {
			return nil, err
		}
	}
	basicFormatter, err := newFormatter(config)
	if err != nil {
		return nil, err
	}
	return &PreserveBackslashFormatter{
		BasicFormatter: basicFormatter.(*BasicFormatter),
	}, nil
}
