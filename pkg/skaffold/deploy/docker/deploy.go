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

package docker

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/access"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/debug"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/types"
	dockerutil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/graph"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/log"
	v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	pkgsync "github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
)

type Deployer struct {
	accessor access.Accessor
	logger   log.Logger
	debugger debug.Debugger
	syncer   pkgsync.Syncer

	cfg                *v1.DockerDeploy
	client             dockerutil.LocalDaemon
	deployedContainers map[string]string                    // imageName -> containerID
	pf                 map[string][]*v1.PortForwardResource // imageName -> port forward resources
	network            string
	once               sync.Once
	debugAdapter       debug.Adapter
	tracker            *ContainerTracker
}

type Config interface {
	types.Config
}

func NewDeployer(cfg Config, labels map[string]string, d *v1.DockerDeploy, resources []*v1.PortForwardResource, provider deploy.ComponentProvider) (*Deployer, error) {
	client, err := dockerutil.NewAPIClient(cfg)
	if err != nil {
		return nil, err
	}
	pf := make(map[string][]*v1.PortForwardResource)
	for _, r := range resources {
		if r.Type == "Container" {
			pf[r.Name] = append(pf[r.Name], r)
		}
	}

	tracker := NewContainerTracker()

	return &Deployer{
		cfg:                d,
		client:             client,
		pf:                 pf,
		deployedContainers: make(map[string]string),
		network:            "skaffold-network",
		tracker:            tracker,

		debugAdapter: debug.NewAdapter(cfg.GlobalConfig(), cfg.GetInsecureRegistries()),
		accessor:     provider.Accessor.GetClusterlessAccessor(),
		logger:       provider.Logger.GetClusterlessLogger(),
		debugger:     provider.Debugger.GetClusterlessDebugger(),
		syncer:       provider.Syncer.GetClusterlessSyncer(),
	}, nil
}

func (d *Deployer) Deploy(ctx context.Context, out io.Writer, builds []graph.Artifact) ([]string, error) {
	var err error
	d.once.Do(func() {
		err = d.client.NetworkCreate(ctx, d.network)
	})
	if err != nil {
		return nil, fmt.Errorf("creating skaffold network %s: %w", d.network, err)
	}
	d.tracker.Reset() // this stops the current log streams so we can open new ones
	for _, b := range builds {
		if !util.StrSliceContains(d.cfg.Images, b.ImageName) {
			continue
		}
		if containerID, found := d.deployedContainers[b.ImageName]; found {
			logrus.Debugf("removing old container %s for image %s", containerID, b.ImageName)
			if err := d.client.Delete(ctx, out, containerID); err != nil {
				return nil, fmt.Errorf("failed to remove old container %s for image %s: %w", containerID, b.ImageName, err)
			}
		}
		container, initContainers, err := d.debugAdapter.Transform(b.Tag, b.ImageName, builds)
		if err != nil {
			return nil, errors.Wrap(err, "applying debug transforms")
		}
		id, err := d.client.Run(ctx, out, b.ImageName, b.Tag, d.network, d.pf[b.ImageName], container, initContainers)
		if err != nil {
			return nil, errors.Wrap(err, "creating container in local docker")
		}
		d.deployedContainers[b.ImageName] = id
		d.tracker.Add(b.Tag, id)
	}

	return nil, nil
}

func (d *Deployer) Dependencies() ([]string, error) {
	// noop since there is no deploy config
	// TODO(nkubala): add docker-compose.yml here?
	return nil, nil
}

func (d *Deployer) Cleanup(ctx context.Context, out io.Writer) error {
	// stop, remove, prune?
	for _, id := range d.deployedContainers {
		if err := d.client.Delete(ctx, out, id); err != nil {
			return errors.Wrap(err, "cleaning up deployed container")
		}
	}

	err := d.client.NetworkRemove(ctx, d.network)
	return errors.Wrap(err, "cleaning up skaffold created network")
}

func (d *Deployer) Render(context.Context, io.Writer, []graph.Artifact, bool, string) error {
	// TODO(nkubala): implement
	return errors.New("not implemented")
}
