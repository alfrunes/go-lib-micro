// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package plan

import (
	"testing"
)

func TestHasFeature(t *testing.T) {
	RegisterPlan("test", FeatureSAML)

	if !HasFeatures("test", FeatureSAML) {
		t.Error(`plan "test" should have SAML feature`)
		t.Fail()
	}
	if HasFeatures("test", FeatureAuditLogsFull) {
		t.Error(`plan "test" should not have access to auditlogs`)
		t.Fail()
	}
	if HasFeatures("test", FeatureSAML|FeatureAuditLogsFull) {
		t.Error(`plan "test" only has access to SAML, not auditlogs`)
		t.Fail()
	}
}
