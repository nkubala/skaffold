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

package sync

import (
	gosync "sync"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubectl"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
)

type Provider interface {
	GetClusterlessSyncer() Syncer
	GetKubernetesSyncer(*kubernetes.ImageList) Syncer
	GetNoopSyncer() Syncer
}

type fullProvider struct {
	clusterlessSyncer func() Syncer
	kubernetesSyncer  func(*kubernetes.ImageList) Syncer
	noopSyncer        func() Syncer
}

var (
	provider *fullProvider
	once     gosync.Once
)

func NewSyncProvider(config Config, cli *kubectl.CLI) Provider {
	once.Do(func() {
		provider = &fullProvider{
			clusterlessSyncer: func() Syncer {
				return &containerSyncer{}
			},
			kubernetesSyncer: func(podSelector *kubernetes.ImageList) Syncer {
				return &podSyncer{
					kubectl:    cli,
					namespaces: config.GetNamespaces(),
				}
			},
			noopSyncer: func() Syncer {
				return nil
			},
		}
	})
	return provider
}

func (p *fullProvider) GetClusterlessSyncer() Syncer {
	return p.clusterlessSyncer()
}

func (p *fullProvider) GetKubernetesSyncer(s *kubernetes.ImageList) Syncer {
	return p.kubernetesSyncer(s)
}

func (p *fullProvider) GetNoopSyncer() Syncer {
	return p.noopSyncer()
}

type NoopProvider struct{}

func (p *NoopProvider) GetClusterlessSyncer() Syncer {
	return &NoopSyncer{}
}

func (p *NoopProvider) GetKubernetesSyncer(_ *kubernetes.ImageList) Syncer {
	return &NoopSyncer{}
}

func (p *NoopProvider) GetNoopSyncer() Syncer {
	return &NoopSyncer{}
}
