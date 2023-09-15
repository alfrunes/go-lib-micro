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
	"os"
	"strconv"
	"strings"
)

const (
	planBasic        = "basic"
	planOpenSource   = "os"
	planProfessional = "professional"
	planEnterprise   = "enterprise"
)

var defaultPlans = map[string]Features{
	planBasic:        Features(0),
	planOpenSource:   Features(0),
	planProfessional: FeatureExtSearch | FeatureTwoFactor, //??
	planEnterprise:   ^Features(0),
}

var planToFeatures map[string]Features

func init() {
	// PLAN_OVERRIDES is a comma-separated list of plan overrides of the form:
	// <name>=<integer feature set>,<name2>=<integer feature set 2>
	// Any key=value pair that fails to parse will be silently ignored.
	if overrides, ok := os.LookupEnv("PLAN_OVERRIDES"); ok {
		planToFeatures = make(map[string]Features, strings.Count(overrides, ",")+1)
		for _, override := range strings.Split(overrides, ",") {
			var features Features
			if idx := strings.IndexByte(override, '='); idx > 0 {
				value, err := strconv.ParseUint(override[idx+1:], 10, 16)
				if err != nil {
					continue
				}
				features = Features(value)
				override = override[:idx]
			}
			RegisterPlan(override, features)
		}
	} else {
		planToFeatures = defaultPlans
	}
}

// RegisterPlan registers a new plan.
func RegisterPlan(name string, features Features) {
	planToFeatures[name] = features
}

// HasFeatures returns true if the given plan has ALL the features.
func HasFeatures(plan string, features Features) bool {
	featureSet, ok := planToFeatures[plan]
	if !ok {
		return false
	}
	return featureSet&features == features
}

type Features uint16

// NOTE: Always add new features to the end of this list to preserve backward compatibility.
const (
	FeatureExtSearch            Features = 1 << iota // Extensive device search (regex)
	FeatureDynamicGroups                             // Dynamic groups and deployments
	FeatureRBAC                                      // Role-based access control
	FeatureSAML                                      // SAML authentication
	FeatureTwoFactor                                 // Two-factor authentication
	FeatureAuditLogsPartial                          // Partial access to auditlogs
	FeatureAuditLogsFull                             // Full access to auditlogs history
	FeatureScheduledDeployments                      // Scheduled deployments
	FeaturePhasedRollouts                            // Scheduled deployment batches
	FeatureSynchronizedUpdates                       // Synchronized deployments
	FeatureServerDeltas                              // Server-side delta generation
)
