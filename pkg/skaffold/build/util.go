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

package build

import (
	"context"
	"fmt"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/tag"
)

// MergeWithPreviousBuilds merges previous or prebuilt build artifacts with
// builds. If an artifact is already present in builds, the same artifact from
// previous will be replaced at the same position.
func MergeWithPreviousBuilds(builds, previous []Artifact) []Artifact {
	updatedBuilds := map[string]Artifact{}
	for _, build := range builds {
		updatedBuilds[build.ImageName] = build
	}

	added := map[string]bool{}
	var merged []Artifact

	for _, artifact := range previous {
		if updated, found := updatedBuilds[artifact.ImageName]; found {
			merged = append(merged, updated)
		} else {
			merged = append(merged, artifact)
		}
		added[artifact.ImageName] = true
	}

	for _, artifact := range builds {
		if !added[artifact.ImageName] {
			merged = append(merged, artifact)
		}
	}

	return merged
}

func TagWithDigest(tag, digest string) string {
	return tag + "@" + digest
}

func TagWithImageID(ctx context.Context, tag string, imageID string, localDocker docker.LocalDaemon) (string, error) {
	return localDocker.TagWithImageID(ctx, tag, imageID)
}

func ValidateTagsForArtifacts(artifacts []Artifact) error {
	for _, artifact := range artifacts {
		if artifact.Tag == "" {
			return fmt.Errorf("no tag provided for image [%s]", artifact.ImageName)
		}
	}
	return nil
}

func MergeBuildArtifacts(fromFile, fromCLI []Artifact, artifacts []*latest.Artifact, customTag string) ([]Artifact, error) {
	var buildArtifacts []Artifact
	for _, artifact := range artifacts {
		buildArtifacts = append(buildArtifacts, Artifact{
			ImageName: artifact.ImageName,
		})
	}

	// Tags provided by file take precedence over those provided on the command line
	buildArtifacts = MergeWithPreviousBuilds(fromCLI, buildArtifacts)
	buildArtifacts = MergeWithPreviousBuilds(fromFile, buildArtifacts)

	var err error
	if customTag != "" {
		buildArtifacts, err = applyTagToArtifacts(customTag, buildArtifacts)
	}
	if err != nil {
		return nil, err
	}

	// Check that every image has a non empty tag
	if err := ValidateTagsForArtifacts(buildArtifacts); err != nil {
		return nil, err
	}

	return buildArtifacts, nil
}

func applyTagToArtifacts(t string, artifacts []Artifact) ([]Artifact, error) {
	var result []Artifact
	for _, artifact := range artifacts {
		if artifact.Tag == "" {
			artifact.Tag = artifact.ImageName + ":" + t
		} else {
			newTag, err := tag.SetImageTag(artifact.Tag, t)
			if err != nil {
				return nil, err
			}
			artifact.Tag = newTag
		}
		result = append(result, artifact)
	}
	return result, nil
}

func applyDefaultRepoToArtifacts(artifacts []Artifact, defaultRepoFunc func(string) (string, error)) ([]Artifact, error) {
	for i := range artifacts {
		updatedTag, err := defaultRepoFunc(artifacts[i].Tag)
		if err != nil {
			return nil, err
		}
		artifacts[i].Tag = updatedTag
	}

	return artifacts, nil
}

func GetBuildArtifactsAndSetTags(fromCfg []*latest.Artifact, fromBuildOutput, fromPreBuilt []Artifact, defaulterFn func(string) (string, error), customTag string) ([]Artifact, error) {
	buildArtifacts, err := MergeBuildArtifacts(fromBuildOutput, fromPreBuilt, fromCfg, customTag)
	if err != nil {
		return nil, err
	}

	return applyDefaultRepoToArtifacts(buildArtifacts, defaulterFn)
}
