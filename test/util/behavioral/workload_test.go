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

package behavioral

import (
	"testing"

	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta2"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	utiltestingapi "sigs.k8s.io/kueue/pkg/util/testing/v1beta2"
)

func TestFinishWorkloadsPreservesServerConditions(t *testing.T) {
	g := gomega.NewWithT(t)
	gomega.RegisterTestingT(t)
	ctx := t.Context()

	serverCond := metav1.Condition{
		Type:   kueue.WorkloadQuotaReserved,
		Status: metav1.ConditionTrue,
		Reason: "ServerAdded",
	}
	staleCond := metav1.Condition{
		Type:   kueue.WorkloadAdmitted,
		Status: metav1.ConditionFalse,
		Reason: "Stale",
	}

	serverWL := utiltestingapi.MakeWorkload("wl", "ns").Obj()
	serverWL.Status.Conditions = []metav1.Condition{serverCond}
	k8sClient := utiltesting.NewFakeClient(serverWL)

	// The caller's object is stale and has spare capacity in its conditions slice.
	staleWL := serverWL.DeepCopy()
	staleWL.Status.Conditions = make([]metav1.Condition, 1, 4)
	staleWL.Status.Conditions[0] = staleCond

	FinishWorkloads(ctx, k8sClient, staleWL)

	var got kueue.Workload
	g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(serverWL), &got)).To(gomega.Succeed())
	g.Expect(apimeta.IsStatusConditionTrue(got.Status.Conditions, kueue.WorkloadQuotaReserved)).To(gomega.BeTrue(), "server-added condition must be preserved")
	g.Expect(apimeta.IsStatusConditionTrue(got.Status.Conditions, kueue.WorkloadFinished)).To(gomega.BeTrue())
	g.Expect(apimeta.FindStatusCondition(got.Status.Conditions, kueue.WorkloadAdmitted)).To(gomega.BeNil(), "stale caller condition must not be written")

	g.Expect(staleWL.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(staleWL.Status.Conditions[:cap(staleWL.Status.Conditions)][1:]).To(gomega.Equal(make([]metav1.Condition, 3)), "caller's backing array must not be written")
}
