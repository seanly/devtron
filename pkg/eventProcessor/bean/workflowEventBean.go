/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

import (
	"context"
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/common-lib/utils/registry"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util"
	"time"
)

type CdStageCompleteEvent struct {
	CiProjectDetails              []bean3.CiProjectDetails     `json:"ciProjectDetails"`
	WorkflowId                    int                          `json:"workflowId"`
	WorkflowRunnerId              int                          `json:"workflowRunnerId"`
	CdPipelineId                  int                          `json:"cdPipelineId"`
	TriggeredBy                   int32                        `json:"triggeredBy"`
	StageYaml                     string                       `json:"stageYaml"`
	ArtifactLocation              string                       `json:"artifactLocation"`
	PipelineName                  string                       `json:"pipelineName"`
	CiArtifactDTO                 pipelineConfig.CiArtifactDTO `json:"ciArtifactDTO"`
	PluginRegistryArtifactDetails map[string][]string          `json:"PluginRegistryArtifactDetails"`
	PluginArtifacts               *PluginArtifacts             `json:"pluginArtifacts"`
	IsArtifactUploaded            bool                         `json:"isArtifactUploaded"`
	IsFailed                      bool                         `json:"isFailed"`
}

type UserDeploymentRequest struct {
	Id                    int                         `json:"id"`
	ValuesOverrideRequest *bean.ValuesOverrideRequest `json:"valuesOverrideRequest"` // Internal field - will be extracted from UserDeploymentRequest, handled for backward compatibility
	TriggeredAt           time.Time                   `json:"triggeredAt"`           // Internal field - will be extracted from UserDeploymentRequest, handled for backward compatibility
	TriggeredBy           int32                       `json:"triggeredBy"`           // Internal field - will be extracted from UserDeploymentRequest, handled for backward compatibility
}

func (r *UserDeploymentRequest) WithCdWorkflowRunnerId(id int) *UserDeploymentRequest {
	if r.ValuesOverrideRequest == nil {
		return r
	}
	r.ValuesOverrideRequest.WfrId = id
	return r
}

func (r *UserDeploymentRequest) WithPipelineOverrideId(id int) *UserDeploymentRequest {
	if r.ValuesOverrideRequest == nil {
		return r
	}
	r.ValuesOverrideRequest.PipelineOverrideId = id
	return r
}

type CiCompleteEvent struct {
	CiProjectDetails              []bean3.CiProjectDetails `json:"ciProjectDetails"`
	DockerImage                   string                   `json:"dockerImage" validate:"required,image-validator"`
	Digest                        string                   `json:"digest"`
	PipelineId                    int                      `json:"pipelineId"`
	WorkflowId                    *int                     `json:"workflowId"`
	TriggeredBy                   int32                    `json:"triggeredBy"`
	PipelineName                  string                   `json:"pipelineName"`
	DataSource                    string                   `json:"dataSource"`
	MaterialType                  string                   `json:"materialType"`
	Metrics                       util.CIMetrics           `json:"metrics"`
	AppName                       string                   `json:"appName"`
	IsArtifactUploaded            bool                     `json:"isArtifactUploaded"`
	FailureReason                 string                   `json:"failureReason"` // FailureReason is used for notifying the failure reason to the user. Should be short and user-friendly
	ImageDetailsFromCR            json.RawMessage          `json:"imageDetailsFromCR"`
	PluginRegistryArtifactDetails map[string][]string      `json:"PluginRegistryArtifactDetails"`
	PluginArtifactStage           string                   `json:"pluginArtifactStage"`
	IsScanEnabled                 bool                     `json:"isScanEnabled"`
	TargetPlatforms               []string                 `json:"targetPlatforms"`
	pluginImageDetails            *registry.ImageDetailsFromCR
	PluginArtifacts               *PluginArtifacts `json:"pluginArtifacts"`
}

func (c *CiCompleteEvent) GetPluginImageDetails() *registry.ImageDetailsFromCR {
	if c == nil {
		return nil
	}
	return c.pluginImageDetails
}

func (c *CiCompleteEvent) SetImageDetailsFromCR() error {
	if c.ImageDetailsFromCR == nil {
		return nil
	}
	var imageDetailsFromCR *registry.ImageDetailsFromCR
	err := json.Unmarshal(c.ImageDetailsFromCR, &imageDetailsFromCR)
	if err != nil {
		return err
	}
	c.pluginImageDetails = imageDetailsFromCR
	return nil
}

type DevtronAppReleaseContextType struct {
	CancelParentContext context.CancelFunc
	CancelContext       context.CancelCauseFunc
	RunnerId            int
}

type CiCdStatus struct {
	DevtronOwnerInstance string `json:"devtronOwnerInstance"`
	*v1alpha1.WorkflowStatus
}

func NewCiCdStatus() CiCdStatus {
	return CiCdStatus{
		WorkflowStatus: &v1alpha1.WorkflowStatus{},
	}
}
