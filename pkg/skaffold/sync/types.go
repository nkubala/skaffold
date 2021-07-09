/*
Copyright 2019 The Skaffold Authors

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

package sync

import (
	"context"

	pkgkubectl "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubectl"
)

type syncMap map[string][]string

type Item struct {
	Image  string
	Copy   map[string][]string
	Delete map[string][]string
}

type Syncer interface {
	Sync(context.Context, *Item) error
}

type PodSyncer struct {
	kubectl *pkgkubectl.CLI
	config  Config
}

func NewPodSyncer(cli *pkgkubectl.CLI, config Config) *PodSyncer {
	return &PodSyncer{
		kubectl: cli,
		config:  config,
	}
}

type Config interface {
	GetNamespaces() []string
}

type NoopSyncer struct{}

func (s *NoopSyncer) Sync(context.Context, *Item) error {
	return nil
}
