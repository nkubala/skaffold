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

package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/access"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/debug"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/label"
	dockerutil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker/logger"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker/tracker"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/graph"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/log"
	v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/status"
	pkgsync "github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
)

type Deployer struct {
	accessor access.Accessor
	debugger debug.Debugger
	logger   log.Logger
	monitor  status.Monitor
	syncer   pkgsync.Syncer

	cfg                *v1.DockerDeploy
	tracker            *tracker.ContainerTracker
	client             dockerutil.LocalDaemon
	deployedContainers map[string]string // imageName -> containerID
	network            string
	once               sync.Once
}

func NewDeployer(cfg dockerutil.Config, labeller *label.DefaultLabeller, d *v1.DockerDeploy, resources []*v1.PortForwardResource) (*Deployer, error) {
	client, err := dockerutil.NewAPIClient(cfg)
	if err != nil {
		return nil, err
	}

	tracker := tracker.NewContainerTracker()
	l, err := logger.NewLogger(tracker, cfg)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stdout, "retrieved logger: %+v\n", l)

	return &Deployer{
		cfg:                d,
		client:             client,
		deployedContainers: make(map[string]string),
		network:            fmt.Sprintf("skaffold-network-%s", uuid.New().String()),
		// TODO(nkubala): implement components
		tracker:  tracker,
		accessor: &access.NoopAccessor{},
		debugger: &debug.NoopDebugger{},
		logger:   l,
		monitor:  &status.NoopMonitor{},
		syncer:   &pkgsync.NoopSyncer{},
	}, nil
}

func (d *Deployer) TrackBuildArtifacts(artifacts []graph.Artifact) {
	for _, artifact := range artifacts {
		d.tracker.Add(artifact.ImageName, artifact.Tag)
	}
}

func (d *Deployer) Deploy(ctx context.Context, out io.Writer, builds []graph.Artifact) error {
	var err error
	d.once.Do(func() {
		err = d.client.NetworkCreate(ctx, d.network)
	})
	if err != nil {
		return fmt.Errorf("creating skaffold network %s: %w", d.network, err)
	}
	for _, b := range builds {
		// TODO(nkubala): parallelize this
		if !util.StrSliceContains(d.cfg.Images, b.ImageName) {
			continue
		}
		if containerID, found := d.deployedContainers[b.ImageName]; found {
			logrus.Debugf("removing old container %s for image %s", containerID, b.ImageName)
			if err := d.client.Delete(ctx, out, containerID); err != nil {
				return fmt.Errorf("failed to remove old container %s for image %s: %w", containerID, b.ImageName, err)
			}
		}
		if d.cfg.UseCompose {
			// TODO(nkubala): implement
			return fmt.Errorf("docker compose not yet supported by skaffold")
		}
		opts := dockerutil.ContainerCreateOpts{
			Name:    b.ImageName,
			Image:   b.Tag,
			Network: d.network,
		}
		id, err := d.client.Run(ctx, out, opts)
		if err != nil {
			return errors.Wrap(err, "creating container in local docker")
		}
		d.deployedContainers[b.ImageName] = id
		d.tracker.Add(b.Tag, id)
	}

	return nil
}

func (d *Deployer) Dependencies() ([]string, error) {
	// noop since there is no deploy config
	return nil, nil
}

func (d *Deployer) Cleanup(ctx context.Context, out io.Writer) error {
	for _, id := range d.deployedContainers {
		if err := d.client.Delete(ctx, out, id); err != nil {
			// TODO(nkubala): replace with actionable error
			return errors.Wrap(err, "cleaning up deployed container")
		}
	}

	err := d.client.NetworkRemove(ctx, d.network)
	return errors.Wrap(err, "cleaning up skaffold created network")
}

func (d *Deployer) Render(context.Context, io.Writer, []graph.Artifact, bool, string) error {
	return errors.New("render not implemented for docker deployer")
}

func (d *Deployer) GetAccessor() access.Accessor {
	return d.accessor
}

func (d *Deployer) GetDebugger() debug.Debugger {
	return d.debugger
}

func (d *Deployer) GetLogger() log.Logger {
	return d.logger
}

func (d *Deployer) GetSyncer() pkgsync.Syncer {
	return d.syncer
}

func (d *Deployer) GetStatusMonitor() status.Monitor {
	return d.monitor
}

func (d *Deployer) RegisterLocalImages([]graph.Artifact) {
	// all images are local, so this is a noop
}
