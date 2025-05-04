package filter

import (
	"agentt/internal/guidance/backend"
	"reflect"
	"testing"
)

// Helper function to create a simple summary for testing
func makeSummary(id, typeName, tier string, tags []string) backend.Summary {
	return backend.Summary{
		ID:          id,
		Type:        typeName,
		Tier:        tier,
		Tags:        tags,
		Description: "Test description for " + id,
	}
}

func TestParseFilter_Basic(t *testing.T) {
	tests := []struct {
		name        string
		filterStr   string
		wantNode    FilterNode
		wantErr     bool
		expectPanic bool // Added for cases like nil pointers if structure is wrong
	}{
		{
			name:      "Empty filter",
			filterStr: "",
			wantNode:  nil, // Represents match all
			wantErr:   false,
		},
		{
			name:      "Single term tier:must",
			filterStr: "tier:must",
			wantNode:  &TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
			wantErr:   false,
		},
		{
			name:      "Single term case insensitive key TIER:MUST",
			filterStr: "TIER:MUST",
			wantNode:  &TermNode{Key: "tier", Value: "MUST", Negated: false, IsTagKey: false}, // Key normalized, value case preserved
			wantErr:   false,
		},
		{
			name:      "Single term tag:scope:core",
			filterStr: "tag:scope:core",
			wantNode:  &TermNode{Key: "tag", Value: "scope:core", Negated: false, IsTagKey: true},
			wantErr:   false,
		},
		{
			name:      "Single term type:recipe",
			filterStr: "type:recipe",
			wantNode:  &TermNode{Key: "type", Value: "recipe", Negated: false, IsTagKey: false},
			wantErr:   false,
		},
		{
			name:      "Single term key existence tier:*",
			filterStr: "tier:*",
			wantNode:  &TermNode{Key: "tier", Value: "*", Negated: false, IsTagKey: false},
			wantErr:   false,
		},
		{
			name:      "Single term key existence tag:*",
			filterStr: "tag:*",
			wantNode:  &TermNode{Key: "tag", Value: "*", Negated: false, IsTagKey: true},
			wantErr:   false,
		},
		{
			name:      "Single negated term -tier:must",
			filterStr: "-tier:must",
			wantNode:  &TermNode{Key: "tier", Value: "must", Negated: true, IsTagKey: false},
			wantErr:   false,
		},
		{
			name:      "Single negated term -tag:ignore",
			filterStr: "-tag:ignore",
			wantNode:  &TermNode{Key: "tag", Value: "ignore", Negated: true, IsTagKey: true},
			wantErr:   false,
		},
		{
			name:      "Multiple terms implicit AND",
			filterStr: "tier:must tag:scope:core",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&TermNode{Key: "tag", Value: "scope:core", Negated: false, IsTagKey: true},
			}},
			wantErr: false,
		},
		{
			name:      "Multiple terms with negation implicit AND",
			filterStr: "tier:should -tag:exclude type:behavior",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "should", Negated: false, IsTagKey: false},
				&TermNode{Key: "tag", Value: "exclude", Negated: true, IsTagKey: true},
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
			}},
			wantErr: false,
		},
		{
			name:      "Term with extra whitespace",
			filterStr: "  tier:must   tag:ok  ",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&TermNode{Key: "tag", Value: "ok", Negated: false, IsTagKey: true},
			}},
			wantErr: false,
		},
		{
			name:      "Invalid format - missing colon",
			filterStr: "tier must",
			wantNode:  nil,
			wantErr:   true,
		},
		{
			name:      "Invalid format - missing value",
			filterStr: "tier:",
			wantNode:  nil,
			wantErr:   true,
		},
		{
			name:      "Invalid format - missing key",
			filterStr: ":must",
			wantNode:  nil,
			wantErr:   true,
		},
		{
			name:      "Invalid format - only negation",
			filterStr: "-",
			wantNode:  nil,
			wantErr:   true,
		},
		// --- Tests for explicit NOT (now expected to PASS parsing) ---
		{
			name:      "Explicit NOT single term",
			filterStr: "NOT tier:must",
			wantNode:  &NotNode{Child: &TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false}}, // This is the desired outcome
			wantErr:   false,                                                                                   // Expect new parser to succeed
		},
		{
			name:      "Explicit NOT with implicit AND",
			filterStr: "type:behavior NOT tag:obsolete",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
				&NotNode{Child: &TermNode{Key: "tag", Value: "obsolete", Negated: false, IsTagKey: true}},
			}}, // This is the desired outcome
			wantErr: false, // Expect new parser to succeed
		},
		{
			name:      "Explicit NOT at start of implicit AND",
			filterStr: "NOT tag:obsolete type:behavior",
			wantNode: &AndNode{Children: []FilterNode{
				&NotNode{Child: &TermNode{Key: "tag", Value: "obsolete", Negated: false, IsTagKey: true}},
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
			}}, // This is the desired outcome
			wantErr: false, // Expect new parser to succeed
		},
		{
			name:      "Explicit NOT applied to prefix negation (invalid) - Should fail parsing",
			filterStr: "NOT -tier:must",
			wantNode:  nil,
			wantErr:   true, // New parser should reject this double negation
		},
		{
			name:      "Explicit NOT with nothing following (invalid)",
			filterStr: "tier:must NOT",
			wantNode:  nil,
			wantErr:   true, // New parser should reject
		},
		// --- Tests for explicit AND (now expected to PASS parsing) ---
		{
			name:      "Explicit AND two terms",
			filterStr: "tier:must AND type:behavior",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
			}}, // Desired outcome
			wantErr: false, // Expect new parser to succeed
		},
		{
			name:      "Explicit AND three terms",
			filterStr: "tier:must AND type:behavior AND tag:core",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
				&TermNode{Key: "tag", Value: "core", Negated: false, IsTagKey: true},
			}}, // Desired outcome
			wantErr: false, // Expect new parser to succeed
		},
		{
			name:      "Explicit AND mixed with NOT",
			filterStr: "tier:must AND NOT tag:obsolete",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&NotNode{Child: &TermNode{Key: "tag", Value: "obsolete", Negated: false, IsTagKey: true}},
			}}, // Desired outcome
			wantErr: false, // Expect new parser to succeed
		},
		{
			name:      "Explicit AND mixed case",
			filterStr: "tier:must and type:behavior",
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
			}}, // Desired outcome (same as uppercase AND)
			wantErr: false, // Expect new parser to succeed
		},
		{
			name:      "Explicit AND missing term after",
			filterStr: "tier:must AND",
			wantNode:  nil,
			wantErr:   true, // Should fail
		},
		{
			name:      "Explicit AND missing term before",
			filterStr: "AND type:behavior",
			wantNode:  nil,
			wantErr:   true, // Should fail
		},
		{
			name:      "Multiple ANDs (consecutive - invalid)",
			filterStr: "tier:must AND AND type:behavior",
			wantNode:  nil,
			wantErr:   true, // Should fail
		},
		{
			name:      "Term following term (implicit AND allowed)",
			filterStr: "tier:must type:behavior", // Implicit AND is allowed again.
			wantNode: &AndNode{Children: []FilterNode{
				&TermNode{Key: "tier", Value: "must", Negated: false, IsTagKey: false},
				&TermNode{Key: "type", Value: "behavior", Negated: false, IsTagKey: false},
			}}, // Should parse as AND
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("ParseFilter() panicked unexpectedly: %v", r)
					}
				} else if tt.expectPanic {
					t.Errorf("ParseFilter() did not panic as expected")
				}
			}()

			gotNode, err := ParseFilter(tt.filterStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNode, tt.wantNode) {
				t.Errorf("ParseFilter() gotNode = %v (%s), want %v (%s)", gotNode, nodeToString(gotNode), tt.wantNode, nodeToString(tt.wantNode))
			}
		})
	}
}

