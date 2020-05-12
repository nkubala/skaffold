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

package jib

import (
	"context"
	"fmt"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
)

// Build builds an artifact with Jib.
func (b *Builder) Build(ctx context.Context, out io.Writer, artifact *latest.Artifact, tags []string) (string, error) {
	t, err := DeterminePluginType(artifact.Workspace, artifact.JibArtifact)
	if err != nil {
		return "", err
	}
	var digest string

	switch t {
	case JibMaven:
		if b.pushImages {
			digest, err = b.buildJibMavenToRegistry(ctx, out, artifact.Workspace, artifact.JibArtifact, tags[0])
			if err != nil {
				return "", err
			}
			return b.tagRemote(ctx, out, digest, tags)
		}
		digest, err = b.buildJibMavenToDocker(ctx, out, artifact.Workspace, artifact.JibArtifact, tags[0])
		if err != nil {
			return "", err
		}
		return b.tagLocal(ctx, out, digest, tags)

	case JibGradle:
		if b.pushImages {
			digest, err = b.buildJibGradleToRegistry(ctx, out, artifact.Workspace, artifact.JibArtifact, tags[0])
			if err != nil {
				return "", err
			}
			return b.tagRemote(ctx, out, digest, tags)
		}
		digest, err = b.buildJibGradleToDocker(ctx, out, artifact.Workspace, artifact.JibArtifact, tags[0])
		if err != nil {
			return "", err
		}
		return b.tagLocal(ctx, out, digest, tags)

	default:
		return "", fmt.Errorf("unable to determine Jib builder type for %s", artifact.Workspace)
	}
}

func (b *Builder) tagRemote(ctx context.Context, out io.Writer, imageID string, tags []string) (string, error) {
	// TODO(nkubala): implement
	return "", nil
}

func (b *Builder) tagLocal(ctx context.Context, out io.Writer, digest string, tags []string) (string, error) {
	var err error
	for _, tag := range tags {
		if err = b.localDocker.Tag(ctx, digest, tag); err != nil {
			return "", err
		}
	}
	return "", nil
}
