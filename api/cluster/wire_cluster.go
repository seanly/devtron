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

package cluster

import (
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	read2 "github.com/devtron-labs/devtron/pkg/cluster/environment/read"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/rbac"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	repository2 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"github.com/google/wire"
)

//depends on sql,user,K8sUtil, logger, enforcer, TODO

var ClusterWireSet = wire.NewSet(
	repository.NewClusterRepositoryImpl,
	wire.Bind(new(repository.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	cluster.NewClusterServiceImpl,
	cluster.NewClusterServiceImplExtended,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImplExtended)),
	read.NewClusterReadServiceImpl,
	wire.Bind(new(read.ClusterReadService), new(*read.ClusterReadServiceImpl)),

	rbac.NewClusterRbacServiceImpl,
	wire.Bind(new(rbac.ClusterRbacService), new(*rbac.ClusterRbacServiceImpl)),

	repository.NewClusterDescriptionRepositoryImpl,
	wire.Bind(new(repository.ClusterDescriptionRepository), new(*repository.ClusterDescriptionRepositoryImpl)),
	repository2.NewGenericNoteHistoryRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteHistoryRepository), new(*repository2.GenericNoteHistoryRepositoryImpl)),
	repository2.NewGenericNoteRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteRepository), new(*repository2.GenericNoteRepositoryImpl)),
	genericNotes.NewGenericNoteHistoryServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteHistoryService), new(*genericNotes.GenericNoteHistoryServiceImpl)),
	genericNotes.NewGenericNoteServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteService), new(*genericNotes.GenericNoteServiceImpl)),
	cluster.NewClusterDescriptionServiceImpl,
	wire.Bind(new(cluster.ClusterDescriptionService), new(*cluster.ClusterDescriptionServiceImpl)),

	NewClusterRestHandlerImpl,
	wire.Bind(new(ClusterRestHandler), new(*ClusterRestHandlerImpl)),
	NewClusterRouterImpl,
	wire.Bind(new(ClusterRouter), new(*ClusterRouterImpl)),

	repository3.NewEnvironmentRepositoryImpl,
	wire.Bind(new(repository3.EnvironmentRepository), new(*repository3.EnvironmentRepositoryImpl)),
	environment.NewEnvironmentServiceImpl,
	wire.Bind(new(environment.EnvironmentService), new(*environment.EnvironmentServiceImpl)),
	read2.NewEnvironmentReadServiceImpl,
	wire.Bind(new(read2.EnvironmentReadService), new(*read2.EnvironmentReadServiceImpl)),
	NewEnvironmentRestHandlerImpl,
	wire.Bind(new(EnvironmentRestHandler), new(*EnvironmentRestHandlerImpl)),
	NewEnvironmentRouterImpl,
	wire.Bind(new(EnvironmentRouter), new(*EnvironmentRouterImpl)),
)

// minimal wire to be used with EA
var ClusterWireSetEa = wire.NewSet(
	repository.NewClusterRepositoryImpl,
	wire.Bind(new(repository.ClusterRepository), new(*repository.ClusterRepositoryImpl)),
	rbac.NewClusterRbacServiceImpl,
	wire.Bind(new(rbac.ClusterRbacService), new(*rbac.ClusterRbacServiceImpl)),
	cluster.NewClusterServiceImpl,
	wire.Bind(new(cluster.ClusterService), new(*cluster.ClusterServiceImpl)),
	read.NewClusterReadServiceImpl,
	wire.Bind(new(read.ClusterReadService), new(*read.ClusterReadServiceImpl)),

	repository.NewClusterDescriptionRepositoryImpl,
	wire.Bind(new(repository.ClusterDescriptionRepository), new(*repository.ClusterDescriptionRepositoryImpl)),
	repository2.NewGenericNoteHistoryRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteHistoryRepository), new(*repository2.GenericNoteHistoryRepositoryImpl)),
	repository2.NewGenericNoteRepositoryImpl,
	wire.Bind(new(repository2.GenericNoteRepository), new(*repository2.GenericNoteRepositoryImpl)),
	genericNotes.NewGenericNoteHistoryServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteHistoryService), new(*genericNotes.GenericNoteHistoryServiceImpl)),
	genericNotes.NewGenericNoteServiceImpl,
	wire.Bind(new(genericNotes.GenericNoteService), new(*genericNotes.GenericNoteServiceImpl)),
	cluster.NewClusterDescriptionServiceImpl,
	wire.Bind(new(cluster.ClusterDescriptionService), new(*cluster.ClusterDescriptionServiceImpl)),

	NewClusterRestHandlerImpl,
	wire.Bind(new(ClusterRestHandler), new(*ClusterRestHandlerImpl)),
	NewClusterRouterImpl,
	wire.Bind(new(ClusterRouter), new(*ClusterRouterImpl)),
	repository3.NewEnvironmentRepositoryImpl,
	wire.Bind(new(repository3.EnvironmentRepository), new(*repository3.EnvironmentRepositoryImpl)),
	environment.NewEnvironmentServiceImpl,
	wire.Bind(new(environment.EnvironmentService), new(*environment.EnvironmentServiceImpl)),
	read2.NewEnvironmentReadServiceImpl,
	wire.Bind(new(read2.EnvironmentReadService), new(*read2.EnvironmentReadServiceImpl)),
	NewEnvironmentRestHandlerImpl,
	wire.Bind(new(EnvironmentRestHandler), new(*EnvironmentRestHandlerImpl)),
	NewEnvironmentRouterImpl,
	wire.Bind(new(EnvironmentRouter), new(*EnvironmentRouterImpl)),
)
