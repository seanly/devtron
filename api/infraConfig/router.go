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

package infraConfig

import "github.com/gorilla/mux"

type InfraConfigRouter interface {
	InitInfraConfigRouter(configRouter *mux.Router)
}

type InfraConfigRouterImpl struct {
	infraConfigRestHandler InfraConfigRestHandler
}

func NewInfraProfileRouterImpl(infraConfigRestHandler InfraConfigRestHandler) *InfraConfigRouterImpl {
	return &InfraConfigRouterImpl{
		infraConfigRestHandler: infraConfigRestHandler,
	}
}

func (impl *InfraConfigRouterImpl) InitInfraConfigRouter(configRouter *mux.Router) {
	configRouter.Path("/profile/alpha1").
		Queries("name", "{name}").
		HandlerFunc(impl.infraConfigRestHandler.GetProfile).
		Methods("GET")

	configRouter.Path("/profile/alpha1").
		Queries("name", "{name}").
		HandlerFunc(impl.infraConfigRestHandler.UpdateInfraProfile).
		Methods("PUT")

	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.GetProfileV0).
		Methods("GET")

	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.UpdateInfraProfileV0).
		Methods("PUT")
}
