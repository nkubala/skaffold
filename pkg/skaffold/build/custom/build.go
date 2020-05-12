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

package custom

import (
	"context"
	"fmt"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
)

// Build builds an artifact using a custom script
func (b *Builder) Build(ctx context.Context, out io.Writer, artifact *latest.Artifact, tags []string) (string, error) {
	if err := b.runBuildScript(ctx, out, artifact, tags[0]); err != nil {
		return "", fmt.Errorf("building custom artifact: %w", err)
	}

	if b.pushImages {
		digest, err := docker.RemoteDigest(tags[0], b.insecureRegistries)
		if err != nil {
			return "", err
		}
		for _, tag := range tags {
			if err := docker.AddRemoteTag(digest, tag, b.insecureRegistries); err != nil {
				return "", fmt.Errorf("adding tag to remote image: %s", err.Error())
			}
		}
		return digest, nil
	}

	// look for any of the tags provided in the local daemon.
	// if none are found, the script didn't fulfill its contract, so error.
	// otherwise, for the ones that weren't found, create the tags.
	foundTags := map[string]bool{}
	var (
		foundOne bool
		imageID  string
		err      error
	)
	for _, tag := range tags {
		imageID, err = b.localDocker.ImageID(ctx, tag)
		if err != nil {
			return "", err
		}
		if imageID != "" {
			foundTags[tag] = true
			foundOne = true
		}
	}

	if !foundOne {
		return "", fmt.Errorf("the custom script didn't produce an image with any of the provided tags: %+v", tags)
	}

	for tag, ok := range foundTags {
		if !ok {
			// tag image
			if err := b.localDocker.Tag(ctx, imageID, tag); err != nil {
				return "", fmt.Errorf("adding tag to image: %s", err.Error())
			}
		}
	}

	return imageID, nil
}
