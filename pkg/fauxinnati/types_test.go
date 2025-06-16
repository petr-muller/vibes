package fauxinnati

import (
	"encoding/json"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/petr-muller/vibes/pkg/testhelper"
)

const metallbBgpBfdFrrRpmPromQl = `(
  group by (_id, name) (csv_succeeded{_id="", name=~"metallb-operator[.].*"})
  or on (_id)
  0 * label_replace(group by (_id) (csv_succeeded{_id=""}), "name", "metallb operator not installed", "name", ".*")
)`

func TestGraph_JSON(t *testing.T) {
	rawCandidate420 := testhelper.ReadFromFixture(t, "candidate-4.20.json", ".input")
	var candidate420 Graph
	err := json.Unmarshal(rawCandidate420, &candidate420)
	if err != nil {
		t.Fatalf("Failed to unmarshal candidate-4.20 graph: %v", err)
	}

	tests := []struct {
		name  string
		graph Graph
	}{
		{
			name: "empty graph",
			graph: Graph{
				Nodes:            []Node{},
				Edges:            []Edge{},
				ConditionalEdges: []ConditionalEdge{},
			},
		},
		{
			name: "simple graph with three nodes and two edges, one of them conditional",
			graph: Graph{
				Nodes: []Node{
					{
						Version: semver.MustParse("4.19.0"),
						Image:   "quay.io/openshift-release-dev/ocp-release@sha256:3482dbdce3a6fb2239684d217bba6fc87453eff3bdb72f5237be4beb22a2160b",
						Metadata: map[string]string{
							"io.openshift.upgrades.graph.release.channels":    "candidate-4.19,candidate-4.20",
							"io.openshift.upgrades.graph.release.manifestref": "sha256:3482dbdce3a6fb2239684d217bba6fc87453eff3bdb72f5237be4beb22a2160b",
							"url": "https://access.redhat.com/errata/RHSA-2024:11038",
						},
					},
					{
						Version: semver.MustParse("4.19.1"),
						Image:   "quay.io/openshift-release-dev/ocp-release@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
						Metadata: map[string]string{
							"io.openshift.upgrades.graph.release.channels":    "candidate-4.19,candidate-4.20",
							"io.openshift.upgrades.graph.release.manifestref": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
							"url": "https://access.redhat.com/errata/RHSA-2024:11039",
						},
					},
					{
						Version: semver.MustParse("4.19.2"),
						Image:   "quay.io/openshift-release-dev/ocp-release@sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
						Metadata: map[string]string{
							"io.openshift.upgrades.graph.release.channels":    "candidate-4.19,candidate-4.20",
							"io.openshift.upgrades.graph.release.manifestref": "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
							"url": "https://access.redhat.com/errata/RHSA-2024:11040",
						},
					},
				},
				Edges: []Edge{
					{0, 1}, // 4.19.0 -> 4.19.1
				},
				ConditionalEdges: []ConditionalEdge{
					{
						Edges: []ConditionalUpdate{
							{From: "4.19.1", To: "4.19.2"},
						},
						Risks: []ConditionalUpdateRisk{
							{
								URL:     "https://issues.redhat.com/browse/CNF-17689",
								Name:    "MetallbBgpBfdFrrRpm",
								Message: "Clusters using MetalLB BFD capabilities alongside BGP can fail to establish BGP peering, reducing the availability of LoadBalancer services exposed by MetalLB, or even making them unreachable",
								MatchingRules: []MatchingRule{
									{
										Type: "PromQL",
										PromQL: &PromQLQuery{
											PromQL: metallbBgpBfdFrrRpmPromQl,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "candidate-4.20 graph snapshot from OSUS",
			graph: candidate420,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.MarshalIndent(tt.graph, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal graph: %v", err)
			}

			testhelper.CompareWithFixture(t, data)

			var result Graph
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal graph: %v", err)
			}

			if diff := cmp.Diff(tt.graph, result); diff != "" {
				t.Errorf("Graph marshal roundtrip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNode_JSON(t *testing.T) {
	tests := []struct {
		name string
		node Node
	}{
		{
			name: "4.19.0",
			node: Node{
				Version: semver.MustParse("4.19.0"),
				Image:   "quay.io/openshift-release-dev/ocp-release@sha256:3482dbdce3a6fb2239684d217bba6fc87453eff3bdb72f5237be4beb22a2160b",
				Metadata: map[string]string{
					"io.openshift.upgrades.graph.release.channels":    "candidate-4.19,candidate-4.20",
					"io.openshift.upgrades.graph.release.manifestref": "sha256:3482dbdce3a6fb2239684d217bba6fc87453eff3bdb72f5237be4beb22a2160b",
					"url": "https://access.redhat.com/errata/RHSA-2024:11038",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.MarshalIndent(tt.node, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal node: %v", err)
			}

			testhelper.CompareWithFixture(t, data)
		})
	}
}

func TestEdge_JSON(t *testing.T) {
	tests := []struct {
		name     string
		edge     Edge
		expected string
	}{
		{
			name:     "edge 0->1",
			edge:     Edge{0, 1},
			expected: "[0,1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.edge)
			if err != nil {
				t.Fatalf("Failed to marshal edge: %v", err)
			}

			if diff := cmp.Diff(tt.expected, string(data)); diff != "" {
				t.Errorf("JSON mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConditionalEdge_JSON(t *testing.T) {
	tests := []struct {
		name string
		edge ConditionalEdge
	}{
		{
			name: "one edge with one risk",
			edge: ConditionalEdge{
				Edges: []ConditionalUpdate{{From: "4.18.0", To: "4.19.0"}},
				Risks: []ConditionalUpdateRisk{
					{
						URL:           "https://issues.redhat.com/browse/OCPNODE-3245",
						Name:          "KubeletStartFailingFromRestoreconTimeout",
						Message:       "In some cases, a SELinux restorecon run as a prerequisite of kubelet.service was failing.",
						MatchingRules: []MatchingRule{{Type: "Always"}},
					},
				},
			},
		},
		{
			name: "multiple edges with multiple risks",
			edge: ConditionalEdge{
				Edges: []ConditionalUpdate{
					{From: "4.18.0", To: "4.19.0"},
					{From: "4.19.0", To: "4.20.0"},
				},
				Risks: []ConditionalUpdateRisk{
					{
						URL:     "https://issues.redhat.com/browse/CNF-17689",
						Name:    "MetallbBgpBfdFrrRpm",
						Message: "Clusters using MetalLB BFD capabilities alongside BGP can fail to establish BGP peering, reducing the availability of LoadBalancer services exposed by MetalLB, or even making them unreachable",
						MatchingRules: []MatchingRule{
							{
								Type: "PromQL",
								PromQL: &PromQLQuery{
									PromQL: metallbBgpBfdFrrRpmPromQl,
								},
							},
						},
					},
					{
						URL:           "https://issues.redhat.com/browse/OCPNODE-3245",
						Name:          "KubeletStartFailingFromRestoreconTimeout",
						Message:       "In some cases, a SELinux restorecon run as a prerequisite of kubelet.service was failing.",
						MatchingRules: []MatchingRule{{Type: "Always"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.MarshalIndent(tt.edge, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal conditional edge: %v", err)
			}

			testhelper.CompareWithFixture(t, data)
		})
	}
}

func TestConditionalUpdateRisk_JSON(t *testing.T) {
	tests := []struct {
		name string
		risk ConditionalUpdateRisk
	}{
		{
			name: "always risk",
			risk: ConditionalUpdateRisk{
				URL:           "https://issues.redhat.com/browse/OCPNODE-3245",
				Name:          "KubeletStartFailingFromRestoreconTimeout",
				Message:       "In some cases, a SELinux restorecon run as a prerequisite of kubelet.service was failing.",
				MatchingRules: []MatchingRule{{Type: "Always"}},
			},
		},
		{
			name: "promql risk",
			risk: ConditionalUpdateRisk{
				URL:     "https://issues.redhat.com/browse/CNF-17689",
				Name:    "MetallbBgpBfdFrrRpm",
				Message: "Clusters using MetalLB BFD capabilities alongside BGP can fail to establish BGP peering, reducing the availability of LoadBalancer services exposed by MetalLB, or even making them unreachable",
				MatchingRules: []MatchingRule{
					{
						Type: "PromQL",
						PromQL: &PromQLQuery{
							PromQL: metallbBgpBfdFrrRpmPromQl,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.MarshalIndent(tt.risk, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal risk: %v", err)
			}

			testhelper.CompareWithFixture(t, data)
		})
	}
}
