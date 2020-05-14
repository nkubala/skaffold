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

package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/tag"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/sirupsen/logrus"
)

func (c *cache) lookupArtifacts(ctx context.Context, tags tag.ImageTags, artifacts []*latest.Artifact) []cacheDetails {
	details := make([]cacheDetails, len(artifacts))

	var wg sync.WaitGroup
	for i := range artifacts {
		wg.Add(1)

		i := i
		go func() {
			details[i] = c.lookup(ctx, artifacts[i], tags[artifacts[i].ImageName])
			wg.Done()
		}()
	}
	wg.Wait()

	return details
}

func (c *cache) lookup(ctx context.Context, a *latest.Artifact, tags []string) cacheDetails {
	hash, err := c.hashForArtifact(ctx, a)
	if err != nil {
		return failed{err: fmt.Errorf("getting hash for artifact %q: %s", a.ImageName, err)}
	}

	entry, cacheHit := c.artifactCache[hash]
	if !cacheHit {
		return needsBuilding{hash: hash}
	}

	if c.imagesAreLocal {
		return c.lookupLocal(ctx, hash, tags, entry)
	}
	return c.lookupRemote(ctx, hash, tags, entry)
}

func (c *cache) lookupLocal(ctx context.Context, hash string, tags []string, entry ImageDetails) cacheDetails {
	if entry.ID == "" {
		return needsBuilding{hash: hash}
	}

	var idForTag string
	var err error
	for _, tag := range tags {
		// Check the imageID for the tag
		idForTag, err = c.client.ImageID(ctx, tag)
		if err != nil {
			return failed{err: fmt.Errorf("getting imageID for %s: %v", tag, err)}
		}
	}

	// Image exists locally with the same tag
	if idForTag == entry.ID {
		return found{hash: hash}
	}

	// Image exists locally with a different tag
	if c.client.ImageExists(ctx, entry.ID) {
		return needsLocalTagging{hash: hash, tags: tags, imageID: entry.ID}
	}

	return needsBuilding{hash: hash}
}

func (c *cache) lookupRemote(ctx context.Context, hash string, tags []string, entry ImageDetails) cacheDetails {
	var missingTags []string
	for _, tag := range tags {
		logrus.Infof("checking for tag %s in remote cache", tag)
		if remoteDigest, err := docker.RemoteDigest(tag, c.insecureRegistries); err == nil {
			// Image does not exist remotely with the same tag and digest
			if remoteDigest != entry.Digest {
				logrus.Infof("tag %s not found remotely", tag)
				// return found{hash: hash}
				missingTags = append(missingTags, tag)
			}
		}
	}

	// TODO1(nkubala): is this second computed digest the same as the first? if so we should just use that

	// TODO2(nkubala): also, do we need to loop through missing tags here? or just `if len(tags) > 0`
	// for _, tag := range missingTags {
	logrus.Infof("missing tags: %v", missingTags)
	if len(missingTags) > 0 {
		// Image exists remotely with a different tag
		fqn := tags[0] + "@" + entry.Digest // Actual tag will be ignored but we need the registry and the digest part of it.
		if remoteDigest, err := docker.RemoteDigest(fqn, c.insecureRegistries); err == nil {
			if remoteDigest == entry.Digest {
				return needsRemoteTagging{hash: hash, tags: tags, digest: entry.Digest}
			}
		}

		// Image exists locally
		if entry.ID != "" && c.client != nil && c.client.ImageExists(ctx, entry.ID) {
			logrus.Infof("image %s found locally, going to push", fqn)
			return needsPushing{hash: hash, tags: tags, imageID: entry.ID}
		}
	}

	return needsBuilding{hash: hash}
}
