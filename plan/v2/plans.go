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
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	planBasic        = "basic"
	planOpenSource   = "os"
	planProfessional = "professional"
	planEnterprise   = "enterprise"
	planTrial        = "trial"

	addonTroubleShoot = "troubleshoot"
	addonConfigure    = "configure"
	addonMonitor      = "monitor"
)

func defaultPlans() map[string]Features {
	return map[string]Features{
		addonConfigure:    FeatureConfigure,
		addonMonitor:      FeatureMonitor,
		addonTroubleShoot: FeatureConnect,

		planBasic:        Features(0),
		planOpenSource:   Features(0),
		planProfessional: FeatureExtSearch | FeatureTwoFactor, //TODO
		planTrial:        ^FeatureAuditLogsFull,
		planEnterprise:   ^Features(0),
	}
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
		planToFeatures = defaultPlans()
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

func CombineFeatureSets(plans ...string) (features Features) {
	for _, plan := range plans {
		if f, ok := planToFeatures[plan]; ok {
			features |= f
		}
	}
	return features
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
	FeatureConnect                                   // Remote-terminal and file transfer
	FeatureConnectPlayback                           // Session playback
	FeatureConfigure                                 // Configuration deployments
	FeatureMonitor                                   // Device monitoring and alerting
)

type rule struct {
	// Path expression
	Path     string
	IsPrefix bool
	Methods  []string
	Features
}

var rules = []rule{{
	Features: FeatureExtSearch,
	Path:     `/api/*/*/inventory/devices/search`,
	IsPrefix: true,
}, {
	Features: FeatureConnect,
	Path:     `/api/*/*/deviceconnect/devices`,
	IsPrefix: true,
}, {
	Features: FeatureConnectPlayback,
	Path:     `/api/*/*/deviceconnect/sessions/*/playback`,
}, {
	Features: FeatureMonitor,
	Path:     `/api/*/*/devicemonitor`,
	IsPrefix: true,
}, {
	Features: FeatureRBAC,
	Path:     `/api/*/*/useradm/roles`,
	IsPrefix: true,
}, {
	Features: FeatureRBAC,
	Path:     `/api/*/*/useradm/permission_sets`,
	IsPrefix: true,
}}

type matcher struct {
	Children map[string]*matcher
	Features
}

func (m *matcher) Lookup(path string) (features Features) {
	matchers := []*matcher{m}
	for _, segment := range strings.Split(path, "/") {
		fmt.Println(segment)
		for i := len(matchers) - 1; i >= 0; i-- {
			m := matchers[i]
			fmt.Println(*m)
			if child, ok := m.Children[segment]; ok {
				matchers[i] = child
			} else {
				matchers = append(matchers[:i], matchers[i+1:]...)
			}
			if wc, ok := m.Children["*"]; ok {
				matchers = append(matchers, wc)
			}
		}
	}
	for _, matcher := range matchers {
		features |= matcher.Features
	}
	return features
}

func compileRules(rules []rule) *matcher {
	root := &matcher{
		Children: make(map[string]*matcher),
	}
	sort.Slice(rules, func(i, j int) bool {
		return !rules[i].IsPrefix && rules[j].IsPrefix
	})
	i := sort.Search(len(rules),
		func(i int) bool { return rules[i].IsPrefix },
	)
	for _, rule := range rules[:i] {
		segments := strings.Split(rule.Path, "/")
		var (
			segment string
			ok      bool
			newNode *matcher
			node    *matcher = root
		)
		for _, segment = range segments {
			if newNode, ok = node.Children[segment]; ok {
				node = newNode
				continue
			} else {
				newNode = &matcher{
					Children: make(map[string]*matcher),
				}
				node.Children[segment] = newNode
				node = newNode
			}
		}
		newNode.Features |= rule.Features
	}
	for _, rule := range rules[i:] {
		segments := strings.Split(rule.Path, "/")
		var (
			segment string
			ok      bool
			newNode *matcher
			node    *matcher = root
		)
		for _, segment = range segments {
			if newNode, ok = node.Children[segment]; ok {
				node = newNode
				continue
			} else {
				newNode = &matcher{
					Children: make(map[string]*matcher),
				}
				node.Children[segment] = newNode
				node = newNode
			}
		}
		applyFeaturesRecursive(rule.Features, newNode)
	}
	return root
}

func applyFeaturesRecursive(f Features, node *matcher) {
	node.Features |= f
	for _, child := range node.Children {
		applyFeaturesRecursive(f, child)
	}
}
