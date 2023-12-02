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

package collector

import (
	"github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	"github.com/open-telemetry/opentelemetry-operator/apis/v1alpha2"
	"github.com/open-telemetry/opentelemetry-operator/internal/manifests"
	"github.com/open-telemetry/opentelemetry-operator/internal/manifests/manifestutils"
	"github.com/open-telemetry/opentelemetry-operator/internal/manifests/targetallocator/adapters"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TargetAllocator builds the TargetAllocator CR for the given instance.
func TargetAllocator(params manifests.Params) (*v1alpha2.TargetAllocator, error) {

	taSpec := params.OtelCol.Spec.TargetAllocator
	if !taSpec.Enabled {
		return nil, nil
	}

	name := params.OtelCol.Name
	collectorSelector := metav1.LabelSelector{
		MatchLabels: manifestutils.SelectorLabels(params.OtelCol.ObjectMeta, ComponentOpenTelemetryCollector),
	}
	scrapeConfigs, err := getScrapeConfigs(params.OtelCol.Spec.Config)
	if err != nil {
		return nil, err
	}

	return &v1alpha2.TargetAllocator{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   params.OtelCol.Namespace,
			Annotations: params.OtelCol.Annotations,
			Labels:      params.OtelCol.Labels,
		},
		Spec: v1alpha2.TargetAllocatorSpec{
			OpenTelemetryCommonFields: v1alpha2.OpenTelemetryCommonFields{
				Replicas:                  taSpec.Replicas,
				NodeSelector:              taSpec.NodeSelector,
				Resources:                 taSpec.Resources,
				ServiceAccount:            taSpec.ServiceAccount,
				SecurityContext:           taSpec.SecurityContext,
				PodSecurityContext:        taSpec.PodSecurityContext,
				Image:                     taSpec.Image,
				Affinity:                  taSpec.Affinity,
				TopologySpreadConstraints: taSpec.TopologySpreadConstraints,
				Tolerations:               taSpec.Tolerations,
				Env:                       taSpec.Env,
				PodAnnotations:            params.OtelCol.Spec.PodAnnotations,
			},
			CollectorSelector:  collectorSelector,
			AllocationStrategy: v1alpha2.TargetAllocatorAllocationStrategy(taSpec.AllocationStrategy),
			FilterStrategy:     v1alpha2.TargetAllocatorFilterStrategy(taSpec.FilterStrategy),
			ScrapeConfigs:      scrapeConfigs,
			PrometheusCR:       TargetAllocatorPrometheusCR(taSpec.PrometheusCR),
		},
	}, nil
}

// TargetAllocatorPrometheusCR converts v1alpha1 PrometheusCR subresource to v1alpha2.
func TargetAllocatorPrometheusCR(promCR v1alpha1.OpenTelemetryTargetAllocatorPrometheusCR) v1alpha2.TargetAllocatorPrometheusCR {
	v2promCR := v1alpha2.TargetAllocatorPrometheusCR{
		Enabled:        promCR.Enabled,
		ScrapeInterval: promCR.ScrapeInterval,
	}
	if promCR.ServiceMonitorSelector != nil {
		v2promCR.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchLabels: promCR.ServiceMonitorSelector,
		}
	}
	if promCR.PodMonitorSelector != nil {
		v2promCR.PodMonitorSelector = &metav1.LabelSelector{
			MatchLabels: promCR.PodMonitorSelector,
		}
	}
	return v2promCR
}

func getScrapeConfigs(otelcolConfig string) ([]v1alpha2.ScrapeConfig, error) {
	// Collector supports environment variable substitution, but the TA does not.
	// TA Scrape Configs should have a single "$", as it does not support env var substitution
	prometheusReceiverConfig, err := adapters.UnescapeDollarSignsInPromConfig(otelcolConfig)
	if err != nil {
		return nil, err
	}

	scrapeConfigs, err := adapters.GetScrapeConfigsFromPromConfig(prometheusReceiverConfig)
	if err != nil {
		return nil, err
	}

	v1alpha2scrapeConfigs := make([]v1alpha2.ScrapeConfig, len(scrapeConfigs))

	for i, config := range scrapeConfigs {
		v1alpha2scrapeConfigs[i] = config
	}

	return v1alpha2scrapeConfigs, nil
}
