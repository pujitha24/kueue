/*
Copyright The Kubernetes Authors.

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

package core

import (
	"testing"

	nodev1 "k8s.io/api/node/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	qcache "sigs.k8s.io/kueue/pkg/cache/queue"
	schdcache "sigs.k8s.io/kueue/pkg/cache/scheduler"
	"sigs.k8s.io/kueue/pkg/controller/core/indexer"
	"sigs.k8s.io/kueue/pkg/features"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	utiltestingapi "sigs.k8s.io/kueue/pkg/util/testing/v1beta2"
)

func TestResourceUpdatesHandlerRuntimeClassListsOnlyMatchingWorkloads(t *testing.T) {
	// DRA workloads, so that each listed workload shows up as a reconcile request.
	features.SetFeatureGateDuringTest(t, features.KueueDRAIntegration, true)
	ctx, _ := utiltesting.ContextWithLog(t)
	podSet := func(runtimeClass string) *utiltestingapi.PodSetWrapper {
		p := utiltestingapi.MakePodSet(kueue.DefaultPodSetName, 1).ResourceClaimTemplate("gpu", "tmpl")
		if runtimeClass != "" {
			p = p.RuntimeClass(runtimeClass)
		}
		return p
	}
	cl := utiltesting.NewClientBuilder().
		WithIndex(&kueue.Workload{}, indexer.WorkloadQuotaReservedKey, indexer.IndexWorkloadQuotaReserved).
		WithIndex(&kueue.Workload{}, indexer.WorkloadRuntimeClassKey, indexer.IndexWorkloadRuntimeClass).
		WithObjects(
			utiltestingapi.MakeWorkload("uses-kata", "ns1").Queue("lq").PodSets(*podSet("kata").Obj()).Obj(),
			utiltestingapi.MakeWorkload("uses-gvisor", "ns2").Queue("lq").PodSets(*podSet("gvisor").Obj()).Obj(),
			utiltestingapi.MakeWorkload("no-runtimeclass", "ns3").Queue("lq").PodSets(*podSet("").Obj()).Obj(),
		).Build()
	cqCache := schdcache.New(cl)
	h := &resourceUpdatesHandler{r: NewWorkloadReconciler(cl, qcache.NewManagerForUnitTests(cl, cqCache), cqCache, &utiltesting.EventRecorder{})}
	q := &utiltesting.MockTypedRateLimitingInterface{}

	runtimeClass := &nodev1.RuntimeClass{Handler: "kata"}
	runtimeClass.Name = "kata"
	h.Create(ctx, event.CreateEvent{Object: runtimeClass}, q)

	if len(q.Items) != 1 {
		t.Errorf("reconcile requests = %v, want only ns1/uses-kata", q.Items)
	}
}
