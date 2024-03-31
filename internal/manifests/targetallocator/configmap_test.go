// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package targetallocator

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	"github.com/open-telemetry/opentelemetry-operator/internal/config"
)

func TestDesiredConfigMap(t *testing.T) {
	expectedLabels := map[string]string{
		"app.kubernetes.io/managed-by": "opentelemetry-operator",
		"app.kubernetes.io/instance":   "default.my-instance",
		"app.kubernetes.io/part-of":    "opentelemetry",
		"app.kubernetes.io/version":    "0.47.0",
	}
	collector := collectorInstance()
	targetAllocator := targetAllocatorInstance()
	cfg := config.New()
	params := Params{
		Collector:       &collector,
		TargetAllocator: targetAllocator,
		Config:          cfg,
		Log:             logr.Discard(),
	}

	t.Run("should return expected target allocator config map", func(t *testing.T) {
		expectedLabels["app.kubernetes.io/component"] = "opentelemetry-targetallocator"
		expectedLabels["app.kubernetes.io/name"] = "my-instance-targetallocator"

		expectedData := map[string]string{
			targetAllocatorFilename: `allocation_strategy: consistent-hashing
collector_selector:
  matchlabels:
    app.kubernetes.io/component: opentelemetry-collector
    app.kubernetes.io/instance: default.my-instance
    app.kubernetes.io/managed-by: opentelemetry-operator
    app.kubernetes.io/part-of: opentelemetry
  matchexpressions: []
config:
  scrape_configs:
  - job_name: otel-collector
    scrape_interval: 10s
    static_configs:
    - targets:
      - 0.0.0.0:8888
      - 0.0.0.0:9999
filter_strategy: relabel-config
`,
		}

		actual, err := ConfigMap(params)
		require.NoError(t, err)

		assert.Equal(t, "my-instance-targetallocator", actual.Name)
		assert.Equal(t, expectedLabels, actual.Labels)
		assert.Equal(t, expectedData[targetAllocatorFilename], actual.Data[targetAllocatorFilename])

	})
	t.Run("should return target allocator config map without scrape configs", func(t *testing.T) {
		expectedLabels["app.kubernetes.io/component"] = "opentelemetry-targetallocator"
		expectedLabels["app.kubernetes.io/name"] = "my-instance-targetallocator"

		expectedData := map[string]string{
			targetAllocatorFilename: `allocation_strategy: consistent-hashing
collector_selector:
  matchlabels:
    app.kubernetes.io/component: opentelemetry-collector
    app.kubernetes.io/instance: default.my-instance
    app.kubernetes.io/managed-by: opentelemetry-operator
    app.kubernetes.io/part-of: opentelemetry
  matchexpressions: []
filter_strategy: relabel-config
`,
		}
		targetAllocator = targetAllocatorInstance()
		targetAllocator.Spec.ScrapeConfigs = []v1beta1.AnyConfig{}
		collectorWithoutScrapeConfigs := collectorInstance()
		collectorWithoutScrapeConfigs.Spec.Config = v1beta1.Config{
			Receivers: v1beta1.AnyConfig{
				Object: map[string]any{
					"prometheus": map[string]any{
						"config": map[string]any{
							"scrape_configs": []any{},
						},
					},
				},
			},
		}
		params.TargetAllocator = targetAllocator
		params.Collector = &collectorWithoutScrapeConfigs
		actual, err := ConfigMap(params)
		require.NoError(t, err)

		assert.Equal(t, "my-instance-targetallocator", actual.Name)
		assert.Equal(t, expectedLabels, actual.Labels)
		assert.Equal(t, expectedData[targetAllocatorFilename], actual.Data[targetAllocatorFilename])

	})
	t.Run("should return expected target allocator config map with label selectors", func(t *testing.T) {
		expectedLabels["app.kubernetes.io/component"] = "opentelemetry-targetallocator"
		expectedLabels["app.kubernetes.io/name"] = "my-instance-targetallocator"

		expectedData := map[string]string{
			targetAllocatorFilename: `allocation_strategy: consistent-hashing
collector_selector:
  matchlabels:
    app.kubernetes.io/component: opentelemetry-collector
    app.kubernetes.io/instance: default.my-instance
    app.kubernetes.io/managed-by: opentelemetry-operator
    app.kubernetes.io/part-of: opentelemetry
  matchexpressions: []
config:
  scrape_configs:
  - job_name: otel-collector
    scrape_interval: 10s
    static_configs:
    - targets:
      - 0.0.0.0:8888
      - 0.0.0.0:9999
filter_strategy: relabel-config
prometheus_cr:
  enabled: true
  pod_monitor_selector:
    matchlabels:
      release: my-instance
    matchexpressions: []
  service_monitor_selector:
    matchlabels:
      release: my-instance
    matchexpressions: []
`,
		}
		targetAllocator = targetAllocatorInstance()
		targetAllocator.Spec.PrometheusCR.Enabled = true
		targetAllocator.Spec.PrometheusCR.PodMonitorSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"release": "my-instance",
			},
		}
		targetAllocator.Spec.PrometheusCR.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"release": "my-instance",
			}}
		params.TargetAllocator = targetAllocator
		params.Collector = &collector
		actual, err := ConfigMap(params)
		assert.NoError(t, err)

		assert.Equal(t, "my-instance-targetallocator", actual.Name)
		assert.Equal(t, expectedLabels, actual.Labels)
		assert.Equal(t, expectedData, actual.Data)

	})
	t.Run("should return expected target allocator config map with scrape interval set", func(t *testing.T) {
		expectedLabels["app.kubernetes.io/component"] = "opentelemetry-targetallocator"
		expectedLabels["app.kubernetes.io/name"] = "my-instance-targetallocator"

		expectedData := map[string]string{
			targetAllocatorFilename: `allocation_strategy: consistent-hashing
collector_selector:
  matchlabels:
    app.kubernetes.io/component: opentelemetry-collector
    app.kubernetes.io/instance: default.my-instance
    app.kubernetes.io/managed-by: opentelemetry-operator
    app.kubernetes.io/part-of: opentelemetry
  matchexpressions: []
config:
  scrape_configs:
  - job_name: otel-collector
    scrape_interval: 10s
    static_configs:
    - targets:
      - 0.0.0.0:8888
      - 0.0.0.0:9999
filter_strategy: relabel-config
prometheus_cr:
  enabled: true
  pod_monitor_selector: null
  scrape_interval: 30s
  service_monitor_selector: null
`,
		}

		targetAllocator = targetAllocatorInstance()
		targetAllocator.Spec.PrometheusCR.Enabled = true
		targetAllocator.Spec.PrometheusCR.ScrapeInterval = &metav1.Duration{Duration: time.Second * 30}
		params.TargetAllocator = targetAllocator
		params.Collector = &collector
		actual, err := ConfigMap(params)
		assert.NoError(t, err)

		assert.Equal(t, "my-instance-targetallocator", actual.Name)
		assert.Equal(t, expectedLabels, actual.Labels)
		assert.Equal(t, expectedData, actual.Data)

	})

}

