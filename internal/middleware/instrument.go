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

package middleware

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strconv"
	"time"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "orchestrator_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "status"})
)

var CdDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "cd_duration_seconds",
	Help: "Duration of CD process",
}, []string{"appName", "status", "envName", "deploymentType"})

var GitOpsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "git_ops_duration_seconds",
	Help: "Duration of GitOps",
}, []string{"operationName", "methodName", "status"})

var CiDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "ci_duration_seconds",
	Help:    "Duration of CI process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var CacheDownloadDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "cache_download_duration_seconds",
	Help:    "Duration of Cache Download process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var PreCiDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "pre_ci_duration_seconds",
	Help:    "Duration of Pre CI process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var BuildDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "build_duration_seconds",
	Help:    "Duration of Build process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var PostCiDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "post_ci_duration_seconds",
	Help:    "Duration of Post CI process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var CacheUploadDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "cache_upload_duration_seconds",
	Help:    "Duration of Cache Upload process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var AppListingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "app_listing_duration_seconds",
	Help: "Duration of App Listing process",
}, []string{"MethodName", "AppType"})

var requestCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "orchestrator_http_requests_total",
		Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
	},
	[]string{"path", "method", "status"})

var currentRequestGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "orchestrator_http_requests_current",
	Help: "no of request being served currently",
}, []string{"path", "method"})

var CdTriggerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "cd_trigger_counter",
}, []string{"appName", "envName"})

var CiTriggerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ci_trigger_counter",
}, []string{"appName", "pipelineName"})

var DeploymentStatusCronDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "deployment_status_cron_process_time",
}, []string{"cronName"})

var TerminalSessionRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "initiate_terminal_session_request_counter",
	Help: "count of requests for initiated, established and closed terminal sessions",
}, []string{"sessionAction", "isError"})

var TerminalSessionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "terminal_session_duration",
	Help: "duration of each terminal session",
}, []string{"podName", "namespace", "clusterId"})

var ReTriggerFailedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "re_trigger_failed_counter",
	Help: "ci/ pre cd/ post cd workflow re-trigger failed counter",
}, []string{"workflowType", "workflowId"})

// prometheusMiddleware implements mux.MiddlewareFunc.
func PrometheusMiddleware(next http.Handler) http.Handler {
	//	prometheus.MustRegister(requestCounter)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		method := r.Method
		g := currentRequestGauge.WithLabelValues(path, method)
		g.Inc()
		defer g.Dec()
		d := NewDelegator(w, nil)
		next.ServeHTTP(d, r)
		httpDuration.WithLabelValues(path, method, strconv.Itoa(d.Status())).Observe(time.Since(start).Seconds())
		requestCounter.WithLabelValues(path, method, strconv.Itoa(d.Status())).Inc()
	})
}

func IncTerminalSessionRequestCounter(sessionAction string, isError string) {
	TerminalSessionRequestCounter.WithLabelValues(sessionAction, isError).Inc()
}

func RecordTerminalSessionDurationMetrics(podName, namespace, clusterId string, sessionDuration float64) {
	TerminalSessionDuration.WithLabelValues(podName, namespace, clusterId).Observe(sessionDuration)
}
