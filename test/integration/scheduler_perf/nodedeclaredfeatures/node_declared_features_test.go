/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodedeclaredfeatures

import (
	"fmt"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/version"
	_ "k8s.io/component-base/logs/json/register"
	ndf "k8s.io/component-helpers/nodedeclaredfeatures"
	ndffeatures "k8s.io/component-helpers/nodedeclaredfeatures/features"
	perf "k8s.io/kubernetes/test/integration/scheduler_perf"
	"k8s.io/kubernetes/test/utils/ktesting"
)

func TestMain(m *testing.M) {
	if err := perf.InitTests(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	m.Run()
}

// mockFeature is a mock implementation of the Feature interface for testing.
type mockFeature struct {
	name               string
	discover           func(cfg *ndf.NodeConfiguration) bool
	inferForScheduling func(podInfo *ndf.PodInfo) bool
	inferForUpdate     func(oldPodInfo, newPodInfo *ndf.PodInfo) bool
	maxVersion         *version.Version
}

func (f *mockFeature) Name() string {
	return f.name
}

func (f *mockFeature) Discover(cfg *ndf.NodeConfiguration) bool {
	return true
}

func (f *mockFeature) InferForScheduling(podInfo *ndf.PodInfo) bool {
	// Check if any container has an env var matching the feature name.
	if podInfo.Spec != nil && podInfo.Spec.Resources != nil {
		return true
	}
	return false
}

func (f *mockFeature) InferForUpdate(oldPodInfo, newPodInfo *ndf.PodInfo) bool {
	return false
}

func (f *mockFeature) MaxVersion() *version.Version {
	return f.maxVersion
}

func createMockFeature(name string, maxVersionStr string) ndf.Feature {
	var v *version.Version
	if maxVersionStr != "" {
		v = version.MustParseSemantic(maxVersionStr)
	}
	return &mockFeature{
		name:       name,
		maxVersion: v,
	}
}

func setupFeatures(numFeatures int) func() {
	nodeFeatures := make([]ndf.Feature, 0, numFeatures)
	for i := 1; i <= numFeatures; i++ {
		featureName := fmt.Sprintf("Feature%d", i)
		nodeFeatures = append(nodeFeatures, createMockFeature(featureName, ""))
	}
	originalAllFeatures := ndffeatures.AllFeatures
	ndffeatures.AllFeatures = nodeFeatures
	return func() {
		ndffeatures.AllFeatures = originalAllFeatures
	}
}

// defaultDeclaredFeaturesCount is used when numFeatures is not specified in performance-config.yaml.
const defaultDeclaredFeaturesCount = 5

func prepareNodeDeclaredFeatures(tCtx ktesting.TContext, w *perf.Workload) (func(), error) {
	numFeatures, err := w.Params.Get("numFeatures")
	if err != nil {
		// Default to 5 if not specified in the config.
		numFeatures = defaultDeclaredFeaturesCount
	}

	return setupFeatures(numFeatures), nil
}

func TestSchedulerPerf(t *testing.T) {
	perf.RunIntegrationPerfScheduling(t, "performance-config.yaml", perf.WithPrepareFn(prepareNodeDeclaredFeatures))
}

func BenchmarkPerfScheduling(b *testing.B) {
	perf.RunBenchmarkPerfScheduling(b, "performance-config.yaml", "nodedeclaredfeatures", nil, perf.WithPrepareFn(prepareNodeDeclaredFeatures))
}
