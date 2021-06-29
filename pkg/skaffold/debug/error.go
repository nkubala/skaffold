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

package debug

import (
	sErrors "github.com/GoogleContainerTools/skaffold/pkg/skaffold/errors"
	"github.com/GoogleContainerTools/skaffold/proto/v1"
)

// DebugHelperRetrieveErr is thrown when debug helpers could not be retrieved.
// This error occurs in skaffold debug command when transforming the manifest before deploying.
func DebugHelperRetrieveErr(err error) error {
	return sErrors.NewError(err,
		proto.ActionableErr{
			Message: err.Error(),
			ErrCode: proto.StatusCode_DEPLOY_DEBUG_HELPER_RETRIEVE_ERR,
		})
}
