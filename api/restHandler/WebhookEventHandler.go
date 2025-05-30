/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package restHandler

import (
	bean3 "github.com/devtron-labs/devtron/pkg/build/git/gitHost/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost/read"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type WebhookEventHandler interface {
	OnWebhookEvent(w http.ResponseWriter, r *http.Request)
}

type WebhookEventHandlerImpl struct {
	logger                        *zap.SugaredLogger
	gitHostReadService            read.GitHostReadService
	eventClient                   client.EventClient
	webhookSecretValidator        gitWebhook.WebhookSecretValidator
	webhookEventDataConfig        pipeline.WebhookEventDataConfig
	ciPipelineEventPublishService out.CIPipelineEventPublishService
}

func NewWebhookEventHandlerImpl(logger *zap.SugaredLogger, eventClient client.EventClient,
	webhookSecretValidator gitWebhook.WebhookSecretValidator, webhookEventDataConfig pipeline.WebhookEventDataConfig,
	ciPipelineEventPublishService out.CIPipelineEventPublishService,
	gitHostReadService read.GitHostReadService) *WebhookEventHandlerImpl {
	return &WebhookEventHandlerImpl{
		logger:                        logger,
		eventClient:                   eventClient,
		webhookSecretValidator:        webhookSecretValidator,
		webhookEventDataConfig:        webhookEventDataConfig,
		ciPipelineEventPublishService: ciPipelineEventPublishService,
		gitHostReadService:            gitHostReadService,
	}
}

func (impl WebhookEventHandlerImpl) OnWebhookEvent(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("webhook event came")

	// get git host Id and secret from request
	vars := mux.Vars(r)
	var gitHostId int
	var err error
	var gitHostName string
	var gitHost *bean3.GitHostRequest
	if gitHostId, err = strconv.Atoi(vars["gitHostId"]); err != nil {
		gitHostName = vars["gitHostId"]
		// get git host from DB
		gitHost, err = impl.gitHostReadService.GetByName(gitHostName)
		if err != nil {
			impl.logger.Errorw("Error in getting git host from DB by Name", "err", err, "gitHostName", gitHostName)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		gitHostId = gitHost.Id

	} else {
		// get git host from DB
		gitHost, err = impl.gitHostReadService.GetById(gitHostId)
		if err != nil {
			impl.logger.Errorw("Error in getting git host from DB by Id", "err", err, "gitHostId", gitHostId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}

	secretFromRequest := vars["secret"]
	impl.logger.Debugw("webhook event request data", "gitHostIdentifier", vars["gitHostId"], "secretFromRequest", secretFromRequest)

	// validate signature
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		impl.logger.Errorw("Cannot read the request body:", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	isValidSig := impl.webhookSecretValidator.ValidateSecret(r, secretFromRequest, requestBodyBytes, gitHost)
	impl.logger.Debug("Secret validation result: " + strconv.FormatBool(isValidSig))
	if !isValidSig {
		impl.logger.Error("Signature mismatch")
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	// validate event type if configured
	var eventType string
	if len(gitHost.EventTypeHeader) > 0 {
		eventType = r.Header.Get(gitHost.EventTypeHeader)
		impl.logger.Debug("eventType: " + eventType)
		if len(eventType) == 0 {
			impl.logger.Errorw("Event type not known ", "eventType", eventType)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	// make request to handle this webhook
	webhookEvent := &bean.CIPipelineGitWebhookEvent{
		GitHostId:          gitHostId,
		GitHostName:        gitHostName,
		EventType:          eventType,
		RequestPayloadJson: string(requestBodyBytes),
	}

	// save in DB
	err = impl.webhookEventDataConfig.Save(webhookEvent)
	if err != nil {
		impl.logger.Errorw("Error while saving webhook data", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// write event
	err = impl.ciPipelineEventPublishService.PublishGitWebhookEvent(webhookEvent)
	if err != nil {
		impl.logger.Errorw("Error while handling webhook in git-sensor", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
}
