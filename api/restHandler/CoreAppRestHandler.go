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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	app2 "github.com/devtron-labs/devtron/api/restHandler/app/pipeline/configure"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	appWorkflowBean "github.com/devtron-labs/devtron/pkg/appWorkflow/bean"
	read2 "github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/read"
	repository3 "github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/read"
	pipelineBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	common2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	bean3 "github.com/devtron-labs/devtron/pkg/chart/bean"
	read5 "github.com/devtron-labs/devtron/pkg/chart/read"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	read3 "github.com/devtron-labs/devtron/pkg/team/read"
	"net/http"
	"strconv"
	"strings"
	"time"

	appBean "github.com/devtron-labs/devtron/api/appbean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/models"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/strings/slices"
)

const (
	APP_DELETE_FAILED_RESP              = "App deletion failed, please try deleting from Devtron UI"
	APP_CREATE_SUCCESSFUL_RESP          = "App created successfully."
	APP_WORKFLOW_CREATE_SUCCESSFUL_RESP = "App workflow created successfully."
)

type CoreAppRestHandler interface {
	GetAppAllDetail(w http.ResponseWriter, r *http.Request)
	CreateApp(w http.ResponseWriter, r *http.Request)
	CreateAppWorkflow(w http.ResponseWriter, r *http.Request)
	GetAppWorkflow(w http.ResponseWriter, r *http.Request)
	GetAppWorkflowAndOverridesSample(w http.ResponseWriter, r *http.Request)
}

type CoreAppRestHandlerImpl struct {
	logger                  *zap.SugaredLogger
	userAuthService         user.UserService
	validator               *validator.Validate
	enforcerUtil            rbac.EnforcerUtil
	enforcer                casbin.Enforcer
	appCrudOperationService app.AppCrudOperationService
	pipelineBuilder         pipeline.PipelineBuilder
	gitRegistryService      gitProvider.GitRegistryConfig
	gitProviderReadService  read.GitProviderReadService
	chartService            chart.ChartService
	configMapService        pipeline.ConfigMapService
	appListingService       app.AppListingService
	propertiesConfigService pipeline.PropertiesConfigService
	appWorkflowService      appWorkflow.AppWorkflowService
	gitMaterialReadService  read2.GitMaterialReadService
	appWorkflowRepository   appWorkflow2.AppWorkflowRepository
	environmentRepository   repository.EnvironmentRepository
	configMapRepository     chartConfig.ConfigMapRepository
	chartRepo               chartRepoRepository.ChartRepository
	pipelineStageService    pipeline.PipelineStageService
	ciPipelineRepository    pipelineConfig.CiPipelineRepository
	teamReadService         read3.TeamReadService
	chartReadService        read5.ChartReadService
}

func NewCoreAppRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate, enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer, appCrudOperationService app.AppCrudOperationService, pipelineBuilder pipeline.PipelineBuilder, gitRegistryService gitProvider.GitRegistryConfig,
	chartService chart.ChartService, configMapService pipeline.ConfigMapService, appListingService app.AppListingService,
	propertiesConfigService pipeline.PropertiesConfigService, appWorkflowService appWorkflow.AppWorkflowService,
	appWorkflowRepository appWorkflow2.AppWorkflowRepository, environmentRepository repository.EnvironmentRepository, configMapRepository chartConfig.ConfigMapRepository,
	chartRepo chartRepoRepository.ChartRepository, teamService team.TeamService,
	pipelineStageService pipeline.PipelineStageService, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	gitProviderReadService read.GitProviderReadService,
	gitMaterialReadService read2.GitMaterialReadService,
	teamReadService read3.TeamReadService,
	chartReadService read5.ChartReadService) *CoreAppRestHandlerImpl {
	handler := &CoreAppRestHandlerImpl{
		logger:                  logger,
		userAuthService:         userAuthService,
		validator:               validator,
		enforcerUtil:            enforcerUtil,
		enforcer:                enforcer,
		appCrudOperationService: appCrudOperationService,
		pipelineBuilder:         pipelineBuilder,
		gitRegistryService:      gitRegistryService,
		gitProviderReadService:  gitProviderReadService,
		chartService:            chartService,
		configMapService:        configMapService,
		appListingService:       appListingService,
		propertiesConfigService: propertiesConfigService,
		appWorkflowService:      appWorkflowService,
		gitMaterialReadService:  gitMaterialReadService,
		appWorkflowRepository:   appWorkflowRepository,
		environmentRepository:   environmentRepository,
		configMapRepository:     configMapRepository,
		chartRepo:               chartRepo,
		pipelineStageService:    pipelineStageService,
		ciPipelineRepository:    ciPipelineRepository,
		teamReadService:         teamReadService,
		chartReadService:        chartReadService,
	}
	return handler
}

