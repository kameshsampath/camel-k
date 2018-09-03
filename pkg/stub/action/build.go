/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"context"
	"github.com/apache/camel-k/pkg/build"
	"github.com/sirupsen/logrus"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/apache/camel-k/pkg/build/api"
)

type BuildAction struct {
	buildManager	*build.BuildManager
}

func NewBuildAction(ctx context.Context, namespace string) *BuildAction {
	return &BuildAction{
		buildManager: build.NewBuildManager(ctx, namespace),
	}
}

func (b *BuildAction) CanHandle(integration *v1alpha1.Integration) bool {
	return integration.Status.Phase == v1alpha1.IntegrationPhaseBuilding
}

func (b *BuildAction) Handle(integration *v1alpha1.Integration) error {

	buildResult := b.buildManager.Get(integration.Status.Identifier)
	if buildResult.Status == api.BuildStatusNotRequested {
		b.buildManager.Start(api.BuildSource{
			Identifier: integration.Status.Identifier,
			Code: *integration.Spec.Source.Code, // FIXME possible panic
		})
		logrus.Info("Build started")
	} else if buildResult.Status == api.BuildStatusError {
		target := integration.DeepCopy()
		target.Status.Phase = v1alpha1.IntegrationPhaseError
		return sdk.Update(target)
	} else if buildResult.Status == api.BuildStatusCompleted {
		target := integration.DeepCopy()
		target.Status.Phase = v1alpha1.IntegrationPhaseDeploying
		return sdk.Update(target)
	}

	return nil
}