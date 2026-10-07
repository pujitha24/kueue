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

package e2e

import (
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	utiltestingapi "sigs.k8s.io/kueue/pkg/util/testing/v1beta2"
)

func TestAssertMsgForMk(t *testing.T) {
	gomega.RegisterTestingT(t)
	wl := utiltestingapi.MakeWorkload("wl", "ns").Obj()
	key := types.NamespacedName{Namespace: "ns", Name: "wl"}

	msg := AssertMsgForMk(t.Context(), "Workload not admitted in manager", key,
		utiltesting.NewFakeClient(wl), utiltesting.NewFakeClient(), utiltesting.NewFakeClient())()

	for _, want := range []string{"Workload not admitted in manager", "Manager", "Worker1", "Worker2"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q does not contain %q", msg, want)
		}
	}
}