func TestGetScrapeConfigs(t *testing.T) {
	testCases := []struct {
		name    string
		input   v1beta1.Config
		want    []v1beta1.AnyConfig
		wantErr error
	}{
		{
			name: "empty scrape configs list",
			input: v1beta1.Config{
				Receivers: v1beta1.AnyConfig{
					Object: map[string]interface{}{
						"prometheus": map[any]any{
							"config": map[any]any{
								"scrape_configs": []any{},
							},
						},
					},
				},
			},
			want: []v1beta1.AnyConfig{},
		},
		{
			name: "no scrape configs key",
			input: v1beta1.Config{
				Receivers: v1beta1.AnyConfig{
					Object: map[string]interface{}{
						"prometheus": map[any]any{
							"config": map[any]any{},
						},
					},
				},
			},
			wantErr: fmt.Errorf("no scrape_configs available as part of the configuration"),
		},
		{
			name: "one scrape config",
			input: v1beta1.Config{
				Receivers: v1beta1.AnyConfig{
					Object: map[string]interface{}{
						"prometheus": map[string]any{
							"config": map[string]any{
								"scrape_configs": []any{
									map[string]any{
										"job_name": "otel-collector",
										"relabel_configs": []any{
											map[string]any{
												"action": "labelmap",
												"regex":  "__meta_kubernetes_node_label_(.+)",
											},
											map[string]any{
												"action":      "labeldrop",
												"replacement": "$1",
											},
											map[string]any{
												"action":      "labelkeep",
												"regex":       "metrica_*|metricb.*",
												"replacement": "$1",
											},
										},
										"scrape_interval": "10s",
										"static_configs": []any{
											map[string]any{
												"targets": []any{"0.0.0.0:8888"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []v1beta1.AnyConfig{
				{Object: map[string]any{
					"job_name": "otel-collector",
					"relabel_configs": []any{
						map[any]any{
							"action": "labelmap",
							"regex":  "__meta_kubernetes_node_label_(.+)",
						},
						map[any]any{
							"action":      "labeldrop",
							"replacement": "$1",
						},
						map[any]any{
							"action":      "labelkeep",
							"regex":       "metrica_*|metricb.*",
							"replacement": "$1",
						},
					},
					"scrape_interval": "10s",
					"static_configs": []any{
						map[any]any{
							"targets": []any{"0.0.0.0:8888"},
						},
					},
				}},
			},
		},
		{
			name: "regex substitution",
			input: v1beta1.Config{
				Receivers: v1beta1.AnyConfig{
					Object: map[string]interface{}{
						"prometheus": map[string]any{
							"config": map[string]any{
								"scrape_configs": []any{
									map[string]any{
										"job": "somejob",
										"metric_relabel_configs": []map[string]any{
											{
												"action":      "labelmap",
												"regex":       "label_(.+)",
												"replacement": "$$1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []v1beta1.AnyConfig{
				{Object: map[string]interface{}{
					"job": "somejob",
					"metric_relabel_configs": []any{
						map[any]any{
							"action":      "labelmap",
							"regex":       "label_(.+)",
							"replacement": "$1",
						},
					},
				}},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			configStr, err := testCase.input.Yaml()
			require.NoError(t, err)
			actual, err := getScrapeConfigs(configStr)
			assert.Equal(t, testCase.wantErr, err)
			assert.Equal(t, testCase.want, actual)
		})
	}
}
