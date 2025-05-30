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

package grafana

import "github.com/caarlos0/env"

type Config struct {
	Host      string `env:"GRAFANA_HOST" envDefault:"localhost" description:"Host URL for the grafana dashboard"`
	Port      string `env:"GRAFANA_PORT" envDefault:"8090" description:"Port for grafana micro-service"`
	Namespace string `env:"GRAFANA_NAMESPACE" envDefault:"devtroncd" description:"Namespace for grafana"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
