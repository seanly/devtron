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

package appStoreDiscoverRepository

import (
	"fmt"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
)

type AppStoreApplicationVersionRepository interface {
	FindWithFilter(filter *appStoreBean.AppStoreFilter) ([]appStoreBean.AppStoreWithVersion, error)
	FindById(id int) (*AppStoreApplicationVersion, error)
	FindVersionsByAppStoreId(id int) ([]*AppStoreApplicationVersion, error)
	FindChartVersionByAppStoreId(id int) ([]*AppStoreApplicationVersion, error)
	FindByIds(ids []int) ([]*AppStoreApplicationVersion, error)
	GetChartInfoById(id int) (*AppStoreApplicationVersion, error)
	FindLatestVersionByAppStoreIdForChartRepo(id int) (int, error)
	FindLatestVersionByAppStoreIdForOCIRepo(id int) (int, error)
	SearchAppStoreChartByName(chartName string) ([]*appStoreBean.ChartRepoSearch, error)
}

type AppStoreApplicationVersionRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewAppStoreApplicationVersionRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *AppStoreApplicationVersionRepositoryImpl {
	return &AppStoreApplicationVersionRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type FilterQueryUpdateAction string

const (
	QUERY_COLUMN_UPDATE FilterQueryUpdateAction = "column"
	QUERY_JOIN_UPDTAE   FilterQueryUpdateAction = "join"
)

type AppStoreApplicationVersion struct {
	TableName   struct{}  `sql:"app_store_application_version" pg:",discard_unknown_columns"`
	Id          int       `sql:"id,pk"`
	Version     string    `sql:"version"`
	AppVersion  string    `sql:"app_version"`
	Created     time.Time `sql:"created"`
	Deprecated  bool      `sql:"deprecated"`
	Description string    `sql:"description"`
	Digest      string    `sql:"digest"`
	Icon        string    `sql:"icon"`
	Name        string    `sql:"name"`
	Source      string    `sql:"source"`
	Home        string    `sql:"home"`
	ValuesYaml  string    `sql:"values_yaml"`
	ChartYaml   string    `sql:"chart_yaml"`
	AppStoreId  int       `sql:"app_store_id"`
	sql.AuditLog
	RawValues        string `sql:"raw_values"`
	Readme           string `sql:"readme"`
	ValuesSchemaJson string `sql:"values_schema_json"`
	Notes            string `sql:"notes"`
	AppStore         *AppStore
}

func (a *AppStoreApplicationVersion) IsEmpty() bool {
	if a == nil {
		return true
	}
	return a.Id == 0
}

func (impl AppStoreApplicationVersionRepositoryImpl) GetChartInfoById(id int) (*AppStoreApplicationVersion, error) {
	var appStoreWithVersion AppStoreApplicationVersion
	err := impl.dbConnection.Model(&appStoreWithVersion).Column("readme", "values_schema_json", "notes", "id").
		Where("id= ?", id).Select()
	return &appStoreWithVersion, err
}

func updateFindWithFilterQuery(filter *appStoreBean.AppStoreFilter, updateAction FilterQueryUpdateAction) (string, []interface{}) {
	query := ""
	var queryParams []interface{}
	if updateAction == QUERY_COLUMN_UPDATE {
		if len(filter.ChartRepoId) > 0 && len(filter.RegistryId) > 0 {
			query = " ch.name as chart_name, das.id as docker_artifact_store_id"
		} else if len(filter.RegistryId) > 0 {
			query = " das.id as docker_artifact_store_id"
		} else if len(filter.ChartRepoId) > 0 {
			query = " ch.name as chart_name"
		} else {
			query = " ch.name as chart_name, das.id as docker_artifact_store_id"
		}
	}
	//for chart repos, created (derived through index.yaml) column of app_store_application_version is used for finding latest version and for oci repo id is used (because created is null)
	latestAppStoreVersionQueryForChartRepo := " SELECT MAX(created) as created " +
		" FROM app_store_application_version asv " +
		" INNER JOIN app_store aps ON (asv.app_store_id = aps.id and aps.active = true and aps.chart_repo_id is NOT NULL) " +
		" GROUP BY asv.app_store_id "

	latestAppStoreVersionQueryForOCIRepo := " SELECT MAX(asv.id) as id " +
		" FROM app_store_application_version asv " +
		" INNER JOIN app_store aps ON (asv.app_store_id = aps.id and aps.active = true and aps.docker_artifact_store_id is NOT NULL) " +
		" GROUP BY asv.app_store_id "

	combinedWhereClause := fmt.Sprintf("( (asv.created IN (%s) and aps.chart_repo_id is not null ) or (asv.id IN (%s) and aps.docker_artifact_store_id is not null) )", latestAppStoreVersionQueryForChartRepo, latestAppStoreVersionQueryForOCIRepo)

	if updateAction == QUERY_JOIN_UPDTAE {
		if len(filter.ChartRepoId) > 0 && len(filter.RegistryId) > 0 {
			query = " LEFT JOIN chart_repo ch ON (aps.chart_repo_id = ch.id and ch.deleted IS FALSE)" +
				" LEFT JOIN docker_artifact_store das ON aps.docker_artifact_store_id = das.id" +
				" LEFT JOIN oci_registry_config oci ON oci.docker_artifact_store_id = das.id" +
				fmt.Sprintf(" WHERE ( (%s) AND (ch.active IS TRUE OR (das.active IS TRUE AND oci.deleted IS FALSE AND oci.is_chart_pull_active IS TRUE)))", combinedWhereClause) +
				" AND (ch.id IN (?) OR das.id IN (?))"
			queryParams = append(queryParams, pg.In(filter.ChartRepoId), pg.In(filter.RegistryId))
		} else if len(filter.RegistryId) > 0 {
			query = " LEFT JOIN docker_artifact_store das ON aps.docker_artifact_store_id = das.id" +
				" LEFT JOIN oci_registry_config oci ON oci.docker_artifact_store_id = das.id" +
				fmt.Sprintf(" WHERE asv.id IN (%s) AND (das.active IS TRUE AND oci.deleted IS FALSE AND oci.is_chart_pull_active IS TRUE)", latestAppStoreVersionQueryForOCIRepo) +
				" AND das.id IN (?)"
			queryParams = append(queryParams, pg.In(filter.RegistryId))
		} else if len(filter.ChartRepoId) > 0 {
			query = " LEFT JOIN chart_repo ch ON (aps.chart_repo_id = ch.id and ch.deleted IS FALSE)" +
				fmt.Sprintf(" WHERE asv.created IN (%s) AND ch.active IS TRUE", latestAppStoreVersionQueryForChartRepo) +
				" AND ch.id IN (?)"
			queryParams = append(queryParams, pg.In(filter.ChartRepoId))
		} else {
			query = " LEFT JOIN chart_repo ch ON (aps.chart_repo_id = ch.id and ch.deleted IS FALSE)" +
				" LEFT JOIN docker_artifact_store das ON aps.docker_artifact_store_id = das.id" +
				" LEFT JOIN oci_registry_config oci ON oci.docker_artifact_store_id = das.id" +
				fmt.Sprintf(" WHERE (%s AND (ch.active IS TRUE OR (das.active IS TRUE AND oci.deleted IS FALSE AND oci.is_chart_pull_active IS TRUE)))", combinedWhereClause)
		}
	}
	return query, queryParams
}

func (impl *AppStoreApplicationVersionRepositoryImpl) FindWithFilter(filter *appStoreBean.AppStoreFilter) ([]appStoreBean.AppStoreWithVersion, error) {
	var appStoreWithVersion []appStoreBean.AppStoreWithVersion
	var queryParams []interface{}
	query := `SELECT asv.version, asv.icon, asv.deprecated, asv.id as app_store_application_version_id, 
			  asv.description, aps.*, `

	queryColumnUpdate, queryParamsColumnUpdate := updateFindWithFilterQuery(filter, QUERY_COLUMN_UPDATE)
	query += queryColumnUpdate
	queryParams = append(queryParams, queryParamsColumnUpdate...)

	query = query + " FROM app_store_application_version asv " +
		" INNER JOIN app_store aps ON (asv.app_store_id = aps.id and aps.active = ?) "
	queryParams = append(queryParams, "true")

	queryJoinUpdate, queryParamsJoinUpdate := updateFindWithFilterQuery(filter, QUERY_JOIN_UPDTAE)
	query += queryJoinUpdate
	queryParams = append(queryParams, queryParamsJoinUpdate...)

	if !filter.IncludeDeprecated {
		query = query + " AND asv.deprecated = ? "
		queryParams = append(queryParams, "FALSE")
	}
	if len(filter.AppStoreName) > 0 {
		query = query + " AND aps.name LIKE ? "
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(filter.AppStoreName))
	}
	query = query + " ORDER BY aps.name ASC "
	if filter.Size > 0 {
		query = query + " OFFSET ? LIMIT ? "
		queryParams = append(queryParams, filter.Offset, filter.Size)
	}
	query = query + ";"

	var err error
	if len(filter.ChartRepoId) > 0 && len(filter.RegistryId) > 0 {
		_, err = impl.dbConnection.Query(&appStoreWithVersion, query, queryParams...)
	} else if len(filter.RegistryId) > 0 {
		_, err = impl.dbConnection.Query(&appStoreWithVersion, query, queryParams...)
	} else if len(filter.ChartRepoId) > 0 {
		_, err = impl.dbConnection.Query(&appStoreWithVersion, query, queryParams...)
	} else {
		_, err = impl.dbConnection.Query(&appStoreWithVersion, query, queryParams...)
	}
	if err != nil {
		return nil, err
	}
	return appStoreWithVersion, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindById(id int) (*AppStoreApplicationVersion, error) {
	appStoreWithVersion := &AppStoreApplicationVersion{}
	err := impl.dbConnection.
		Model(appStoreWithVersion).
		Column("app_store_application_version.*", "AppStore", "AppStore.ChartRepo", "AppStore.DockerArtifactStore", "AppStore.DockerArtifactStore.OCIRegistryConfig").
		Join("INNER JOIN app_store aps on app_store_application_version.app_store_id = aps.id").
		Join("LEFT JOIN chart_repo ch on aps.chart_repo_id = ch.id").
		Join("LEFT JOIN docker_artifact_store das on (aps.docker_artifact_store_id = das.id and das.active IS TRUE)").
		Join("LEFT JOIN oci_registry_config orc on orc.docker_artifact_store_id=das.id").
		Relation("AppStore.DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Where("app_store_application_version.id = ?", id).
		Limit(1).
		Select()
	return appStoreWithVersion, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindByIds(ids []int) ([]*AppStoreApplicationVersion, error) {
	var appStoreApplicationVersions []*AppStoreApplicationVersion
	if len(ids) == 0 {
		return appStoreApplicationVersions, nil
	}
	err := impl.dbConnection.
		Model(&appStoreApplicationVersions).
		Column("app_store_application_version.*", "AppStore", "AppStore.ChartRepo", "AppStore.DockerArtifactStore").
		Where("app_store_application_version.id in (?)", pg.In(ids)).
		Join("INNER JOIN app_store aps on app_store_application_version.app_store_id = aps.id").
		Join("LEFT JOIN chart_repo ch on aps.chart_repo_id = ch.id").
		Join("LEFT JOIN docker_artifact_store das on (aps.docker_artifact_store_id = das.id and das.active IS TRUE)").
		Select()
	return appStoreApplicationVersions, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindChartVersionByAppStoreId(appStoreId int) ([]*AppStoreApplicationVersion, error) {
	var appStoreWithVersion []*AppStoreApplicationVersion
	err := impl.dbConnection.
		Model(&appStoreWithVersion).
		Column("app_store_application_version.version", "app_store_application_version.id").
		Where("app_store_application_version.app_store_id = ?", appStoreId).
		Select()
	return appStoreWithVersion, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindVersionsByAppStoreId(id int) ([]*AppStoreApplicationVersion, error) {
	var appStoreApplicationVersions []*AppStoreApplicationVersion
	err := impl.dbConnection.
		Model(&appStoreApplicationVersions).
		Column("app_store_application_version.id", "app_store_application_version.version").
		Where("app_store_id = ?", id).
		Order("created DESC").
		Select()
	return appStoreApplicationVersions, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) FindLatestVersionByAppStoreIdForChartRepo(id int) (int, error) {
	var appStoreApplicationVersionId int
	queryTemp := "SELECT asv.id AS app_store_application_version_id  FROM app_store_application_version AS asv  JOIN app_store AS ap ON asv.app_store_id = ap.id WHERE ap.id = ? order by created desc limit 1;"
	_, err := impl.dbConnection.Query(&appStoreApplicationVersionId, queryTemp, id)
	return appStoreApplicationVersionId, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) FindLatestVersionByAppStoreIdForOCIRepo(id int) (int, error) {
	var appStoreApplicationVersionId int
	queryTemp := "SELECT MAX(asv.id) AS app_store_application_version_id  FROM app_store_application_version AS asv  JOIN app_store AS ap ON asv.app_store_id = ap.id WHERE ap.id = ?;"
	_, err := impl.dbConnection.Query(&appStoreApplicationVersionId, queryTemp, id)
	return appStoreApplicationVersionId, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) SearchAppStoreChartByName(chartName string) ([]*appStoreBean.ChartRepoSearch, error) {
	var chartRepos []*appStoreBean.ChartRepoSearch
	//for chart repos, created (derived through index.yaml) column of app_store_application_version is used for finding latest version and for oci repo id is used (because created is null)
	queryTemp := `select asv.id as app_store_application_version_id, asv.version, asv.deprecated, aps.id as chart_id, 
					aps.name as chart_name, chr.id as chart_repo_id, chr.name as chart_repo_name , das.id as docker_artifact_store_id 
					from app_store_application_version asv 
					inner join app_store aps on asv.app_store_id = aps.id 
					left join chart_repo chr on aps.chart_repo_id = chr.id 
					left join docker_artifact_store das on aps.docker_artifact_store_id = das.id 
					where aps.name like ? and 
					( 
						( aps.docker_artifact_store_id is NOT NULL and asv.id = (SELECT MAX(id) FROM app_store_application_version WHERE app_store_id = asv.app_store_id)) 
						or 
						(aps.chart_repo_id is NOT NULL and  asv.created = (SELECT MAX(created) FROM app_store_application_version WHERE app_store_id = asv.app_store_id)) 
					) 
					and aps.active=? order by aps.name asc;`
	_, err := impl.dbConnection.Query(&chartRepos, queryTemp, util.GetLIKEClauseQueryParam(chartName), true)
	if err != nil {
		return nil, err
	}
	return chartRepos, err
}
