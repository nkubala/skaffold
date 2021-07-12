/*
Copyright 2021 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package component

import (
	"reflect"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/access"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/label"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/portforward"
	v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

type mockConfig struct {
	portforward.Config
	opts config.PortForwardOptions

	statusCheck *bool
	runMode     config.RunMode
}

func (m mockConfig) Mode() config.RunMode                            { return m.runMode }
func (m mockConfig) PortForwardOptions() config.PortForwardOptions   { return m.opts }
func (m mockConfig) PortForwardResources() []*v1.PortForwardResource { return nil }
func (m mockConfig) GetKubeContext() string                          { return "" }
func (m mockConfig) GetKubeNamespace() string                        { return "" }
func (m mockConfig) GetKubeConfig() string                           { return "" }
func (m mockConfig) DefaultPipeline() v1.Pipeline                    { return v1.Pipeline{} }
func (m mockConfig) LoadImages() bool                                { return true }
func (m mockConfig) Muted() config.Muted                             { return config.Muted{} }
func (m mockConfig) PipelineForImage(string) (v1.Pipeline, bool)     { return v1.Pipeline{}, true }
func (m mockConfig) StatusCheck() *bool                              { return m.statusCheck }
func (m mockConfig) GetNamespaces() []string                         { return nil }
func (m mockConfig) StatusCheckDeadlineSeconds() int                 { return 1 }
func (m mockConfig) Tail() bool                                      { return true }

func TestGetAccessor(t *testing.T) {
	tests := []struct {
		description string
		enabled     bool
		isNoop      bool
	}{
		{
			description: "unspecified parameter defaults to disabled",
			isNoop:      true,
		},
		{
			description: "portForwardEnabled parameter set to true",
			enabled:     true,
		},
		{
			description: "portForwardEnabled parameter set to false",
			isNoop:      true,
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			opts := config.PortForwardOptions{}
			if test.enabled {
				opts.Append("1") // default enabled mode
			}
			m := NewKubernetesProvider(mockConfig{opts: opts}, label.NewLabeller(false, nil, ""), nil, nil).Accessor()
			t.CheckDeepEqual(test.isNoop, reflect.Indirect(reflect.ValueOf(m)).Type() == reflect.TypeOf(access.NoopAccessor{}))
		})
	}
}
