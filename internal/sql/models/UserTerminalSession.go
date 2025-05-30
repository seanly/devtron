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

package models

type UserTerminalSessionRequest struct {
	Id            int          `json:"id"`
	UserId        int32        `json:"userId"`
	ClusterId     int          `json:"clusterId" validate:"number,gt=0"`
	NodeName      string       `json:"nodeName" validate:"required,min=1"`
	BaseImage     string       `json:"baseImage" validate:"required,min=1"`
	ShellName     string       `json:"shellName" validate:"required,min=1"`
	Namespace     string       `json:"namespace" validate:"required,min=1"`
	NodeTaints    []NodeTaints `json:"taints"`
	Manifest      string       `json:"manifest"`
	PodName       string       `json:"podName"`
	ContainerName string       `json:"containerName"`
	ForceDelete   bool         `json:"forceDelete"`
	DebugNode     bool         `json:"debugNode"`
}
type UserTerminalShellSessionRequest struct {
	TerminalAccessId int    `json:"terminalAccessId" validate:"number,gt=0"`
	ShellName        string `json:"shellName" validate:"required,min=1"`
	NameSpace        string `json:"namespace" validate:"required,min=1"`
	ContainerName    string `json:"containerName"`
}
type NodeTaints struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect,omitempty"`
}

type Container struct {
	ContainerName string `json:"containerName"`
	Image         string `json:"image"`
}

type UserTerminalPodEvents struct {
	Status         string      `json:"status"`
	ErrorReason    string      `json:"errorReason"`
	EventsResponse interface{} `json:"eventsResponse"`
}

type UserTerminalSessionConfig struct {
	MaxSessionPerUser                 int    `env:"MAX_SESSION_PER_USER" envDefault:"5" description:"max no of cluster terminal pods can be created by an user"`
	TerminalPodStatusSyncTimeInSecs   int    `env:"TERMINAL_POD_STATUS_SYNC_In_SECS" envDefault:"600" description:"this is the time interval at which the status of the cluster terminal pod"`
	TerminalPodDefaultNamespace       string `env:"TERMINAL_POD_DEFAULT_NAMESPACE" envDefault:"default" description:"Cluster terminal default namespace"`
	TerminalPodInActiveDurationInMins int    `env:"TERMINAL_POD_INACTIVE_DURATION_IN_MINS" envDefault:"10" description:"Timeout for cluster terminal to be inactive"`
}

type UserTerminalSessionResponse struct {
	UserTerminalSessionId string            `json:"userTerminalSessionId"`
	UserId                int32             `json:"userId"`
	TerminalAccessId      int               `json:"terminalAccessId"`
	Status                TerminalPodStatus `json:"status"`
	ErrorReason           string            `json:"errorReason"`
	PodName               string            `json:"podName"`
	NodeName              string            `json:"nodeName"`
	IsValidShell          bool              `json:"isValidShell"`
	ShellName             string            `json:"shellName"`
	Containers            []Container       `json:"containers"`
	PodExists             bool              `json:"podExists"`
	DebugNode             bool              `json:"debugNode"`
	NameSpace             string            `json:"namespace"`
}

const TerminalAccessPodNameTemplate = "terminal-access-" + TerminalAccessClusterIdTemplateVar + "-" + TerminalAccessUserIdTemplateVar + "-" + TerminalAccessRandomIdVar
const TerminalAccessClusterIdTemplateVar = "${cluster_id}"
const TerminalAccessUserIdTemplateVar = "${user_id}"
const TerminalAccessRandomIdVar = "${random_id}"
const TerminalAccessPodNameVar = "${pod_name}"
const TerminalAccessNodeNameVar = "${node_name}"
const TerminalAccessBaseImageVar = "${base_image}"
const TerminalAccessNamespaceVar = "${default_namespace}"
const TerminalAccessPodTemplateName = "terminal-access-pod"
const TerminalAccessRoleTemplateName = "terminal-access-role"
const TerminalAccessClusterRoleBindingTemplateName = "terminal-access-role-binding"
const TerminalAccessClusterRoleBindingTemplate = TerminalAccessPodNameTemplate + "-crb"
const TerminalAccessServiceAccountTemplateName = "terminal-access-service-account"
const TerminalAccessServiceAccountTemplate = TerminalAccessPodNameTemplate + "-sa"
const MaxSessionLimitReachedMsg = "session-limit-reached"
const AUTO_SELECT_NODE string = "autoSelectNode"
const ShellNotSupported string = "%s is not supported for the selected image"
const AutoSelectShell string = "*"

type TerminalPodStatus string

const (
	TerminalPodStarting   TerminalPodStatus = "Starting"
	TerminalPodRunning    TerminalPodStatus = "Running"
	TerminalPodTerminated TerminalPodStatus = "Terminated"
	TerminalPodError      TerminalPodStatus = "Error"
)