// Helper to safely get string representation of a potentially nil node
func nodeToString(node FilterNode) string {
	if node == nil {
		return "<nil>"
	}
	return node.String()
}

func TestFilterNode_Evaluate(t *testing.T) {
	summaryMustCore := makeSummary("b1", "behavior", "must", []string{"scope:core", "prio:high"})
	summaryShouldCore := makeSummary("b2", "behavior", "should", []string{"scope:core", "prio:medium"})
	summaryRecipeCore := makeSummary("r1", "recipe", "", []string{"scope:core", "tech:go"})
	summaryNoTags := makeSummary("b4", "behavior", "must", []string{})
	summaryNilTags := makeSummary("b5", "behavior", "must", nil)
	summaryWildcard := makeSummary("w1", "behavior", "should", []string{"scope:core", "prefix:value", "value:suffix", "prefix:value:suffix", "subvalue"})

	tests := []struct {
		name    string
		filter  string // Use ParseFilter to generate the node
		summary backend.Summary
		want    bool
		wantErr bool // For parse errors
	}{
		// Simple Matches
		{"Match tier:must", "tier:must", summaryMustCore, true, false},
		{"Match TIER:MUST (case)", "TIER:MUST", summaryMustCore, true, false},
		{"No match tier:must", "tier:must", summaryShouldCore, false, false},
		{"Match tier:should", "tier:should", summaryShouldCore, true, false},
		{"Match type:behavior", "type:behavior", summaryMustCore, true, false},
		{"Match type:recipe", "type:recipe", summaryRecipeCore, true, false},
		{"No match type:recipe", "type:recipe", summaryMustCore, false, false},
		{"Match tag:scope:core", "tag:scope:core", summaryMustCore, true, false},
		{"Match tag:prio:high", "tag:prio:high", summaryMustCore, true, false},
		{"Match TAG:scope:core (case key)", "TAG:scope:core", summaryMustCore, true, false},
		{"Match tag:SCOPE:CORE (case value)", "tag:SCOPE:CORE", summaryMustCore, true, false},
		{"No match tag:missing", "tag:missing", summaryMustCore, false, false},
		{"Match id:b1", "id:b1", summaryMustCore, true, false},
		{"No match id:b2", "id:b2", summaryMustCore, false, false},

		// Negation
		{"Negated tier -tier:should", "-tier:should", summaryMustCore, true, false},
		{"Negated tier -tier:must", "-tier:must", summaryMustCore, false, false},
		{"Negated tag -tag:scope:infra", "-tag:scope:infra", summaryMustCore, true, false},
		{"Negated tag -tag:scope:core", "-tag:scope:core", summaryMustCore, false, false},
		{"Negated type -type:recipe", "-type:recipe", summaryMustCore, true, false},
		{"Negated type -type:behavior", "-type:behavior", summaryMustCore, false, false},
		{"Negated id -id:b2", "-id:b2", summaryMustCore, true, false},
		{"Negated id -id:b1", "-id:b1", summaryMustCore, false, false},

		// Key Existence
		{"Exist tier:* on must", "tier:*", summaryMustCore, true, false},
		{"Exist tier:* on should", "tier:*", summaryShouldCore, true, false},
		{"Exist tier:* on recipe (no tier)", "tier:*", summaryRecipeCore, false, false},
		{"Exist -tier:* on recipe", "-tier:*", summaryRecipeCore, true, false},
		{"Exist tag:* with tags", "tag:*", summaryMustCore, true, false},
		{"Exist tags:* with tags", "tags:*", summaryMustCore, true, false},
		{"Exist tag:* no tags", "tag:*", summaryNoTags, false, false},
		{"Exist tag:* nil tags", "tag:*", summaryNilTags, false, false},
		{"Exist -tag:* no tags", "-tag:*", summaryNoTags, true, false},
		{"Exist description:*", "description:*", summaryMustCore, true, false}, // Assuming desc is always set
		{"Exist id:*", "id:*", summaryMustCore, true, false},                   // Assuming ID is always set
		{"Exist type:*", "type:*", summaryMustCore, true, false},               // Assuming Type is always set

		// Implicit AND
		{"AND tier:must tag:scope:core", "tier:must tag:scope:core", summaryMustCore, true, false},
		{"AND TIER:MUST TAG:scope:core (case)", "TIER:MUST TAG:scope:core", summaryMustCore, true, false},
		{"AND tier:must tag:scope:infra", "tier:must tag:scope:infra", summaryMustCore, false, false},   // Tag mismatch
		{"AND tier:should tag:scope:core", "tier:should tag:scope:core", summaryMustCore, false, false}, // Tier mismatch
		{"AND tier:must -tag:scope:infra", "tier:must -tag:scope:infra", summaryMustCore, true, false},
		{"AND tier:must -tag:scope:core", "tier:must -tag:scope:core", summaryMustCore, false, false}, // Negated tag matches
		{"AND tier:must type:behavior", "tier:must type:behavior", summaryMustCore, true, false},
		{"AND tier:must type:recipe", "tier:must type:recipe", summaryMustCore, false, false},
		{"AND tier:* tag:scope:core", "tier:* tag:scope:core", summaryMustCore, true, false},
		{"AND tier:* tag:scope:core recipe", "tier:* tag:scope:core", summaryRecipeCore, false, false}, // Has tag but tier:* fails
		{"AND -tier:* tag:scope:core recipe", "-tier:* tag:scope:core", summaryRecipeCore, true, false},

		// --- Wildcard Tag Tests (New - Expected to Fail Evaluation Initially) ---
		{"Tag Wildcard Prefix Match", "tag:scope:*", summaryWildcard, true, false},
		{"Tag Wildcard Prefix No Match", "tag:nomatch:*", summaryWildcard, false, false},
		{"Tag Wildcard Suffix Match", "tag:*:core", summaryWildcard, true, false},
		{"Tag Wildcard Suffix No Match", "tag:*:nomatch", summaryWildcard, false, false},
		{"Tag Wildcard Substring Match 1", "tag:*value*", summaryWildcard, true, false}, // Matches prefix:value, value:suffix, prefix:value:suffix, subvalue
		{"Tag Wildcard Substring Match 2", "tag:*fix*", summaryWildcard, true, false},   // Matches value:suffix, prefix:value:suffix
		{"Tag Wildcard Substring No Match", "tag:*nomatch*", summaryWildcard, false, false},
		{"Tag Wildcard Prefix+Suffix Match", "tag:prefix*:*suffix", summaryWildcard, true, false}, // Matches prefix:value:suffix
		{"Tag Wildcard Prefix+Suffix No Match (Prefix)", "tag:nomatch*:*suffix", summaryWildcard, false, false},
		{"Tag Wildcard Prefix+Suffix No Match (Suffix)", "tag:prefix*:*nomatch", summaryWildcard, false, false},
		{"Tag Wildcard Exact Match Still Works", "tag:scope:core", summaryWildcard, true, false},
		{"Tag Wildcard Negated Prefix", "-tag:scope:*", summaryWildcard, false, false},
		{"Tag Wildcard Negated Suffix", "-tag:*:core", summaryWildcard, false, false},
		{"Tag Wildcard Negated Substring", "-tag:*value*", summaryWildcard, false, false},

		// Empty/Invalid Filters
		{"Empty filter matches all", "", summaryMustCore, true, false},
		{"Whitespace filter matches all", "   ", summaryMustCore, true, false},
		{"Invalid filter parse err", "tier must", summaryMustCore, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseFilter(tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilter() error = %v, wantErr %v for filter '%s'", err, tt.wantErr, tt.filter)
				return
			}
			if tt.wantErr {
				return // Stop here if parse error was expected
			}

			var got bool
			if node == nil {
				// No filter means match all
				got = true
			} else {
				got = node.Evaluate(tt.summary)
			}

			if got != tt.want {
				t.Errorf("Evaluate() got = %v, want %v for filter '%s' on summary %s (Node: %s)", got, tt.want, tt.filter, tt.summary.ID, nodeToString(node))
			}
		})
	}
}
