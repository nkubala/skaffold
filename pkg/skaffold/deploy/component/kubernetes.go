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
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/access"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/debug"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/label"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubectl"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/debugging"
	k8sloader "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/loader"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/logger"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/portforward"
	k8sstatus "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/status"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/loader"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/log"
	v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/status"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync"
)

type KubernetesConfig interface {
	portforward.Config
	k8sloader.Config
	k8sstatus.Config

	Tail() bool
	PipelineForImage(imageName string) (v1.Pipeline, bool)
	DefaultPipeline() v1.Pipeline
	Mode() config.RunMode
	GetNamespaces() []string
}

type kubernetesProvider struct {
	podSelector *kubernetes.ImageList
	cli         *kubectl.CLI

	k8sAccessor map[string]access.Accessor
	k8sMonitor  map[string]status.Monitor // keyed on KubeContext. TODO: make KubeContext a struct type.

	labeller *label.DefaultLabeller

	config KubernetesConfig
}

func NewKubernetesProvider(config KubernetesConfig, labeller *label.DefaultLabeller, podSelector *kubernetes.ImageList, cli *kubectl.CLI) Provider {
	return kubernetesProvider{
		config:      config,
		labeller:    labeller,
		podSelector: podSelector,
		cli:         cli,
		k8sAccessor: make(map[string]access.Accessor),
		k8sMonitor:  make(map[string]status.Monitor),
	}
}

func (k kubernetesProvider) Accessor() access.Accessor {
	if !k.config.PortForwardOptions().Enabled() {
		return &access.NoopAccessor{}
	}
	context := k.config.GetKubeContext()

	if k.k8sAccessor[context] == nil {
		k.k8sAccessor[context] = portforward.NewForwarderManager(kubectl.NewCLI(k.config, ""),
			k.podSelector,
			k.labeller.RunIDSelector(),
			k.config.Mode(),
			k.config.PortForwardOptions(),
			k.config.PortForwardResources())
	}
	return k.k8sAccessor[context]
}

func (k kubernetesProvider) Debugger() debug.Debugger {
	if k.config.Mode() != config.RunModes.Debug {
		return &debug.NoopDebugger{}
	}

	return debugging.NewContainerManager(k.podSelector)
}

func (k kubernetesProvider) ImageLoader() loader.ImageLoader {
	if k.config.LoadImages() {
		return k8sloader.NewImageLoader(k.config.GetKubeContext(), kubectl.NewCLI(k.config, ""))
	}
	return &loader.NoopImageLoader{}
}

func (k kubernetesProvider) Logger() log.Logger {
	return logger.NewLogAggregator(k.cli, k.podSelector, k.config)
}

func (k kubernetesProvider) Monitor() status.Monitor {
	enabled := k.config.StatusCheck()
	if enabled != nil && !*enabled { // assume disabled only if explicitly set to false
		return &status.NoopMonitor{}
	}
	context := k.config.GetKubeContext()
	if k.k8sMonitor[context] == nil {
		k.k8sMonitor[context] = k8sstatus.NewStatusMonitor(k.config, k.labeller)
	}
	return k.k8sMonitor[context]
}

func (k kubernetesProvider) Syncer() sync.Syncer {
	return sync.NewPodSyncer(k.cli, k.config)
}
