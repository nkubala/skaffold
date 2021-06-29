/*
Copyright 2020 The Skaffold Authors

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

package deploy

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/access"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/debug"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/log"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/status"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync"
)

// ComponentProvider serves as a clean way to send three providers
// as params to the Deployer constructors
type ComponentProvider struct {
	Accessor access.Provider
	Debugger debug.Provider
	Logger   log.Provider
	Monitor  status.Provider
	Syncer   sync.Provider
}

// NoopComponentProvider is for tests
var NoopComponentProvider = ComponentProvider{
	Accessor: &access.NoopProvider{},
	Debugger: &debug.NoopProvider{},
	Logger:   &log.NoopProvider{},
	Monitor:  &status.NoopProvider{},
	Syncer:   &sync.NoopProvider{},
}
