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
	commonBean "github.com/devtron-labs/common-lib/workflow"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

const IMAGE_SCANNER_ENDPOINT = "IMAGE_SCANNER_ENDPOINT"

type WorkflowPipelineType string

const (
	CI_WORKFLOW_PIPELINE_TYPE  WorkflowPipelineType = "CI"
	CD_WORKFLOW_PIPELINE_TYPE  WorkflowPipelineType = "CD"
	JOB_WORKFLOW_PIPELINE_TYPE WorkflowPipelineType = "JOB"
)

type RefPluginObject struct {
	Id    int           `json:"id"`
	Steps []*StepObject `json:"steps"`
}

type PrePostAndRefPluginStepsResponse struct {
	PreStageSteps    []*StepObject
	PostStageSteps   []*StepObject
	RefPluginData    []*RefPluginObject
	VariableSnapshot map[string]string
	PrePostAndRefPluginStepsResponseEnt
}

type StepObject struct {
	Name                     string                       `json:"name"`
	Index                    int                          `json:"index"`
	StepType                 string                       `json:"stepType"`               // REF_PLUGIN or INLINE
	ExecutorType             string                       `json:"executorType,omitempty"` //SHELL, DOCKERFILE, CONTAINER_IMAGE
	RefPluginId              int                          `json:"refPluginId,omitempty"`
	Script                   string                       `json:"script,omitempty"`
	InputVars                []*commonBean.VariableObject `json:"inputVars"`
	ExposedPorts             map[int]int                  `json:"exposedPorts"` //map of host:container
	OutputVars               []*commonBean.VariableObject `json:"outputVars"`
	TriggerSkipConditions    []*ConditionObject           `json:"triggerSkipConditions"`
	SuccessFailureConditions []*ConditionObject           `json:"successFailureConditions"`
	DockerImage              string                       `json:"dockerImage"`
	Command                  string                       `json:"command"`
	Args                     []string                     `json:"args"`
	CustomScriptMount        *MountPath                   `json:"customScriptMount"` // destination path - storeScriptAt
	SourceCodeMount          *MountPath                   `json:"sourceCodeMount"`   // destination path - mountCodeToContainerPath
	ExtraVolumeMounts        []*MountPath                 `json:"extraVolumeMounts"` // filePathMapping
	ArtifactPaths            []string                     `json:"artifactPaths"`
	TriggerIfParentStageFail bool                         `json:"triggerIfParentStageFail"`
}

type ConditionObject struct {
	ConditionType       string `json:"conditionType"`       //TRIGGER, SKIP, SUCCESS, FAIL
	ConditionOnVariable string `json:"conditionOnVariable"` //name of variable
	ConditionalOperator string `json:"conditionalOperator"`
	ConditionalValue    string `json:"conditionalValue"`
}

type MountPath struct {
	SourcePath      string `json:"sourcePath"`
	DestinationPath string `json:"destinationPath"`
}

type ContainerResources struct {
	MinCpu        string `json:"minCpu"`
	MaxCpu        string `json:"maxCpu"`
	MinStorage    string `json:"minStorage"`
	MaxStorage    string `json:"maxStorage"`
	MinEphStorage string `json:"minEphStorage"`
	MaxEphStorage string `json:"maxEphStorage"`
	MinMem        string `json:"minMem"`
	MaxMem        string `json:"maxMem"`
}
type CiProjectDetails struct {
	GitRepository   string `json:"gitRepository"`
	MaterialName    string `json:"materialName"`
	CheckoutPath    string `json:"checkoutPath"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
	CommitHash      string `json:"commitHash"`
	GitTag          string `json:"gitTag"`
	CommitTime      string `json:"commitTime"`
	//Branch        string          `json:"branch"`
	Type        string               `json:"type"`
	Message     string               `json:"message"`
	Author      string               `json:"author"`
	GitOptions  GitOptions           `json:"gitOptions"`
	SourceType  constants.SourceType `json:"sourceType"`
	SourceValue string               `json:"sourceValue"`
	WebhookData pipelineConfig.WebhookData
}
type GitOptions struct {
	UserName              string             `json:"userName"`
	Password              string             `json:"password"`
	SshPrivateKey         string             `json:"sshPrivateKey"`
	AccessToken           string             `json:"accessToken"`
	AuthMode              constants.AuthMode `json:"authMode"`
	TlsKey                string             `json:"tlsKey"`
	TlsCert               string             `json:"tlsCert"`
	CaCert                string             `json:"caCert"`
	EnableTLSVerification bool               `json:"enableTLSVerification"`
}

type NodeConstraints struct {
	ServiceAccount   string
	TaintKey         string
	TaintValue       string
	NodeLabel        map[string]string
	SkipNodeSelector bool
}

type LimitReqCpuMem struct {
	LimitCpu string
	LimitMem string
	ReqCpu   string
	ReqMem   string
}