func (handler CoreAppRestHandlerImpl) GetAppAllDetail(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppAllDetail", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac implementation for app (user should be admin)
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object); !ok {
		handler.logger.Errorw("Unauthorized User for app update action", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac implementation ends here for app

	handler.logger.Debugw("Getting app detail v2", "appId", appId)

	//get/build app metadata starts
	appMetadataResp, err, statusCode := handler.buildAppMetadata(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build app metadata ends

	//get/build git materials starts
	gitMaterialsResp, err, statusCode := handler.buildAppGitMaterials(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build git materials ends

	//get/build docker config starts
	dockerConfig, err, statusCode := handler.buildDockerConfig(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build docker config ends

	//get/build global deployment template starts
	globalDeploymentTemplateResp, err, statusCode := handler.buildAppDeploymentTemplate(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build global deployment template ends

	//get/build app workflows starts
	//using empty workflow name because it is optional, if not provided then workflows will be fetched on the basis of app
	wfCloneRequest := &appWorkflowBean.WorkflowCloneRequest{AppId: appId}
	appWorkflows, err, statusCode := handler.buildAppWorkflows(wfCloneRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build app workflows ends

	//get/build global config maps starts
	globalConfigMapsResp, err, statusCode := handler.buildAppGlobalConfigMaps(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build global config maps ends

	//get/build global secrets starts
	globalSecretsResp, err, statusCode := handler.buildAppGlobalSecrets(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build global secrets ends

	//get/build environment override starts
	environmentOverrides, err, statusCode := handler.buildEnvironmentOverrides(r.Context(), appId, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build environment override ends

	//build full object for response
	appDetail := &appBean.AppDetail{
		Metadata:                 appMetadataResp,
		GitMaterials:             gitMaterialsResp,
		DockerConfig:             dockerConfig,
		GlobalDeploymentTemplate: globalDeploymentTemplateResp,
		AppWorkflows:             appWorkflows,
		GlobalConfigMaps:         globalConfigMapsResp,
		GlobalSecrets:            globalSecretsResp,
		EnvironmentOverrides:     environmentOverrides,
	}
	//end

	common.WriteJsonResp(w, nil, appDetail, http.StatusOK)
}

func (handler CoreAppRestHandlerImpl) CreateApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")

	ctx := r.Context()
	var createAppRequest appBean.AppDetail
	err = decoder.Decode(&createAppRequest)
	if err != nil {
		handler.logger.Errorw("request err, CreateApp by API", "err", err, "CreateApp", createAppRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//to add more validations here
	handler.logger.Infow("request payload, CreateApp by API", "CreateApp", createAppRequest)
	err = handler.validator.Struct(createAppRequest)
	if err != nil {
		handler.logger.Errorw("validation err, CreateApp by API", "err", err, "CreateApp", createAppRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac starts
	team, err := handler.teamReadService.FindByTeamName(createAppRequest.Metadata.ProjectName)
	if err != nil {
		handler.logger.Errorw("Error in getting team", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if team == nil {
		handler.logger.Errorw("no project found by name in CreateApp request by API")
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// with admin roles, you have to access for all the apps of the project to create new app. (admin or manager with specific app permission can't create app.)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, fmt.Sprintf("%s/%s", team.Name, "*")); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac ends

	handler.logger.Infow("creating app v2", "createAppRequest", createAppRequest)

	// validate payload starts
	createAppWorkflowReq := appBean.AppWorkflowCloneDto{
		AppName:              createAppRequest.Metadata.AppName,
		AppWorkflows:         createAppRequest.AppWorkflows,
		EnvironmentOverrides: createAppRequest.EnvironmentOverrides,
	}
	err, statusCode := handler.ValidateAppWorkflowRequest(&createAppWorkflowReq, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	// validate payload ends

	//creating blank app starts
	createBlankAppResp, err, statusCode := handler.createBlankApp(createAppRequest.Metadata, userId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//creating blank app ends

	//declaring appId for creating other components of app
	appId := createBlankAppResp.Id

	var errResp *multierror.Error

	//creating git material starts
	if createAppRequest.GitMaterials != nil {
		err, statusCode = handler.createGitMaterials(appId, createAppRequest.GitMaterials, userId)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating git material ends

	//creating docker config
	if createAppRequest.DockerConfig != nil {
		err, statusCode = handler.createDockerConfig(appId, createAppRequest.DockerConfig, userId)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating docker config ends

	//creating deployment template starts
	if createAppRequest.GlobalDeploymentTemplate != nil {
		err, statusCode = handler.createDeploymentTemplate(ctx, appId, createAppRequest.GlobalDeploymentTemplate, userId)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating deployment template ends

	//creating global configMaps starts
	if createAppRequest.GlobalConfigMaps != nil {
		err, statusCode = handler.createGlobalConfigMaps(appId, userId, createAppRequest.GlobalConfigMaps)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating global configMaps ends

	//creating global secrets starts
	if createAppRequest.GlobalSecrets != nil {
		err, statusCode = handler.createGlobalSecrets(appId, userId, createAppRequest.GlobalSecrets)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating global secrets ends

	//creating workflow starts
	if createAppRequest.AppWorkflows != nil {
		err, statusCode = handler.createWorkflows(ctx, appId, userId, createAppRequest.AppWorkflows)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating workflow ends

	//creating environment override starts
	if createAppRequest.EnvironmentOverrides != nil {
		err, statusCode = handler.createEnvOverrides(ctx, appId, userId, createAppRequest.EnvironmentOverrides)
		if err != nil {
			errResp = multierror.Append(errResp, err)
			errInAppDelete := handler.deleteApp(ctx, appId, userId)
			if errInAppDelete != nil {
				errResp = multierror.Append(errResp, fmt.Errorf("%s : %w", APP_DELETE_FAILED_RESP, errInAppDelete))
			}
			common.WriteJsonResp(w, errResp, nil, statusCode)
			return
		}
	}
	//creating environment override ends

	common.WriteJsonResp(w, nil, APP_CREATE_SUCCESSFUL_RESP, http.StatusOK)
}

//GetApp related methods starts

// get/build app metadata
func (handler CoreAppRestHandlerImpl) buildAppMetadata(appId int) (*appBean.AppMetadata, error, int) {
	handler.logger.Debugw("Getting app detail - meta data", "appId", appId)

	appMetaInfo, err := handler.appCrudOperationService.GetAppMetaInfo(appId, app.ZERO_INSTALLED_APP_ID, app.ZERO_ENVIRONMENT_ID)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo in GetAppAllDetail", "err", err, "appId", appId)
		return nil, err, http.StatusInternalServerError
	}

	if appMetaInfo == nil {
		err = errors.New("invalid appId - appMetaInfo is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId)
		return nil, err, http.StatusBadRequest
	}

	var appLabelsRes []*appBean.AppLabel
	if len(appMetaInfo.Labels) > 0 {
		for _, label := range appMetaInfo.Labels {
			appLabelsRes = append(appLabelsRes, &appBean.AppLabel{
				Key:       label.Key,
				Value:     label.Value,
				Propagate: label.Propagate,
			})
		}
	}
	appMetadataResp := &appBean.AppMetadata{
		AppName:     appMetaInfo.AppName,
		ProjectName: appMetaInfo.ProjectName,
		Labels:      appLabelsRes,
	}

	return appMetadataResp, nil, http.StatusOK
}

// get/build git materials
func (handler CoreAppRestHandlerImpl) buildAppGitMaterials(appId int) ([]*appBean.GitMaterial, error, int) {
	handler.logger.Debugw("Getting app detail - git materials", "appId", appId)

	gitMaterials := handler.pipelineBuilder.GetMaterialsForAppId(appId)
	var gitMaterialsResp []*appBean.GitMaterial
	if len(gitMaterials) > 0 {
		for _, gitMaterial := range gitMaterials {
			gitRegistry, err := handler.gitProviderReadService.FetchOneGitProvider(strconv.Itoa(gitMaterial.GitProviderId))
			if err != nil {
				handler.logger.Errorw("service err, getGitProvider in GetAppAllDetail", "err", err, "appId", appId)
				return nil, err, http.StatusInternalServerError
			}

			gitMaterialsResp = append(gitMaterialsResp, &appBean.GitMaterial{
				GitRepoUrl:      gitMaterial.Url,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				GitProviderUrl:  gitRegistry.Url,
			})
		}
	}
	return gitMaterialsResp, nil, http.StatusOK
}

// get/build docker build config
func (handler CoreAppRestHandlerImpl) buildDockerConfig(appId int) (*appBean.DockerConfig, error, int) {
	handler.logger.Debugw("Getting app detail - docker build", "appId", appId)

	ciConfig, err := handler.pipelineBuilder.GetCiPipeline(appId)
	if errResponse, ok := err.(*util2.ApiError); ok && errResponse.UserMessage == "no ci pipeline exists" {
		handler.logger.Warnw("docker config not available for app, GetCiPipeline in GetAppAllDetail", "err", err, "appId", appId)
		return nil, nil, http.StatusOK
	}

	if err != nil {
		handler.logger.Errorw("service err, GetCiPipeline in GetAppAllDetail", "err", err, "appId", appId)
		return nil, err, http.StatusInternalServerError
	}

	//getting gitMaterialUrl by id
	gitMaterial, err := handler.gitMaterialReadService.FindById(ciConfig.CiBuildConfig.GitMaterialId)
	if err != nil {
		handler.logger.Errorw("error in fetching materialUrl by ID in GetAppAllDetail", "err", err, "gitMaterialId", ciConfig.CiBuildConfig.GitMaterialId)
		return nil, err, http.StatusInternalServerError
	}

	dockerConfig := &appBean.DockerConfig{
		DockerRegistry:   ciConfig.DockerRegistry,
		DockerRepository: ciConfig.DockerRepository,
		CiBuildConfig:    ciConfig.CiBuildConfig,
		CheckoutPath:     gitMaterial.CheckoutPath,
	}

	return dockerConfig, nil, http.StatusOK
}

// get/build global deployment template
func (handler CoreAppRestHandlerImpl) buildAppDeploymentTemplate(appId int) (*appBean.DeploymentTemplate, error, int) {
	handler.logger.Debugw("Getting app detail - deployment template", "appId", appId)

	//for global template, to bypass env overrides using envId = 0
	return handler.buildAppEnvironmentDeploymentTemplate(appId, 0)
}

// get/build environment deployment template
// using this method for global as well, for global pass envId = 0
func (handler CoreAppRestHandlerImpl) buildAppEnvironmentDeploymentTemplate(appId int, envId int) (*appBean.DeploymentTemplate, error, int) {
	handler.logger.Debugw("Getting app detail - environment deployment template", "appId", appId, "envId", envId)

	chartRefData, err := handler.chartService.ChartRefAutocompleteForAppOrEnv(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, ChartRefAutocompleteForAppOrEnv in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		return nil, err, http.StatusInternalServerError
	}

	if chartRefData == nil {
		err = errors.New("invalid appId/envId - chartRefData is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId, "envId", envId)
		return nil, err, http.StatusBadRequest
	}

	appDeploymentTemplate, err := handler.chartReadService.FindLatestChartForAppByAppId(appId)
	if err != nil {
		if err != pg.ErrNoRows {
			handler.logger.Errorw("service err, GetDeploymentTemplate in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
			return nil, err, http.StatusInternalServerError
		} else {
			handler.logger.Warnw("no charts configured for app, GetDeploymentTemplate in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
			return nil, nil, http.StatusOK
		}
	}

	if appDeploymentTemplate == nil {
		err = errors.New("invalid appId - deploymentTemplate is null")
		handler.logger.Errorw("Validation error ", "err", err, "appId", appId, "envId", envId)
		return nil, err, http.StatusBadRequest
	}

	//set deployment template & showAppMetrics && isOverride
	var showAppMetrics bool
	var deploymentTemplateRaw json.RawMessage
	var chartRefId int
	var isOverride bool
	var isBasicViewLocked bool
	var currentViewEditor models.ChartsViewEditorType
	if envId > 0 {
		//on env level
		env, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, envId, chartRefData.LatestEnvChartRef)
		if err != nil {
			handler.logger.Errorw("service err, GetEnvironmentProperties in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
			return nil, err, http.StatusInternalServerError
		}
		chartRefId = chartRefData.LatestEnvChartRef
		if env.EnvironmentConfig.IsOverride {
			deploymentTemplateRaw = env.EnvironmentConfig.EnvOverrideValues
			showAppMetrics = *env.AppMetrics
			isOverride = true
			isBasicViewLocked = env.EnvironmentConfig.IsBasicViewLocked
			currentViewEditor = env.EnvironmentConfig.CurrentViewEditor
		} else {
			showAppMetrics = appDeploymentTemplate.IsAppMetricsEnabled
			deploymentTemplateRaw = appDeploymentTemplate.DefaultAppOverride
			isBasicViewLocked = appDeploymentTemplate.IsBasicViewLocked
			currentViewEditor = appDeploymentTemplate.CurrentViewEditor
		}
	} else {
		//on app level
		showAppMetrics = appDeploymentTemplate.IsAppMetricsEnabled
		deploymentTemplateRaw = appDeploymentTemplate.DefaultAppOverride
		chartRefId = chartRefData.LatestAppChartRef
		isBasicViewLocked = appDeploymentTemplate.IsBasicViewLocked
		currentViewEditor = appDeploymentTemplate.CurrentViewEditor
	}

	var deploymentTemplateObj map[string]interface{}
	if deploymentTemplateRaw != nil {
		err = json.Unmarshal([]byte(deploymentTemplateRaw), &deploymentTemplateObj)
		if err != nil {
			handler.logger.Errorw("service err, un-marshaling fail in deploymentTemplate", "err", err, "appId", appId)
			return nil, err, http.StatusInternalServerError
		}
	}

	deploymentTemplateResp := &appBean.DeploymentTemplate{
		ChartRefId:        chartRefId,
		Template:          deploymentTemplateObj,
		ShowAppMetrics:    showAppMetrics,
		IsOverride:        isOverride,
		IsBasicViewLocked: isBasicViewLocked,
		CurrentViewEditor: currentViewEditor,
	}

	return deploymentTemplateResp, nil, http.StatusOK
}

// validate and build workflows
func (handler CoreAppRestHandlerImpl) buildAppWorkflows(request *appWorkflowBean.WorkflowCloneRequest) ([]*appBean.AppWorkflow, error, int) {
	handler.logger.Debugw("Getting app detail - workflows", "appId", request.AppId)
	var workflowsList []appWorkflowBean.AppWorkflowDto
	var err error
	if len(request.WorkflowName) != 0 {
		workflow, err := handler.appWorkflowService.FindAppWorkflowByName(request.WorkflowName, request.AppId)
		if err != nil {
			handler.logger.Errorw("error in fetching workflow by name", "err", err, "workflowName", request.WorkflowName, "appId", request.AppId)
			return nil, err, http.StatusInternalServerError
		}
		workflowsList = []appWorkflowBean.AppWorkflowDto{workflow}
	} else if request.WorkflowId > 0 {
		workflow, err := handler.appWorkflowService.FindAppWorkflowById(request.WorkflowId, request.AppId)
		if err != nil {
			handler.logger.Errorw("error in fetching workflow by id", "err", err, "workflowName", request.WorkflowName, "appId", request.AppId)
			return nil, err, http.StatusInternalServerError
		}
		workflowsList = []appWorkflowBean.AppWorkflowDto{workflow}
	} else {
		workflowsList, err = handler.appWorkflowService.FindAppWorkflows(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in fetching workflows for app in GetAppAllDetail", "err", err)
			return nil, err, http.StatusInternalServerError
		}
	}

	var appWorkflowsResp []*appBean.AppWorkflow
	for _, workflow := range workflowsList {

		workflowResp := &appBean.AppWorkflow{
			Name: workflow.Name,
		}

		var cdPipelinesResp []*appBean.CdPipelineDetails
		for _, workflowMapping := range workflow.AppWorkflowMappingDto {
			if workflowMapping.Type == appWorkflow2.CIPIPELINE {
				ciPipeline, err := handler.pipelineBuilder.GetCiPipelineById(workflowMapping.ComponentId)
				if err != nil {
					handler.logger.Errorw("service err, GetCiPipelineById in GetAppAllDetail", "err", err, "appId", request.AppId)
					return nil, err, http.StatusInternalServerError
				}

				ciPipelineResp, err := handler.buildCiPipelineResp(request.AppId, ciPipeline)
				if err != nil {
					handler.logger.Errorw("service err, buildCiPipelineResp in GetAppAllDetail", "err", err, "appId", request.AppId)
					return nil, err, http.StatusInternalServerError
				}
				workflowResp.CiPipeline = ciPipelineResp
			}

			if workflowMapping.Type == appWorkflow2.CDPIPELINE {
				cdPipeline, err := handler.pipelineBuilder.GetCdPipelineById(workflowMapping.ComponentId)
				if err != nil {
					handler.logger.Errorw("service err, GetCdPipelineById in GetAppAllDetail", "err", err, "appId", request.AppId)
					return nil, err, http.StatusInternalServerError
				}
				if request.EnvironmentId > 0 && request.EnvironmentId != cdPipeline.EnvironmentId {
					// if environment id present in request it should match cd pipeline, else skip
					continue
				}
				cdPipelineResp, err := handler.buildCdPipelineResp(request.AppId, cdPipeline)
				if err != nil {
					handler.logger.Errorw("service err, buildCdPipelineResp in GetAppAllDetail", "err", err, "appId", request.AppId)
					return nil, err, http.StatusInternalServerError
				}
				cdPipelinesResp = append(cdPipelinesResp, cdPipelineResp)
			}
		}

		workflowResp.CdPipelines = cdPipelinesResp
		appWorkflowsResp = append(appWorkflowsResp, workflowResp)

	}

	return appWorkflowsResp, nil, http.StatusOK
}

// build ci pipeline resp
func (handler CoreAppRestHandlerImpl) buildCiPipelineResp(appId int, ciPipeline *bean.CiPipeline) (*appBean.CiPipelineDetails, error) {
	handler.logger.Debugw("Getting app detail - build ci pipeline resp", "appId", appId)

	if ciPipeline == nil {
		return nil, nil
	}

	ciPipelineResp := &appBean.CiPipelineDetails{
		Name:                     ciPipeline.Name,
		IsManual:                 ciPipeline.IsManual,
		DockerBuildArgs:          ciPipeline.DockerArgs,
		VulnerabilityScanEnabled: ciPipeline.ScanEnabled,
		IsExternal:               ciPipeline.IsExternal,
		ParentCiPipeline:         ciPipeline.ParentCiPipeline,
		ParentAppId:              ciPipeline.ParentAppId,
		LinkedCount:              ciPipeline.LinkedCount,
		PipelineType:             string(ciPipeline.PipelineType),
	}

	//build ciPipelineMaterial resp
	var ciPipelineMaterialsConfig []*appBean.CiPipelineMaterialConfig
	for _, ciMaterial := range ciPipeline.CiMaterial {
		gitMaterial, err := handler.gitMaterialReadService.FindById(ciMaterial.GitMaterialId)
		if err != nil {
			handler.logger.Errorw("service err, GitMaterialById in GetAppAllDetail", "err", err, "appId", appId)
			return nil, err
		}
		ciPipelineMaterialConfig := &appBean.CiPipelineMaterialConfig{
			Type:          ciMaterial.Source.Type,
			Value:         ciMaterial.Source.Value,
			CheckoutPath:  gitMaterial.CheckoutPath,
			GitMaterialId: gitMaterial.Id,
		}
		ciPipelineMaterialsConfig = append(ciPipelineMaterialsConfig, ciPipelineMaterialConfig)
	}

	ciPipelineResp.CiPipelineMaterialsConfig = ciPipelineMaterialsConfig

	//build docker pre-build script
	var beforeDockerBuildScriptsResp []*appBean.BuildScript
	for _, beforeDockerBuildScript := range ciPipeline.BeforeDockerBuildScripts {
		beforeDockerBuildScriptResp := &appBean.BuildScript{
			Name:                beforeDockerBuildScript.Name,
			Index:               beforeDockerBuildScript.Index,
			Script:              beforeDockerBuildScript.Script,
			ReportDirectoryPath: beforeDockerBuildScript.OutputLocation,
		}
		beforeDockerBuildScriptsResp = append(beforeDockerBuildScriptsResp, beforeDockerBuildScriptResp)
	}
	ciPipelineResp.BeforeDockerBuildScripts = beforeDockerBuildScriptsResp

	//build docker post build script
	var afterDockerBuildScriptsResp []*appBean.BuildScript
	for _, afterDockerBuildScript := range ciPipeline.AfterDockerBuildScripts {
		afterDockerBuildScriptResp := &appBean.BuildScript{
			Name:                afterDockerBuildScript.Name,
			Index:               afterDockerBuildScript.Index,
			Script:              afterDockerBuildScript.Script,
			ReportDirectoryPath: afterDockerBuildScript.OutputLocation,
		}
		afterDockerBuildScriptsResp = append(afterDockerBuildScriptsResp, afterDockerBuildScriptResp)
	}
	ciPipelineResp.AfterDockerBuildScripts = afterDockerBuildScriptsResp

	//getting pre stage and post stage details
	preStageDetail, postStageDetail, err := handler.pipelineStageService.GetCiPipelineStageDataDeepCopy(ciPipeline.Id)
	if err != nil {
		handler.logger.Errorw("error in getting pre & post stage detail by ciPipelineId", "err", err, "ciPipelineId", ciPipeline.Id)
		return nil, err
	}
	ciPipelineResp.PreBuildStage = preStageDetail
	ciPipelineResp.PostBuildStage = postStageDetail
	return ciPipelineResp, nil
}

// build cd pipeline resp
func (handler CoreAppRestHandlerImpl) buildCdPipelineResp(appId int, cdPipeline *bean.CDPipelineConfigObject) (*appBean.CdPipelineDetails, error) {
	handler.logger.Debugw("Getting app detail - build cd pipeline resp", "appId", appId)

	if cdPipeline == nil {
		return nil, nil
	}

	cdPipelineResp := &appBean.CdPipelineDetails{
		Name:                   cdPipeline.Name,
		EnvironmentName:        cdPipeline.EnvironmentName,
		TriggerType:            cdPipeline.TriggerType,
		DeploymentStrategyType: cdPipeline.DeploymentTemplate,
		RunPreStageInEnv:       cdPipeline.RunPreStageInEnv,
		RunPostStageInEnv:      cdPipeline.RunPostStageInEnv,
		IsClusterCdActive:      cdPipeline.CdArgoSetup,
	}

	//build DeploymentStrategies resp
	var deploymentTemplateStrategiesResp []*appBean.DeploymentStrategy
	for _, strategy := range cdPipeline.Strategies {
		deploymentTemplateStrategyResp := &appBean.DeploymentStrategy{
			DeploymentStrategyType: strategy.DeploymentTemplate,
			IsDefault:              strategy.Default,
		}
		var configObj map[string]interface{}
		if strategy.Config != nil {
			err := json.Unmarshal([]byte(strategy.Config), &configObj)
			if err != nil {
				handler.logger.Errorw("service err, un-marshaling fail in config object in cd", "err", err, "appId", appId)
				return nil, err
			}
		}
		deploymentTemplateStrategyResp.Config = configObj
		deploymentTemplateStrategiesResp = append(deploymentTemplateStrategiesResp, deploymentTemplateStrategyResp)
	}
	cdPipelineResp.DeploymentStrategies = deploymentTemplateStrategiesResp

	//set pre-deploy and post-deploy stage steps for multi step execution
	cdPipelineMigrated, err := pipeline.ConvertStageYamlScriptsToPipelineStageSteps(cdPipeline)
	if err != nil {
		handler.logger.Errorw("service err, InitiateMigrationOfStageScriptsToPipelineStageSteps", "err", err, "appId", appId, "pipelineId", cdPipeline.Id)
		return nil, err
	}
	cdPipelineResp.PreDeployStage = cdPipelineMigrated.PreDeployStage
	cdPipelineResp.PostDeployStage = cdPipelineMigrated.PostDeployStage
	//set pre stage config maps secret names
	preStageConfigMapSecretNames := cdPipeline.PreStageConfigMapSecretNames
	cdPipelineResp.PreStageConfigMapSecretNames = &appBean.CdStageConfigMapSecretNames{
		ConfigMaps: preStageConfigMapSecretNames.ConfigMaps,
		Secrets:    preStageConfigMapSecretNames.Secrets,
	}

	//set post stage config maps secret names
	postStageConfigMapSecretNames := cdPipeline.PostStageConfigMapSecretNames
	cdPipelineResp.PostStageConfigMapSecretNames = &appBean.CdStageConfigMapSecretNames{
		ConfigMaps: postStageConfigMapSecretNames.ConfigMaps,
		Secrets:    postStageConfigMapSecretNames.Secrets,
	}

	return cdPipelineResp, nil
}

// get/build global config maps
func (handler CoreAppRestHandlerImpl) buildAppGlobalConfigMaps(appId int) ([]*appBean.ConfigMap, error, int) {
	handler.logger.Debugw("Getting app detail - global config maps", "appId", appId)

	configMapData, err := handler.configMapService.CMGlobalFetch(appId)
	if err != nil {
		handler.logger.Errorw("service err, CMGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		return nil, err, http.StatusInternalServerError
	}

	return handler.buildAppConfigMaps(appId, 0, configMapData)
}

// get/build environment config maps
func (handler CoreAppRestHandlerImpl) buildAppEnvironmentConfigMaps(appId int, envId int) ([]*appBean.ConfigMap, error, int) {
	handler.logger.Debugw("Getting app detail - environment config maps", "appId", appId, "envId", envId)

	configMapData, err := handler.configMapService.CMEnvironmentFetch(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CMEnvironmentFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		return nil, err, http.StatusInternalServerError
	}

	return handler.buildAppConfigMaps(appId, envId, configMapData)
}

// get/build config maps
func (handler CoreAppRestHandlerImpl) buildAppConfigMaps(appId int, envId int, configMapData *bean2.ConfigDataRequest) ([]*appBean.ConfigMap, error, int) {
	handler.logger.Debugw("Getting app detail - config maps", "appId", appId, "envId", envId)

	var configMapsResp []*appBean.ConfigMap
	if configMapData != nil && len(configMapData.ConfigData) > 0 {
		for _, configMap := range configMapData.ConfigData {

			//initialise
			configMapRes := &appBean.ConfigMap{
				Name:       configMap.Name,
				IsExternal: configMap.External,
				UsageType:  configMap.Type,
			}

			//set data
			data := configMap.Data
			if configMap.Data == nil {
				//it means env cm is inheriting from base cm
				data = configMap.DefaultData
			}
			var dataObj map[string]interface{}
			if data != nil {
				err := json.Unmarshal(data, &dataObj)
				if err != nil {
					handler.logger.Errorw("service err, un-marshaling of data fail in config map", "err", err, "appId", appId)
					return nil, err, http.StatusInternalServerError
				}
			}
			configMapRes.Data = dataObj

			//set data volume usage type
			if configMap.Type == util.ConfigMapSecretUsageTypeVolume {
				dataVolumeUsageConfig := &appBean.ConfigMapSecretDataVolumeUsageConfig{
					FilePermission: configMap.FilePermission,
					SubPath:        configMap.SubPath,
				}
				considerGlobalDefaultData := envId > 0 && configMap.Data == nil
				if considerGlobalDefaultData {
					dataVolumeUsageConfig.MountPath = configMap.DefaultMountPath
				} else {
					dataVolumeUsageConfig.MountPath = configMap.MountPath
				}

				configMapRes.DataVolumeUsageConfig = dataVolumeUsageConfig
			}

			configMapsResp = append(configMapsResp, configMapRes)
		}
	}
	return configMapsResp, nil, http.StatusOK
}

// get/build global secrets
func (handler CoreAppRestHandlerImpl) buildAppGlobalSecrets(appId int) ([]*appBean.Secret, error, int) {
	handler.logger.Debugw("Getting app detail - global secret", "appId", appId)

	secretData, err := handler.configMapService.CSGlobalFetch(appId)
	if err != nil {
		handler.logger.Errorw("service err, CSGlobalFetch in GetAppAllDetail", "err", err, "appId", appId)
		return nil, err, http.StatusInternalServerError
	}

	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {

		for _, secretConfig := range secretData.ConfigData {
			secretDataWithData, err := handler.configMapService.CSGlobalFetchForEdit(secretConfig.Name, secretData.Id)
			if err != nil {
				handler.logger.Errorw("service err, CSGlobalFetch-CSGlobalFetchForEdit in GetAppAllDetail", "err", err, "appId", appId)
				return nil, err, http.StatusInternalServerError
			}

			secretRes, err, statusCode := handler.buildAppSecrets(appId, 0, secretDataWithData)
			if err != nil {
				handler.logger.Errorw("service err, CSGlobalFetch-buildAppSecrets in GetAppAllDetail", "err", err, "appId", appId)
				return nil, err, statusCode
			}

			for _, secret := range secretRes {
				secretsResp = append(secretsResp, secret)
			}
		}
	}

	return secretsResp, nil, http.StatusOK
}

// get/build environment secrets
func (handler CoreAppRestHandlerImpl) buildAppEnvironmentSecrets(appId int, envId int) ([]*appBean.Secret, error, int) {
	handler.logger.Debugw("Getting app detail - env secrets", "appId", appId, "envId", envId)

	secretData, err := handler.configMapService.CSEnvironmentFetch(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, CSEnvironmentFetch in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
		return nil, err, http.StatusInternalServerError
	}

	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {

		for _, secretConfig := range secretData.ConfigData {
			secretDataWithData, err := handler.configMapService.CSEnvironmentFetchForEdit(secretConfig.Name, secretData.Id, appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, CSEnvironmentFetchForEdit in GetAppAllDetail", "err", err, "appId", appId, "envId", envId)
				return nil, err, http.StatusInternalServerError
			}
			secretDataWithData.ConfigData[0].DefaultData = secretConfig.DefaultData

			secretRes, err, statusCode := handler.buildAppSecrets(appId, envId, secretDataWithData)
			if err != nil {
				handler.logger.Errorw("service err, CSGlobalFetch-buildAppSecrets in GetAppAllDetail", "err", err, "appId", appId)
				return nil, err, statusCode
			}

			for _, secret := range secretRes {
				secretsResp = append(secretsResp, secret)
			}
		}
	}

	return secretsResp, nil, http.StatusOK
}

// get/build secrets
func (handler CoreAppRestHandlerImpl) buildAppSecrets(appId int, envId int, secretData *bean2.ConfigDataRequest) ([]*appBean.Secret, error, int) {
	handler.logger.Debugw("Getting app detail - secrets", "appId", appId, "envId", envId)

	var secretsResp []*appBean.Secret
	if secretData != nil && len(secretData.ConfigData) > 0 {
		for _, secret := range secretData.ConfigData {

			//initialise
			globalSecret := &appBean.Secret{
				Name:         secret.Name,
				RoleArn:      secret.RoleARN,
				IsExternal:   secret.External,
				UsageType:    secret.Type,
				ExternalType: secret.ExternalSecretType,
			}

			//set data
			data := secret.Data
			var dataObj map[string]interface{}
			if data != nil {
				err := json.Unmarshal([]byte(data), &dataObj)
				if err != nil {
					handler.logger.Errorw("service err, un-marshaling of data fail in secret", "err", err, "appId", appId)
					return nil, err, http.StatusInternalServerError
				}
			}
			globalSecret.Data = dataObj

			//set external data
			externalSecrets := secret.ExternalSecret
			var externalSecretsResp []*appBean.ExternalSecret
			if len(externalSecrets) > 0 {
				for _, externalSecret := range externalSecrets {
					externalSecretsResp = append(externalSecretsResp, &appBean.ExternalSecret{
						Name:     externalSecret.Name,
						Key:      externalSecret.Key,
						Property: externalSecret.Property,
						IsBinary: externalSecret.IsBinary,
					})
				}
			}
			globalSecret.ExternalSecretData = externalSecretsResp

			//set data volume usage type
			if secret.Type == util.ConfigMapSecretUsageTypeVolume {
				globalSecret.DataVolumeUsageConfig = &appBean.ConfigMapSecretDataVolumeUsageConfig{
					SubPath:        secret.SubPath,
					FilePermission: secret.FilePermission,
					ESOSubPath:     secret.ESOSubPath,
				}
				considerGlobalDefaultData := envId > 0 && secret.Data == nil
				if considerGlobalDefaultData {
					globalSecret.DataVolumeUsageConfig.MountPath = secret.DefaultMountPath
				} else {
					globalSecret.DataVolumeUsageConfig.MountPath = secret.MountPath
				}
			}

			secretsResp = append(secretsResp, globalSecret)
		}
	}
	return secretsResp, nil, http.StatusOK
}

// get/build environment overrides
func (handler CoreAppRestHandlerImpl) buildEnvironmentOverrides(ctx context.Context, appId int, token string) (map[string]*appBean.EnvironmentOverride, error, int) {
	handler.logger.Debugw("Getting app detail - env override", "appId", appId)

	appEnvironments, err := handler.appListingService.FetchOtherEnvironment(ctx, appId)
	if err != nil {
		handler.logger.Errorw("service err, Fetch app environments in GetAppAllDetail", "err", err, "appId", appId)
		return nil, err, http.StatusInternalServerError
	}

	environmentOverrides := make(map[string]*appBean.EnvironmentOverride)
	if len(appEnvironments) > 0 {
		for _, appEnvironment := range appEnvironments {
			environmentOverride, err, _ := handler.buildEnvironmentOverride(appId, appEnvironment.EnvironmentId, token)
			if err != nil {
				handler.logger.Errorw("service err", "err", err)
				return nil, err, http.StatusInternalServerError
			}
			override := environmentOverride[appEnvironment.EnvironmentName]
			environmentOverrides[appEnvironment.EnvironmentName] = &appBean.EnvironmentOverride{
				Secrets:            override.Secrets,
				ConfigMaps:         override.ConfigMaps,
				DeploymentTemplate: override.DeploymentTemplate,
			}
		}
	}
	return environmentOverrides, nil, http.StatusOK
}

// get/build environment overrides
func (handler CoreAppRestHandlerImpl) buildEnvironmentOverride(appId int, environmentId int, token string) (map[string]*appBean.EnvironmentOverride, error, int) {
	handler.logger.Debugw("Getting app detail - env override", "appId", appId)
	environmentOverrides := make(map[string]*appBean.EnvironmentOverride)
	//check RBAC for environment
	object := handler.enforcerUtil.GetEnvRBACNameByAppId(appId, environmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
		handler.logger.Errorw("Unauthorized User for env update action", "appId", appId, "envId", environmentId)
		return nil, fmt.Errorf("unauthorized user"), http.StatusForbidden
	}
	//RBAC end

	environment, err := handler.environmentRepository.FindById(environmentId)
	if err != nil {
		handler.logger.Errorw("service err, for fetching environment model", "err", err, "appId", appId)
		return nil, err, http.StatusInternalServerError
	}

	envDeploymentTemplateResp, err, statusCode := handler.buildAppEnvironmentDeploymentTemplate(appId, environmentId)
	if err != nil {
		return nil, err, statusCode
	}
	envSecretsResp, err, statusCode := handler.buildAppEnvironmentSecrets(appId, environmentId)
	if err != nil {
		return nil, err, statusCode
	}

	envConfigMapsResp, err, statusCode := handler.buildAppEnvironmentConfigMaps(appId, environmentId)
	if err != nil {
		return nil, err, statusCode
	}

	environmentOverrides[environment.Name] = &appBean.EnvironmentOverride{
		Secrets:            envSecretsResp,
		ConfigMaps:         envConfigMapsResp,
		DeploymentTemplate: envDeploymentTemplateResp,
	}

	return environmentOverrides, nil, http.StatusOK
}

//GetApp related methods ends

//Create App related methods starts

// create a blank app with metadata
func (handler CoreAppRestHandlerImpl) createBlankApp(appMetadata *appBean.AppMetadata, userId int32) (*bean.CreateAppDTO, error, int) {
	handler.logger.Infow("Create App - creating blank app", "appMetadata", appMetadata)

	//validating app metadata
	err := handler.validator.Struct(appMetadata)
	if err != nil {
		handler.logger.Errorw("validation err, AppMetadata in create app by API", "err", err, "AppMetadata", appMetadata)
		return nil, err, http.StatusBadRequest
	}

	team, err := handler.teamReadService.FindByTeamName(appMetadata.ProjectName)
	if err != nil {
		handler.logger.Infow("no project found by name in CreateApp request by API")
		return nil, err, http.StatusBadRequest
	}

	handler.logger.Infow("Create App - creating blank app with metadata", "appMetadata", appMetadata)

	createAppRequest := &bean.CreateAppDTO{
		AppName: appMetadata.AppName,
		TeamId:  team.Id,
		UserId:  userId,
	}

	var appLabels []*bean.Label
	for _, requestLabel := range appMetadata.Labels {
		appLabel := &bean.Label{
			Key:       requestLabel.Key,
			Value:     requestLabel.Value,
			Propagate: requestLabel.Propagate,
		}
		appLabels = append(appLabels, appLabel)
	}
	createAppRequest.AppLabels = appLabels

	createAppResp, err := handler.pipelineBuilder.CreateApp(createAppRequest)
	if err != nil {
		handler.logger.Errorw("service err, CreateApp in CreateBlankApp", "err", err, "CreateApp", createAppRequest)
		return nil, err, http.StatusInternalServerError
	}

	return createAppResp, nil, http.StatusOK
}

// delete app
func (handler CoreAppRestHandlerImpl) deleteApp(ctx context.Context, appId int, userId int32) error {
	handler.logger.Infow("Delete app", "appid", appId)

	//finding all workflows for app
	workflowsList, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.logger.Errorw("error in fetching workflows for app in DeleteApp", "err", err)
		return err
	}

	//deleting all ci, cd pipelines & workflows before deleting app
	if len(workflowsList) > 0 {

		// delete all CD pipelines for app starts
		cdPipelines, err := handler.pipelineBuilder.GetCdPipelinesForApp(appId)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("service err, GetCdPipelines in DeleteApp", "err", err, "appId", appId)
			return err
		}
		if err != pg.ErrNoRows {
			for _, cdPipeline := range cdPipelines.Pipelines {
				cdPipelineDeleteRequest := &bean.CDPatchRequest{
					AppId:            appId,
					UserId:           userId,
					Action:           bean.CD_DELETE,
					ForceDelete:      true,
					NonCascadeDelete: false,
					Pipeline:         cdPipeline,
				}
				_, err = handler.pipelineBuilder.PatchCdPipelines(cdPipelineDeleteRequest, ctx)
				if err != nil {
					handler.logger.Errorw("err in deleting cd pipeline in DeleteApp", "err", err, "payload", cdPipelineDeleteRequest)
					return err
				}
			}

		}
		// delete all CD pipelines for app ends

		// delete all CI pipelines for app starts
		ciPipelines, err := handler.pipelineBuilder.GetCiPipeline(appId)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("service err, GetCiPipelines in DeleteApp", "err", err, "appId", appId)
			return err
		}
		if err != pg.ErrNoRows {

			for _, ciPipeline := range ciPipelines.CiPipelines {
				ciPipelineDeleteRequest := &bean.CiPatchRequest{
					AppId:      appId,
					UserId:     userId,
					Action:     bean.DELETE,
					CiPipeline: ciPipeline,
				}
				_, err := handler.pipelineBuilder.PatchCiPipeline(ciPipelineDeleteRequest)
				if err != nil {
					handler.logger.Errorw("err in deleting ci pipeline in DeleteApp", "err", err, "payload", ciPipelineDeleteRequest)
					return err
				}
			}
		}
		// delete all CI pipelines for app ends

		// delete all workflows for app starts
		for _, workflow := range workflowsList {
			err = handler.appWorkflowService.DeleteAppWorkflow(workflow.Id, userId)
			if err != nil {
				handler.logger.Errorw("service err, DeleteAppWorkflow ")
				return err
			}
		}
		// delete all workflows for app ends
	}

	// delete app
	err = handler.pipelineBuilder.DeleteApp(appId, userId)
	if err != nil {
		handler.logger.Errorw("service error, DeleteApp", "err", err, "appId", appId)
		return err
	}
	return nil
}

// create git materials
func (handler CoreAppRestHandlerImpl) createGitMaterials(appId int, gitMaterials []*appBean.GitMaterial, userId int32) (error, int) {
	handler.logger.Infow("Create App - creating git materials", "appId", appId, "GitMaterials", gitMaterials)

	createMaterialRequest := &bean.CreateMaterialDTO{
		AppId:  appId,
		UserId: userId,
	}

	for _, material := range gitMaterials {
		err := handler.validator.Struct(material)
		if err != nil {
			handler.logger.Errorw("validation err, gitMaterial in CreateGitMaterials", "err", err, "GitMaterial", material)
			return err, http.StatusBadRequest
		}

		//finding gitProvider to update gitMaterial
		gitProvider, err := handler.gitProviderReadService.FindByUrl(material.GitProviderUrl)
		if err != nil {
			handler.logger.Errorw("service err, FindByUrl in CreateGitMaterials", "err", err, "gitProviderUrl", material.GitProviderUrl)
			return err, http.StatusInternalServerError
		}

		//validating git material by git provider auth mode
		var hasPrefixResult bool
		var expectedUrlPrefix string
		if gitProvider.AuthMode == constants.AUTH_MODE_SSH {
			hasPrefixResult = strings.HasPrefix(material.GitRepoUrl, app2.SSH_URL_PREFIX)
			expectedUrlPrefix = app2.SSH_URL_PREFIX
		} else {
			hasPrefixResult = strings.HasPrefix(material.GitRepoUrl, app2.HTTPS_URL_PREFIX)
			expectedUrlPrefix = app2.HTTPS_URL_PREFIX
		}
		if !hasPrefixResult {
			handler.logger.Errorw("validation err, CreateGitMaterials : invalid git material url", "err", err, "gitMaterialUrl", material.GitRepoUrl)
			return fmt.Errorf("validation for url failed, expected url prefix : %s", expectedUrlPrefix), http.StatusBadRequest
		}

		gitMaterialRequest := &bean.GitMaterial{
			Url:             material.GitRepoUrl,
			GitProviderId:   gitProvider.Id,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
		}

		createMaterialRequest.Material = append(createMaterialRequest.Material, gitMaterialRequest)
	}

	_, err := handler.pipelineBuilder.CreateMaterialsForApp(createMaterialRequest)
	if err != nil {
		handler.logger.Errorw("service err, CreateMaterialsForApp in CreateGitMaterials", "err", err, "CreateMaterial", createMaterialRequest)
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

// create docker config
func (handler CoreAppRestHandlerImpl) createDockerConfig(appId int, dockerConfig *appBean.DockerConfig, userId int32) (error, int) {
	handler.logger.Infow("Create App - creating docker config", "appId", appId, "DockerConfig", dockerConfig)
	dockerBuildConfig := dockerConfig.DockerBuildConfig
	if dockerBuildConfig != nil {
		dockerConfig.CheckoutPath = dockerBuildConfig.GitCheckoutPath
		dockerConfig.CiBuildConfig = &pipelineBean.CiBuildConfigBean{
			CiBuildType: pipelineBean.SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &pipelineBean.DockerBuildConfig{
				DockerfilePath:     dockerBuildConfig.DockerfileRelativePath,
				DockerBuildOptions: dockerBuildConfig.DockerBuildOptions,
				Args:               dockerBuildConfig.Args,
				TargetPlatform:     dockerBuildConfig.TargetPlatform,
				BuildContext:       dockerBuildConfig.BuildContext,
			},
		}
	}
	createDockerConfigRequest := &bean.CiConfigRequest{
		AppId:            appId,
		UserId:           userId,
		DockerRegistry:   dockerConfig.DockerRegistry,
		DockerRepository: dockerConfig.DockerRepository,
	}

	//finding gitMaterial by appId and checkoutPath
	gitMaterial, err := handler.gitMaterialReadService.FindByAppIdAndCheckoutPath(appId, dockerConfig.CheckoutPath)
	if err != nil {
		handler.logger.Errorw("service err, FindByAppIdAndCheckoutPath in CreateDockerConfig", "err", err, "appId", appId)
		return err, http.StatusInternalServerError
	}

	ciBuildConfig := dockerConfig.CiBuildConfig
	ciBuildConfig.GitMaterialId = gitMaterial.Id
	createDockerConfigRequest.CiBuildConfig = ciBuildConfig

	_, err = handler.pipelineBuilder.CreateCiPipeline(createDockerConfigRequest)
	if err != nil {
		handler.logger.Errorw("service err, CreateCiPipeline in CreateDockerConfig", "err", err, "createRequest", createDockerConfigRequest)
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

// create global template
func (handler CoreAppRestHandlerImpl) createDeploymentTemplate(ctx context.Context, appId int, deploymentTemplate *appBean.DeploymentTemplate, userId int32) (error, int) {
	handler.logger.Infow("Create App - creating deployment template", "appId", appId, "DeploymentStrategy", deploymentTemplate)

	createDeploymentTemplateRequest := bean3.TemplateRequest{
		AppId:               appId,
		ChartRefId:          deploymentTemplate.ChartRefId,
		IsAppMetricsEnabled: deploymentTemplate.ShowAppMetrics,
		UserId:              userId,
		IsBasicViewLocked:   deploymentTemplate.IsBasicViewLocked,
		CurrentViewEditor:   deploymentTemplate.CurrentViewEditor,
	}

	//marshalling template
	template, err := json.Marshal(deploymentTemplate.Template)
	if err != nil {
		handler.logger.Errorw("service err, could not json marshal template in CreateDeploymentTemplate", "err", err, "appId", appId, "template", deploymentTemplate.Template)
		return err, http.StatusInternalServerError
	}
	templateRequest := json.RawMessage(template)
	createDeploymentTemplateRequest.ValuesOverride = templateRequest

	//creating deployment template
	_, err = handler.chartService.Create(createDeploymentTemplateRequest, ctx)
	if err != nil {
		handler.logger.Errorw("service err, Create in CreateDeploymentTemplate", "err", err, "createRequest", createDeploymentTemplateRequest)
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// create global CMs
func (handler CoreAppRestHandlerImpl) createGlobalConfigMaps(appId int, userId int32, configMaps []*appBean.ConfigMap) (error, int) {
	handler.logger.Infow("Create App - creating global configMap", "appId", appId)

	var appLevelId int
	for _, configMap := range configMaps {

		//getting app level by app id
		if appLevelId == 0 {
			appLevel, err := handler.configMapRepository.GetByAppIdAppLevel(appId)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("error in getting app level by app id in createGlobalConfigMaps", "appId", appId)
				return err, http.StatusInternalServerError
			}

			if appLevel != nil {
				appLevelId = appLevel.Id
			}
		}

		//marshalling configMap data, i.e. key-value pairs
		configMapKeyValueData, err := json.Marshal(configMap.Data)
		if err != nil {
			handler.logger.Errorw("service err, could not json marshal configMap data in CreateGlobalConfigMap", "err", err, "appId", appId, "configMapData", configMap.Data)
			return err, http.StatusInternalServerError
		}

		// build
		configMapData := &bean2.ConfigData{
			Name:     configMap.Name,
			External: configMap.IsExternal,
			Data:     json.RawMessage(configMapKeyValueData),
			Type:     configMap.UsageType,
		}
		dataVolumeUsageConfig := configMap.DataVolumeUsageConfig
		if dataVolumeUsageConfig != nil {
			configMapData.MountPath = dataVolumeUsageConfig.MountPath
			configMapData.SubPath = dataVolumeUsageConfig.SubPath
			configMapData.FilePermission = dataVolumeUsageConfig.FilePermission
		}

		// service call
		var configMapDataRequest []*bean2.ConfigData
		configMapDataRequest = append(configMapDataRequest, configMapData)
		configMapRequest := &bean2.ConfigDataRequest{
			AppId:      appId,
			UserId:     userId,
			Id:         appLevelId,
			ConfigData: configMapDataRequest,
		}
		//using same var for every request, since appId and userID are same
		_, err = handler.configMapService.CMGlobalAddUpdate(configMapRequest)
		if err != nil {
			handler.logger.Errorw("service err, CMGlobalAddUpdate in CreateGlobalConfigMap", "err", err, "appId", appId, "configMapRequest", configMapRequest)
			return err, http.StatusInternalServerError
		}
	}

	return nil, http.StatusOK

}

// create global secrets
func (handler CoreAppRestHandlerImpl) createGlobalSecrets(appId int, userId int32, secrets []*appBean.Secret) (error, int) {
	handler.logger.Infow("Create App - creating global secrets", "appId", appId)

	var appLevelId int
	for _, secret := range secrets {
		//getting app level by app id
		if appLevelId == 0 {
			appLevel, err := handler.configMapRepository.GetByAppIdAppLevel(appId)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("error in getting app level by app id in createGlobalSecrets", "appId", appId)
				return err, http.StatusInternalServerError
			}

			if appLevel != nil {
				appLevelId = appLevel.Id
			}
		}

		// build
		secretData := &bean2.ConfigData{
			Name:               secret.Name,
			External:           secret.IsExternal,
			Type:               secret.UsageType,
			ExternalSecretType: secret.ExternalType,
			RoleARN:            secret.RoleArn,
		}

		dataVolumeUsageConfig := secret.DataVolumeUsageConfig
		if dataVolumeUsageConfig != nil {
			secretData.MountPath = dataVolumeUsageConfig.MountPath
			secretData.SubPath = dataVolumeUsageConfig.SubPath
			secretData.FilePermission = dataVolumeUsageConfig.FilePermission
			secretData.ESOSubPath = dataVolumeUsageConfig.ESOSubPath
		}

		if secret.IsExternal {
			var externalDataRequests []bean2.ExternalSecret
			for _, externalData := range secret.ExternalSecretData {
				externalDataRequest := bean2.ExternalSecret{
					Name:     externalData.Name,
					IsBinary: externalData.IsBinary,
					Key:      externalData.Key,
					Property: externalData.Property,
				}
				externalDataRequests = append(externalDataRequests, externalDataRequest)
			}
			secretData.ExternalSecret = externalDataRequests
		} else {
			secretKeyValueData, err := json.Marshal(secret.Data)
			if err != nil {
				handler.logger.Errorw("service err, could not json marshal secret data in CreateGlobalSecret", "err", err, "appId", appId)
				return err, http.StatusInternalServerError
			}
			secretData.Data = secretKeyValueData
		}

		// service call
		var secretDataRequest []*bean2.ConfigData
		secretDataRequest = append(secretDataRequest, secretData)
		secretRequest := &bean2.ConfigDataRequest{
			AppId:      appId,
			UserId:     userId,
			Id:         appLevelId,
			ConfigData: secretDataRequest,
		}
		//using same var for every request, since appId and userID are same
		_, err := handler.configMapService.CSGlobalAddUpdate(secretRequest)
		if err != nil {
			handler.logger.Errorw("service err, CSGlobalAddUpdate in CreateGlobalSecret", "err", err, "appId", appId)
			return err, http.StatusInternalServerError
		}
	}

	return nil, http.StatusOK
}

// create app workflows
func (handler CoreAppRestHandlerImpl) createWorkflows(ctx context.Context, appId int, userId int32, workflows []*appBean.AppWorkflow) (error, int) {
	handler.logger.Infow("Create App - creating workflows", "appId", appId, "workflows size", len(workflows))
	for _, workflow := range workflows {
		//Create workflow starts (we need to create workflow with given name)
		workflowId, err := handler.createWorkflowInDb(workflow.Name, appId, userId)
		if err != nil {
			handler.logger.Errorw("err in saving new workflow", err, "appId", appId)
			return err, http.StatusInternalServerError
		}
		//Creating workflow ends

		//Creating CI pipeline starts
		ciPipeline, err := handler.createCiPipeline(appId, userId, workflowId, workflow.CiPipeline)
		if err != nil {
			if err.Error() == pipelineBean.PIPELINE_NAME_ALREADY_EXISTS_ERROR {
				handler.logger.Errorw("service err, DeleteAppWorkflow ", "err", err)
				return err, http.StatusBadRequest
			}
			err1 := handler.appWorkflowService.DeleteAppWorkflow(workflowId, userId)
			if err1 != nil {
				handler.logger.Errorw("service err, DeleteAppWorkflow ")
				return err1, http.StatusInternalServerError
			}
			handler.logger.Errorw("err in saving ci pipelines", err, "appId", appId)
			return err, http.StatusInternalServerError
		}
		//Creating CI pipeline ends

		//Creating CD pipeline starts
		err = handler.createCdPipelines(ctx, appId, userId, workflowId, ciPipeline.Id, workflow.CdPipelines)
		if err != nil {
			ciPipelineDeleteRequest := &bean.CiPatchRequest{
				AppId:      appId,
				UserId:     userId,
				Action:     bean.DELETE,
				CiPipeline: ciPipeline,
			}
			_, err1 := handler.pipelineBuilder.PatchCiPipeline(ciPipelineDeleteRequest)
			if err1 != nil {
				handler.logger.Errorw("err in deleting ci pipeline in DeleteApp", "err", err, "payload", ciPipelineDeleteRequest)
				return err1, http.StatusInternalServerError
			}
			err1 = handler.appWorkflowService.DeleteAppWorkflow(workflowId, userId)
			if err1 != nil {
				handler.logger.Errorw("service err, DeleteAppWorkflow ")
				return err1, http.StatusInternalServerError
			}
			handler.logger.Errorw("err in saving cd pipelines", err, "appId", appId)
			return err, http.StatusInternalServerError
		}
		//Creating CD pipeline ends
	}
	return nil, http.StatusOK
}

func (handler CoreAppRestHandlerImpl) createWorkflowInDb(workflowName string, appId int, userId int32) (int, error) {
	//checking if workflow name  already exist or not
	ok, err := handler.appWorkflowService.IsWorkflowNameFound(workflowName, appId)
	if err != nil {
		return 0, err
	}
	// if workflow name already exists then we will assign a new name to the workflow
	if ok {
		workflowName = util.GenerateNewWorkflowName(workflowName)
	}
	wf := &appWorkflow2.AppWorkflow{
		Name:   workflowName,
		AppId:  appId,
		Active: true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedBy: userId,
		},
	}
	savedAppWf, err := handler.appWorkflowRepository.SaveAppWorkflow(wf)
	if err != nil {
		handler.logger.Errorw("err in saving new workflow", err, "appId", appId)
		return 0, err
	}

	return savedAppWf.Id, nil
}

func (handler CoreAppRestHandlerImpl) createCiPipeline(appId int, userId int32, workflowId int, ciPipelineData *appBean.CiPipelineDetails) (*bean.CiPipeline, error) {

	// if ci pipeline is of external type, then throw error as we are not supporting it as of now
	if ciPipelineData.ParentCiPipeline == 0 && ciPipelineData.ParentAppId == 0 && ciPipelineData.IsExternal {
		err := errors.New("external ci pipeline creation is not supported yet")
		handler.logger.Error("external ci pipeline creation is not supported yet")
		return nil, err
	}

	// build ci pipeline materials starts
	var ciMaterialsRequest []*bean.CiMaterial
	for _, ciMaterial := range ciPipelineData.CiPipelineMaterialsConfig {
		var gitMaterial *repository3.GitMaterial
		var err error
		if ciPipelineData.ParentCiPipeline == 0 && ciPipelineData.ParentAppId == 0 {
			//finding gitMaterial by appId and checkoutPath
			gitMaterial, err = handler.gitMaterialReadService.FindByAppIdAndCheckoutPath(appId, ciMaterial.CheckoutPath)
		} else {
			//if linked CI find git material by it's parentAppId and Id
			gitMaterial, err = handler.gitMaterialReadService.FindByAppIdAndGitMaterialId(ciPipelineData.ParentAppId, ciMaterial.GitMaterialId)
		}
		if err != nil {
			handler.logger.Errorw("service err, FindByAppIdAndCheckoutPath in CreateWorkflows", "err", err, "appId", appId)
			return nil, err
		}

		if gitMaterial == nil {
			err = errors.New("gitMaterial is nil")
			handler.logger.Errorw("gitMaterial is nil", "checkoutPath", ciMaterial.CheckoutPath)
			return nil, err
		}

		ciMaterialRequest := &bean.CiMaterial{
			GitMaterialId:   gitMaterial.Id,
			GitMaterialName: gitMaterial.Name,
			Source: &bean.SourceTypeConfig{
				Type:  ciMaterial.Type,
				Value: ciMaterial.Value,
			},
			CheckoutPath: gitMaterial.CheckoutPath,
		}
		ciMaterialsRequest = append(ciMaterialsRequest, ciMaterialRequest)
	}
	// build ci pipeline materials ends

	// build model
	ciPipelineRequest := &bean.CiPatchRequest{
		AppId:         appId,
		UserId:        userId,
		AppWorkflowId: workflowId,
		Action:        bean.CREATE,
		CiPipeline: &bean.CiPipeline{
			Name:                     ciPipelineData.Name,
			IsManual:                 ciPipelineData.IsManual,
			IsExternal:               ciPipelineData.IsExternal,
			Active:                   true,
			BeforeDockerBuildScripts: convertCiBuildScripts(ciPipelineData.BeforeDockerBuildScripts),
			AfterDockerBuildScripts:  convertCiBuildScripts(ciPipelineData.AfterDockerBuildScripts),
			DockerArgs:               ciPipelineData.DockerBuildArgs,
			ScanEnabled:              ciPipelineData.VulnerabilityScanEnabled,
			CiMaterial:               ciMaterialsRequest,
			PreBuildStage:            ciPipelineData.PreBuildStage,
			PostBuildStage:           ciPipelineData.PostBuildStage,
			ParentCiPipeline:         ciPipelineData.ParentCiPipeline,
			ParentAppId:              ciPipelineData.ParentAppId,
			LinkedCount:              ciPipelineData.LinkedCount,
			PipelineType:             common2.PipelineType(ciPipelineData.PipelineType),
		},
	}

	// service call
	res, err := handler.pipelineBuilder.PatchCiPipeline(ciPipelineRequest)
	if err != nil {
		handler.logger.Errorw("service err, PatchCiPipelines", "err", err, "appId", appId)
		return nil, err
	}

	return res.CiPipelines[0], nil
}

func (handler CoreAppRestHandlerImpl) createCdPipelines(ctx context.Context, appId int, userId int32, workflowId int, ciPipelineId int, cdPipelines []*appBean.CdPipelineDetails) error {

	var cdPipelineRequestConfigs []*bean.CDPipelineConfigObject
	for _, cdPipeline := range cdPipelines {
		//getting environment ID by name
		envName := cdPipeline.EnvironmentName
		envModel, err := handler.environmentRepository.FindByName(envName)
		if err != nil {
			handler.logger.Errorw("err in fetching environment details by name", "appId", appId, "envName", envName)
			return err
		}

		if envModel == nil {
			err = errors.New("environment not found for name " + envName)
			handler.logger.Errorw("environment not found for name", "envName", envName)
			return err
		}

		// RBAC has already been verified in Validation function

		// build model
		cdPipelineRequestConfig := &bean.CDPipelineConfigObject{
			Name:                          cdPipeline.Name,
			EnvironmentId:                 envModel.Id,
			Namespace:                     envModel.Namespace,
			AppWorkflowId:                 workflowId,
			CiPipelineId:                  ciPipelineId,
			DeploymentAppType:             cdPipeline.DeploymentAppType,
			PreStage:                      convertCdStages(cdPipeline.PreStage),
			PostStage:                     convertCdStages(cdPipeline.PostStage),
			DeploymentTemplate:            cdPipeline.DeploymentStrategyType,
			TriggerType:                   cdPipeline.TriggerType,
			CdArgoSetup:                   cdPipeline.IsClusterCdActive,
			RunPreStageInEnv:              cdPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             cdPipeline.RunPostStageInEnv,
			PreDeployStage:                cdPipeline.PreDeployStage,
			PostDeployStage:               cdPipeline.PostDeployStage,
			PreStageConfigMapSecretNames:  convertCdPreStageCMorCSNames(cdPipeline.PreStageConfigMapSecretNames),
			PostStageConfigMapSecretNames: convertCdPostStageCMorCSNames(cdPipeline.PostStageConfigMapSecretNames),
		}
		convertedDeploymentStrategies, err := convertCdDeploymentStrategies(cdPipeline.DeploymentStrategies)
		if err != nil {
			handler.logger.Errorw("err in converting deployment strategies for creating cd pipeline", "appId", appId, "Strategies", cdPipeline.DeploymentStrategies)
			return err
		}
		cdPipelineRequestConfig.Strategies = convertedDeploymentStrategies

		cdPipelineRequestConfigs = append(cdPipelineRequestConfigs, cdPipelineRequestConfig)
	}

	// service call
	cdPipelinesRequest := &bean.CdPipelines{
		AppId:     appId,
		UserId:    userId,
		Pipelines: cdPipelineRequestConfigs,
	}
	_, err := handler.pipelineBuilder.CreateCdPipelines(cdPipelinesRequest, ctx)
	if err != nil {
		handler.logger.Errorw("service err, CreateCdPipeline", "err", err, "payload", cdPipelinesRequest)
		return err
	}
	return nil
}

// create environment overrides
func (handler CoreAppRestHandlerImpl) createEnvOverrides(ctx context.Context, appId int, userId int32, environmentOverrides map[string]*appBean.EnvironmentOverride) (error, int) {
	handler.logger.Infow("Create App - creating env overrides", "appId", appId)

	for envName, envOverrideValues := range environmentOverrides {
		envModel, err := handler.environmentRepository.FindByName(envName)

		if err != nil {
			handler.logger.Errorw("err in fetching environment details by name in CreateEnvOverrides", "appId", appId, "envName", envName)
			return err, http.StatusInternalServerError
		}

		if envModel == nil {
			err = errors.New("environment not found for name " + envName)
			handler.logger.Errorw("environment not found for name", "envName", envName)
			return err, http.StatusInternalServerError
		}

		// RBAC has already been verified in Validation function

		envId := envModel.Id

		//creating deployment template override
		envDeploymentTemplate := envOverrideValues.DeploymentTemplate
		if envDeploymentTemplate != nil && envDeploymentTemplate.IsOverride {
			err := handler.createEnvDeploymentTemplate(appId, userId, envModel.Id, envOverrideValues.DeploymentTemplate)
			if err != nil {
				handler.logger.Errorw("err in creating deployment template for env override", "appId", appId, "envName", envName)
				return err, http.StatusInternalServerError
			}
		}

		//creating configMap override
		err = handler.createEnvCM(appId, userId, envId, envOverrideValues.ConfigMaps)
		if err != nil {
			handler.logger.Errorw("err in creating config map for env override", "appId", appId, "envName", envName)
			return err, http.StatusInternalServerError
		}

		//creating secrets override
		err = handler.createEnvSecret(appId, userId, envModel.Id, envOverrideValues.Secrets)
		if err != nil {
			handler.logger.Errorw("err in creating secret for env override", "appId", appId, "envName", envName)
			return err, http.StatusInternalServerError
		}

	}
	return nil, http.StatusOK
}

// create template overrides
func (handler CoreAppRestHandlerImpl) createEnvDeploymentTemplate(appId int, userId int32, envId int, deploymentTemplateOverride *appBean.DeploymentTemplate) error {
	handler.logger.Infow("Create App - creating template override", "appId", appId)

	// build object
	template, err := json.Marshal(deploymentTemplateOverride.Template)
	if err != nil {
		handler.logger.Errorw("json marshaling error env override template in createEnvDeploymentTemplate", "appId", appId, "envId", envId)
		return err
	}
	chartRefId := deploymentTemplateOverride.ChartRefId
	envConfigProperties := &bean2.EnvironmentProperties{
		IsOverride:        true,
		Active:            true,
		ManualReviewed:    true,
		Status:            models.CHARTSTATUS_NEW,
		EnvOverrideValues: template,
		IsBasicViewLocked: deploymentTemplateOverride.IsBasicViewLocked,
		CurrentViewEditor: deploymentTemplateOverride.CurrentViewEditor,
		ChartRefId:        chartRefId,
		EnvironmentId:     envId,
		UserId:            userId,
	}

	// if chart not found for chart_ref then create
	chartEntry, err := handler.chartRepo.FindChartByAppIdAndRefId(appId, chartRefId)
	if err != nil {
		if pg.ErrNoRows == err {
			templateRequest := bean3.TemplateRequest{
				AppId:               appId,
				ChartRefId:          chartRefId,
				ValuesOverride:      util.GetEmptyJSON(),
				UserId:              userId,
				IsAppMetricsEnabled: deploymentTemplateOverride.ShowAppMetrics,
			}
			newChartEntry, err := handler.chartService.CreateChartFromEnvOverride(context.Background(), templateRequest)
			if err != nil {
				handler.logger.Errorw("service err, CreateChartFromEnvOverride", "err", err, "appId", appId, "envId", envId, "chartRefId", chartRefId)
				return err
			}
			chartEntry.Id = newChartEntry.Id
			chartEntry.AppId = newChartEntry.AppId
		} else {
			handler.logger.Errorw("service err, FindChartByAppIdAndRefId", "err", err, "appId", appId, "envId", envId, "chartRefId", chartRefId)
			return err
		}
	}

	// create if required
	appMetrics := false
	if envConfigProperties.AppMetrics != nil {
		appMetrics = *envConfigProperties.AppMetrics
	}
	chartEntry.GlobalOverride = string(envConfigProperties.EnvOverrideValues)

	overrideCreateRequest := &bean2.EnvironmentOverrideCreateInternalDTO{
		Chart:               chartEntry,
		EnvironmentId:       envConfigProperties.EnvironmentId,
		UserId:              envConfigProperties.UserId,
		ManualReviewed:      envConfigProperties.ManualReviewed,
		ChartStatus:         models.CHARTSTATUS_SUCCESS,
		IsOverride:          true,
		IsAppMetricsEnabled: appMetrics,
		IsBasicViewLocked:   envConfigProperties.IsBasicViewLocked,
		Namespace:           envConfigProperties.Namespace,
		CurrentViewEditor:   envConfigProperties.CurrentViewEditor,
		MergeStrategy:       envConfigProperties.MergeStrategy,
	}

	_, updatedAppMetrics, err := handler.propertiesConfigService.CreateIfRequired(overrideCreateRequest, nil)
	if err != nil {
		handler.logger.Errorw("service err, CreateIfRequired", "err", err, "appId", appId, "envId", envId, "chartRefId", chartRefId)
		return err
	}
	envConfigProperties.AppMetrics = &updatedAppMetrics

	//getting environment properties for db table id(this properties get created when cd pipeline is created)
	env, err := handler.propertiesConfigService.GetEnvironmentProperties(appId, envId, deploymentTemplateOverride.ChartRefId)
	if err != nil {
		handler.logger.Errorw("service err, GetEnvConfOverride", "err", err, "appId", appId, "envId", envId, "chartRefId", deploymentTemplateOverride.ChartRefId)
		return err
	}

	//updating env template override
	envConfigProperties.Id = env.EnvironmentConfig.Id
	envConfigProperties.Namespace = env.Namespace
	_, err = handler.propertiesConfigService.UpdateEnvironmentProperties(appId, envConfigProperties, userId)
	if err != nil {
		handler.logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "appId", appId, "envId", envId)
		return err
	}
	return nil
}

// create CM overrides
func (handler CoreAppRestHandlerImpl) createEnvCM(appId int, userId int32, envId int, CmOverrides []*appBean.ConfigMap) error {
	handler.logger.Infow("Create App - creating CM override", "appId", appId, "envId", envId)

	var envLevelId int

	for _, cmOverride := range CmOverrides {
		//getting env level by app id and envId
		if envLevelId == 0 {
			envLevel, err := handler.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("error in getting app level by app id in createEnvCM", "appId", appId, "envId", envId)
				return err
			}
			if envLevel != nil {
				envLevelId = envLevel.Id
			}
		}

		cmOverrideData, err := json.Marshal(cmOverride.Data)
		if err != nil {
			handler.logger.Errorw("service err, could not json marshal template in CreateEnvCM", "err", err, "appId", appId, "envId", envId)
			return err
		}

		// build
		configData := &bean2.ConfigData{
			Name:     cmOverride.Name,
			External: cmOverride.IsExternal,
			Type:     cmOverride.UsageType,
			Data:     json.RawMessage(cmOverrideData),
		}
		cmOverrideDataVolumeUsageConfig := cmOverride.DataVolumeUsageConfig
		if cmOverrideDataVolumeUsageConfig != nil {
			configData.MountPath = cmOverrideDataVolumeUsageConfig.MountPath
			configData.SubPath = cmOverrideDataVolumeUsageConfig.SubPath
			configData.FilePermission = cmOverrideDataVolumeUsageConfig.FilePermission
		}

		var configDataRequest []*bean2.ConfigData
		configDataRequest = append(configDataRequest, configData)

		// service call
		cmEnvRequest := &bean2.ConfigDataRequest{
			AppId:         appId,
			UserId:        userId,
			EnvironmentId: envId,
			Id:            envLevelId,
			ConfigData:    configDataRequest,
		}

		_, err = handler.configMapService.CMEnvironmentAddUpdate(cmEnvRequest)
		if err != nil {
			handler.logger.Errorw("service err, CMEnvironmentAddUpdate in CreateEnvCM", "err", err, "payload", cmEnvRequest)
			return err
		}
	}

	return nil
}

// create secret overrides
func (handler CoreAppRestHandlerImpl) createEnvSecret(appId int, userId int32, envId int, secretOverrides []*appBean.Secret) error {
	handler.logger.Infow("Create App - creating secret overrides", "appId", appId)

	var envLevelId int
	for _, secretOverride := range secretOverrides {
		//getting env level by app id
		if envLevelId == 0 {
			envLevel, err := handler.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("error in getting app level by app id in createEnvSecret", "appId", appId, "envId", envId)
				return err
			}
			if envLevel != nil {
				envLevelId = envLevel.Id
			}
		}

		// build
		secretOverrideData, err := json.Marshal(secretOverride.Data)
		if err != nil {
			handler.logger.Errorw("service err, could not json marshal secret data in CreateEnvSecret", "err", err, "appId", appId, "envId", envId)
			return err
		}

		secretData := &bean2.ConfigData{
			Name:               secretOverride.Name,
			External:           secretOverride.IsExternal,
			ExternalSecretType: secretOverride.ExternalType,
			Type:               secretOverride.UsageType,
			Data:               secretOverrideData,
			RoleARN:            secretOverride.RoleArn,
			ExternalSecret:     convertCSExternalSecretData(secretOverride.ExternalSecretData),
		}
		secretOverrideDataVolumeUsageConfig := secretOverride.DataVolumeUsageConfig
		if secretOverrideDataVolumeUsageConfig != nil {
			secretData.MountPath = secretOverrideDataVolumeUsageConfig.MountPath
			secretData.SubPath = secretOverrideDataVolumeUsageConfig.SubPath
			secretData.FilePermission = secretOverrideDataVolumeUsageConfig.FilePermission
			secretData.ESOSubPath = secretOverrideDataVolumeUsageConfig.ESOSubPath
		}
		var secretDataRequest []*bean2.ConfigData
		secretDataRequest = append(secretDataRequest, secretData)

		// service call
		secretEnvRequest := &bean2.ConfigDataRequest{
			AppId:         appId,
			UserId:        userId,
			EnvironmentId: envId,
			Id:            envLevelId,
			ConfigData:    secretDataRequest,
		}
		_, err = handler.configMapService.CSEnvironmentAddUpdate(secretEnvRequest)
		if err != nil {
			handler.logger.Errorw("service err, CSEnvironmentAddUpdate", "err", err, "appId", appId, "envId", envId)
			return err
		}
	}

	return nil
}

//Create App related methods ends

//private methods for data conversion below

func convertCSExternalSecretData(externalSecretsData []*appBean.ExternalSecret) []bean2.ExternalSecret {
	var convertedExternalSecretsData []bean2.ExternalSecret
	for _, externalSecretData := range externalSecretsData {
		convertedExternalSecret := bean2.ExternalSecret{
			Key:      externalSecretData.Key,
			Name:     externalSecretData.Name,
			Property: externalSecretData.Property,
			IsBinary: externalSecretData.IsBinary,
		}
		convertedExternalSecretsData = append(convertedExternalSecretsData, convertedExternalSecret)
	}
	return convertedExternalSecretsData
}

func convertCiBuildScripts(buildScripts []*appBean.BuildScript) []*bean.CiScript {
	var convertedBuildScripts []*bean.CiScript
	for _, buildScript := range buildScripts {
		convertedBuildScript := &bean.CiScript{
			Index:          buildScript.Index,
			Name:           buildScript.Name,
			Script:         buildScript.Script,
			OutputLocation: buildScript.ReportDirectoryPath,
		}
		convertedBuildScripts = append(convertedBuildScripts, convertedBuildScript)
	}
	return convertedBuildScripts
}

func convertCdStages(cdStage *appBean.CdStage) bean.CdStage {

	convertedCdStage := bean.CdStage{}

	if cdStage != nil {
		convertedCdStage.TriggerType = cdStage.TriggerType
		convertedCdStage.Name = cdStage.Name
		convertedCdStage.Config = cdStage.Config
	}

	return convertedCdStage
}

func convertCdPreStageCMorCSNames(preStageNames *appBean.CdStageConfigMapSecretNames) bean.PreStageConfigMapSecretNames {

	convertPreStageNames := bean.PreStageConfigMapSecretNames{}
	if preStageNames != nil {
		convertPreStageNames.ConfigMaps = preStageNames.ConfigMaps
		convertPreStageNames.Secrets = preStageNames.Secrets
	}

	return convertPreStageNames
}

func convertCdPostStageCMorCSNames(postStageNames *appBean.CdStageConfigMapSecretNames) bean.PostStageConfigMapSecretNames {
	convertPostStageNames := bean.PostStageConfigMapSecretNames{}
	if postStageNames != nil {
		convertPostStageNames.ConfigMaps = postStageNames.ConfigMaps
		convertPostStageNames.Secrets = postStageNames.Secrets
	}

	return convertPostStageNames
}

func convertCdDeploymentStrategies(deploymentStrategies []*appBean.DeploymentStrategy) ([]bean.Strategy, error) {
	var convertedStrategies []bean.Strategy
	for _, deploymentStrategy := range deploymentStrategies {
		convertedStrategy := bean.Strategy{
			DeploymentTemplate: deploymentStrategy.DeploymentStrategyType,
			Default:            deploymentStrategy.IsDefault,
		}
		strategyConfig, err := json.Marshal(deploymentStrategy.Config)
		if err != nil {
			return nil, err
		}
		convertedStrategy.Config = strategyConfig
		convertedStrategies = append(convertedStrategies, convertedStrategy)
	}
	return convertedStrategies, nil
}

func (handler CoreAppRestHandlerImpl) validateCdPipelines(cdPipelines []*appBean.CdPipelineDetails, appName, token string) (error, int) {
	for _, cdPipeline := range cdPipelines {
		envName := cdPipeline.EnvironmentName
		envModel, err := handler.environmentRepository.FindByName(envName)
		if err != nil || envModel == nil {
			return fmt.Errorf("invalid environment name %s for cd pipeline", envName), http.StatusBadRequest
		}
		if envModel.IsVirtualEnvironment {
			return fmt.Errorf("virtual environment '%s' for cd pipeline is not supported yet", envName), http.StatusUnprocessableEntity
		}
		// validation RBAC starts
		object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(appName, envModel.Id)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
			return fmt.Errorf("unauthorized user for the environment %s", envName), http.StatusForbidden
		}
		// validation RBAC ends

		// validate deploymentStrategy for each cd pipeline
		for _, deploymentStrategy := range cdPipeline.DeploymentStrategies {
			_, err = json.Marshal(deploymentStrategy.Config)
			if err != nil {
				return fmt.Errorf("err in converting deployment strategies for creating the cd pipeline %s", cdPipeline.Name), http.StatusBadRequest
			}
		}
	}
	return nil, http.StatusOK
}

func (handler CoreAppRestHandlerImpl) ValidateAppWorkflowRequest(createAppWorkflowRequest *appBean.AppWorkflowCloneDto, token string) (error, int) {
	// validation for app workflow request
	if createAppWorkflowRequest.AppWorkflows != nil {
		for _, workflow := range createAppWorkflowRequest.AppWorkflows {
			// if ci pipeline is of external type, then throw error as we are not supporting it as of now
			if workflow.CiPipeline.ParentCiPipeline == 0 && workflow.CiPipeline.ParentAppId == 0 && workflow.CiPipeline.IsExternal {
				return fmt.Errorf("external ci pipeline creation is not supported yet"), http.StatusUnprocessableEntity
			}
			// for linked CI, parentCiPipeline should be a valid ciPipelineId and the given ParentAppId should be matched
			if workflow.CiPipeline.ParentCiPipeline != 0 && workflow.CiPipeline.ParentAppId != 0 {
				if !workflow.CiPipeline.IsExternal {
					return fmt.Errorf("invalid value of isExternal '%v', for linked CI isExternal must be TRUE", workflow.CiPipeline.IsExternal), http.StatusBadRequest
				}
				ciPipeline, err := handler.ciPipelineRepository.FindById(workflow.CiPipeline.ParentCiPipeline)
				if err != nil {
					return fmt.Errorf("error in finding ci pipeline with the given parentCiPipeline '%v'", workflow.CiPipeline.ParentCiPipeline), http.StatusBadRequest
				}
				if ciPipeline.AppId != workflow.CiPipeline.ParentAppId {
					return fmt.Errorf("invalid parentAppId '%v' for the given parentCiPipeline '%v'", workflow.CiPipeline.ParentAppId, workflow.CiPipeline.ParentCiPipeline), http.StatusBadRequest
				}
				parentMaterialMap := make(map[int]*pipelineConfig.CiPipelineMaterial)
				for _, material := range ciPipeline.CiPipelineMaterials {
					parentMaterialMap[material.GitMaterialId] = material
				}
				for _, requestPipelineMaterial := range workflow.CiPipeline.CiPipelineMaterialsConfig {
					parentMaterial, ok := parentMaterialMap[requestPipelineMaterial.GitMaterialId]
					if !ok {
						return fmt.Errorf("invalid material id - request material id should match parent material id for linked ci,  request material id - '%v' ", requestPipelineMaterial.GitMaterialId), http.StatusBadRequest
					}
					if requestPipelineMaterial.Value != parentMaterial.Value {
						return fmt.Errorf(" parentMaterialValue and request material value should match for linked ci - parent material value - %v child value %v ", requestPipelineMaterial.Value), http.StatusBadRequest
					}
					if requestPipelineMaterial.Type != parentMaterial.Type {
						return fmt.Errorf(" parentMaterialType and request material value should match for linked ci - parent material type - %v child type %v ", requestPipelineMaterial.Type), http.StatusBadRequest
					}
					if requestPipelineMaterial.CheckoutPath != parentMaterial.CheckoutPath {
						return fmt.Errorf(" parentMaterialType and request material CheckoutPath should match for linked ci - parent material CheckoutPath - %v child CheckoutPath %v ", requestPipelineMaterial.CheckoutPath), http.StatusBadRequest
					}
				}
			}
			ciMaterialCheckoutPaths := make([]string, 0)
			for _, ciPipelineMaterialConfig := range workflow.CiPipeline.CiPipelineMaterialsConfig {
				// value for webhook type CiPipelineMaterial should be a valid json string
				if ciPipelineMaterialConfig.Type == constants.SOURCE_TYPE_WEBHOOK {
					var jsonValueMap map[string]interface{}
					err := json.Unmarshal([]byte(ciPipelineMaterialConfig.Value), &jsonValueMap)
					if err != nil {
						return fmt.Errorf("invalid value for the ciPipelineMaterialsConfig type %s", string(ciPipelineMaterialConfig.Type)), http.StatusBadRequest
					}
					if _, ok := jsonValueMap["eventId"]; !ok {
						return fmt.Errorf("invalid value for the ciPipelineMaterialsConfig type %s, eventId not found", string(ciPipelineMaterialConfig.Type)), http.StatusBadRequest
					}

					if _, ok := jsonValueMap["condition"]; !ok {
						return fmt.Errorf("invalid value for the ciPipelineMaterialsConfig type %s, condition not found", string(ciPipelineMaterialConfig.Type)), http.StatusBadRequest
					}
				}
				// CiPipelineMaterial checkout paths should be unique
				if slices.Contains(ciMaterialCheckoutPaths, ciPipelineMaterialConfig.CheckoutPath) {
					return fmt.Errorf(""), http.StatusBadRequest
				}
				ciMaterialCheckoutPaths = append(ciMaterialCheckoutPaths, ciPipelineMaterialConfig.CheckoutPath)
			}
			// validate environment name and rbac object of payload
			if workflow.CdPipelines != nil {
				err, statusCode := handler.validateCdPipelines(workflow.CdPipelines, createAppWorkflowRequest.AppName, token)
				if err != nil {
					return err, statusCode
				}
			}
		}
	}
	// validation for environment override request
	if createAppWorkflowRequest.EnvironmentOverrides != nil {
		for envName, _ := range createAppWorkflowRequest.EnvironmentOverrides {
			envModel, err := handler.environmentRepository.FindByName(envName)

			if err != nil || envModel == nil {
				return fmt.Errorf("invalid environment name '%s' for environment override", envName), http.StatusBadRequest
			}
			if envModel.IsVirtualEnvironment {
				return fmt.Errorf("virtual environment '%s' for cd pipeline is not supported yet", envName), http.StatusUnprocessableEntity
			}
			// validate RBAC starts
			object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(createAppWorkflowRequest.AppName, envModel.Id)
			if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
				return fmt.Errorf("unauthorized user for the environment '%s'", envName), http.StatusForbidden
			}
			// validate RBAC ends
		}
	}
	return nil, http.StatusOK
}

func (handler CoreAppRestHandlerImpl) CreateAppWorkflow(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")

	ctx := r.Context()
	var createAppRequest appBean.AppWorkflowCloneDto
	err = decoder.Decode(&createAppRequest)
	if err != nil {
		handler.logger.Errorw("request err, CreateApp by API", "err", err, "CreateApp", createAppRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//to add more validations here
	handler.logger.Infow("request payload, CreateApp by API", "CreateApp", createAppRequest)
	err = handler.validator.Struct(createAppRequest)
	if err != nil {
		handler.logger.Errorw("validation err, CreateApp by API", "err", err, "CreateApp", createAppRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.appCrudOperationService.GetAppMetaInfoByAppName(createAppRequest.AppName)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo in GetAppAllDetail", "err", err, "appName", createAppRequest.AppName)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	createAppRequest.AppId = app.AppId
	object := fmt.Sprintf("%s/%s", app.ProjectName, app.AppName)
	// with admin roles, you have to access for all the apps of the project to create new app. (admin or manager with specific app permission can't create app.)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac ends

	// validate payload starts
	err, statusCode := handler.ValidateAppWorkflowRequest(&createAppRequest, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	// validate payload ends

	handler.logger.Infow("creating app workflow created ", "createAppRequest", createAppRequest)

	//creating workflow starts
	if createAppRequest.AppWorkflows != nil {
		if len(createAppRequest.AppWorkflows) != 1 {
			common.WriteJsonResp(w, err, "please provide only one workflow at one time", http.StatusBadRequest)
			return
		}
		err, statusCode = handler.createWorkflows(ctx, createAppRequest.AppId, userId, createAppRequest.AppWorkflows)
		if err != nil {
			common.WriteJsonResp(w, err, nil, statusCode)
			return
		}
	}
	//creating workflow ends

	//creating environment override starts
	if createAppRequest.EnvironmentOverrides != nil && len(createAppRequest.EnvironmentOverrides) > 0 {
		err, statusCode = handler.createEnvOverrides(ctx, createAppRequest.AppId, userId, createAppRequest.EnvironmentOverrides)
		if err != nil {
			common.WriteJsonResp(w, err, nil, statusCode)
			return
		}
	}
	//creating environment override ends

	common.WriteJsonResp(w, nil, APP_WORKFLOW_CREATE_SUCCESSFUL_RESP, http.StatusOK)
}

func (handler CoreAppRestHandlerImpl) GetAppWorkflow(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppWorkflow", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")

	// get app metadata for appId
	appMetaInfo, err := handler.appCrudOperationService.GetAppMetaInfo(appId, app.ZERO_INSTALLED_APP_ID, app.ZERO_ENVIRONMENT_ID)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo in GetAppWorkflow", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//get/build app workflows starts
	//using empty workflow name because it is optional, if not provided then workflows will be fetched on the basis of app
	wfCloneRequest := &appWorkflowBean.WorkflowCloneRequest{AppId: appId}
	appWorkflows, err, statusCode := handler.buildAppWorkflows(wfCloneRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build app workflows ends

	//get/build environment override starts
	environmentOverrides, err, statusCode := handler.buildEnvironmentOverrides(r.Context(), appId, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build environment override ends

	//build full object for response
	appDetail := &appBean.AppWorkflowCloneDto{
		AppId:                appId,
		AppName:              appMetaInfo.AppName,
		AppWorkflows:         appWorkflows,
		EnvironmentOverrides: environmentOverrides,
	}
	//end

	common.WriteJsonResp(w, nil, appDetail, http.StatusOK)
}

func (handler CoreAppRestHandlerImpl) GetAppWorkflowAndOverridesSample(w http.ResponseWriter, r *http.Request) {

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppWorkflow", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.appCrudOperationService.GetAppMetaInfo(appId, app.ZERO_INSTALLED_APP_ID, app.ZERO_ENVIRONMENT_ID)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo in GetAppAllDetail", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	wfCloneRequest := &appWorkflowBean.WorkflowCloneRequest{AppId: appId}
	workflowName := r.URL.Query().Get("workflowName")
	wfCloneRequest.WorkflowName = workflowName
	environmentIdStr := r.URL.Query().Get("environmentId")
	if len(environmentIdStr) > 0 {
		environmentId, err := strconv.Atoi(environmentIdStr)
		if err != nil {
			handler.logger.Errorw("error on GetAppWorkflowAndOverridesSample", "err", err)
			common.WriteJsonResp(w, err, "invalid environmentId", http.StatusBadRequest)
			return
		}
		wfCloneRequest.EnvironmentId = environmentId
	}
	workflowIdStr := r.URL.Query().Get("workflowId")
	if len(workflowIdStr) > 0 {
		workflowId, err := strconv.Atoi(workflowIdStr)
		if err != nil {
			handler.logger.Errorw("error on GetAppWorkflowAndOverridesSample", "err", err)
			common.WriteJsonResp(w, err, "invalid workflowId", http.StatusBadRequest)
			return
		}
		wfCloneRequest.WorkflowId = workflowId
	}
	token := r.Header.Get("token")
	//get/build app workflows starts
	appWorkflows, err, statusCode := handler.buildAppWorkflows(wfCloneRequest)
	if err != nil {
		handler.logger.Errorw("error on GetAppWorkflowAndOverridesSample", "err", err)
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build app workflows ends

	//build full object for response
	appDetail := &appBean.AppWorkflowCloneDto{
		AppId:   appId,
		AppName: app.AppName,
	}
	if appWorkflows != nil && len(appWorkflows) > 0 {
		aw := make([]*appBean.AppWorkflow, 0)
		aw = append(aw, appWorkflows[0])
		appDetail.AppWorkflows = aw
	}

	//get/build environment override starts
	environmentOverrides := make(map[string]*appBean.EnvironmentOverride)
	if wfCloneRequest.EnvironmentId > 0 {
		environmentOverrides, err, _ = handler.buildEnvironmentOverride(appId, wfCloneRequest.EnvironmentId, token)
	} else {
		environmentOverrides, err, _ = handler.buildEnvironmentOverrides(r.Context(), appId, token)
	}
	if err != nil {
		handler.logger.Errorw("error on GetAppWorkflowAndOverridesSample", "err", err)
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	//get/build environment override ends
	if environmentOverrides != nil && len(environmentOverrides) > 0 {
		eo := make(map[string]*appBean.EnvironmentOverride)
		for k, v := range environmentOverrides {
			eo[k] = v
			break
		}
		appDetail.EnvironmentOverrides = eo
	}
	//end

	common.WriteJsonResp(w, nil, appDetail, http.StatusOK)
}
