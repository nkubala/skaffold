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

package cmd

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
)

// DefaultRepoFn takes an image tag and returns either a new tag with the default repo prefixed, or the original tag if
// no default repo is specified.
type DefaultRepoFn func(string) (string, error)

func getBuildArtifactsAndSetTags(artifacts []*latest.Artifact, defaulterFn DefaultRepoFn) ([]build.Artifact, error) {
	buildArtifacts, err := build.MergeBuildArtifacts(fromBuildOutputFile.BuildArtifacts(), preBuiltImages.Artifacts(), artifacts, opts.CustomTag)
	if err != nil {
		return nil, err
	}

	return applyDefaultRepoToArtifacts(buildArtifacts, defaulterFn)
}

func applyDefaultRepoToArtifacts(artifacts []build.Artifact, defaulterFn DefaultRepoFn) ([]build.Artifact, error) {
	for i := range artifacts {
		updatedTag, err := defaulterFn(artifacts[i].Tag)
		if err != nil {
			return nil, err
		}
		artifacts[i].Tag = updatedTag
	}

	return artifacts, nil
}
