package filter

import (
	"agentt/internal/guidance/backend"
	"fmt"
	"strings"
)

// FilterNode represents a node in the filter expression tree.
type FilterNode interface {
	// Evaluate checks if the given summary matches the filter node.
	Evaluate(summary backend.Summary) bool
	// String returns a string representation of the node (for debugging).
	String() string
}

// TermNode represents a basic key-value or key-existence term.
// Examples: "tier:must", "tag:scope:core", "priority:*", "-tag:ignore"
type TermNode struct {
	Key      string
	Value    string // "*" for key existence
	Negated  bool   // True if the term was prefixed with '-' or part of NOT
	IsTagKey bool   // True if Key is "tag"
}

func (n *TermNode) Evaluate(summary backend.Summary) bool {
	var match bool

	// --- Handle Existence Check First --- START ---
	if n.Value == "*" {
		switch n.Key {
		case "tier":
			match = summary.Tier != ""
		case "tag", "tags": // Allow tag:* or tags:*
			match = len(summary.Tags) > 0
		case "description":
			match = summary.Description != ""
		case "id":
			match = summary.ID != "" // Should always exist
		case "type":
			match = summary.Type != "" // Should always exist
		// Add other known summary fields if they can be checked for existence
		default:
			// Unrecognized key for existence check
			match = false
			fmt.Printf("Warning: Unrecognized key '%s' for existence check\n", n.Key) // TODO: Use logger
		}
		// --- Handle Existence Check First --- END ---
	} else {
		// --- Handle Value Matching --- START ---
		if n.Key == "tier" {
			match = strings.EqualFold(summary.Tier, n.Value)
		} else if n.IsTagKey { // Key is "tag"
			// Check if any tag matches n.Value (case-insensitive)
			for _, tag := range summary.Tags {
				if strings.EqualFold(tag, n.Value) {
					match = true
					break
				}
			}
		} else if n.Key == "type" {
			match = strings.EqualFold(summary.Type, n.Value)
		} else if n.Key == "id" {
			match = strings.EqualFold(summary.ID, n.Value)
			// Add other potential top-level summary fields if needed (e.g., description)
		} else {
			// Unrecognized key for value check
			match = false
			fmt.Printf("Warning: Unrecognized key '%s' for value check\n", n.Key) // TODO: Use logger
		}
		// --- Handle Value Matching --- END ---
	}

	if n.Negated {
		return !match
	}
	return match
}

func (n *TermNode) String() string {
	prefix := ""
	if n.Negated {
		prefix = "-"
	}
	return fmt.Sprintf("%s%s:%s", prefix, n.Key, n.Value)
}

// AndNode represents a logical AND of multiple conditions.
// Assumes implicit AND between terms for now.
type AndNode struct {
	Children []FilterNode
}

func (n *AndNode) Evaluate(summary backend.Summary) bool {
	if len(n.Children) == 0 {
		return true // Empty AND is true
	}
	for _, child := range n.Children {
		if !child.Evaluate(summary) {
			return false // Short-circuit
		}
	}
	return true
}

func (n *AndNode) String() string {
	parts := make([]string, len(n.Children))
	for i, child := range n.Children {
		parts[i] = child.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, " AND "))
}

// NotNode represents a logical NOT of a single condition.
// Handles the explicit "NOT" keyword.
type NotNode struct {
	Child FilterNode
}

func (n *NotNode) Evaluate(summary backend.Summary) bool {
	if n.Child == nil {
		return true // NOT applied to nothing? Interpret as true.
	}
	return !n.Child.Evaluate(summary)
}

func (n *NotNode) String() string {
	childStr := "nil"
	if n.Child != nil {
		childStr = n.Child.String()
	}
	return fmt.Sprintf("NOT %s", childStr)
}

type parserState int

const (
	StateExpectTerm     parserState = iota // Expecting a term (foo:bar or -foo:bar) or NOT
	StateExpectOperator                    // Expecting an operator (AND) or end of input
)

// ParseFilter parses the filter string into a FilterNode tree using a state machine.
func ParseFilter(filterString string) (FilterNode, error) {
	filterString = strings.TrimSpace(filterString)
	if filterString == "" {
		return nil, nil
	}

	tokens := strings.Fields(filterString)
	if len(tokens) == 0 {
		return nil, nil
	}

	nodes := make([]FilterNode, 0)
	state := StateExpectTerm
	i := 0

	for i < len(tokens) {
		token := tokens[i]
		tokenUpper := strings.ToUpper(token)

		switch state {
		case StateExpectTerm:
			if tokenUpper == "AND" {
				return nil, fmt.Errorf("syntax error: unexpected 'AND' at token %d, expected term or NOT", i)
			} else if tokenUpper == "NOT" {
				// --- Handle NOT ---
				i++ // Consume NOT
				if i >= len(tokens) {
					return nil, fmt.Errorf("syntax error: NOT must be followed by a term")
				}
				termStr := tokens[i]
				termNode, err := parseTerm(termStr)
				if err != nil {
					return nil, fmt.Errorf("error parsing term after NOT ('%s'): %w", termStr, err)
				}
				if termNode.Negated { // Check term itself isn't negated e.g. NOT -term
					return nil, fmt.Errorf("syntax error: cannot apply NOT to already negated term ('%s')", termStr)
				}
				nodes = append(nodes, &NotNode{Child: termNode})
				state = StateExpectOperator // After NOT term, expect operator or end
				i++                         // Consume term
			} else {
				// --- Handle Term ---
				termNode, err := parseTerm(token)
				if err != nil {
					return nil, fmt.Errorf("error parsing term ('%s'): %w", token, err)
				}
				nodes = append(nodes, termNode)
				state = StateExpectOperator // After term, expect operator or end
				i++                         // Consume term
			}

		case StateExpectOperator:
			// We are expecting an operator (AND) or the end of the expression.
			// However, we also allow an implicit AND if a term (or NOT term) appears.
			tokenUpper := strings.ToUpper(token) // Re-check current token

			if tokenUpper == "AND" {
				// --- Handle Explicit AND ---
				state = StateExpectTerm // After AND, expect term or NOT
				i++                     // Consume AND
			} else if tokenUpper == "NOT" {
				// --- Handle Implicit AND followed by NOT ---
				i++                   // Consume NOT
				if i >= len(tokens) { // Ensure there's a term after NOT
					return nil, fmt.Errorf("syntax error: NOT must be followed by a term (after implicit AND)")
				}
				termStr := tokens[i]
				termNode, err := parseTerm(termStr)
				if err != nil {
					return nil, fmt.Errorf("error parsing term after implicit AND + NOT ('%s'): %w", termStr, err)
				}
				if termNode.Negated {
					return nil, fmt.Errorf("syntax error: cannot apply NOT to already negated term ('%s') after implicit AND", termStr)
				}
				nodes = append(nodes, &NotNode{Child: termNode})
				state = StateExpectOperator // Still expect operator or end
				i++                         // Consume the term after NOT
			} else {
				// --- Handle Implicit AND followed by Term ---
				termNode, err := parseTerm(token)
				if err != nil {
					// If it wasn't AND or NOT, and it's not a valid term, it's a syntax error.
					return nil, fmt.Errorf("syntax error: unexpected token '%s' at token %d, expected 'AND', 'NOT', or term", token, i)
				}
				nodes = append(nodes, termNode)
				state = StateExpectOperator // Still expect operator or end
				i++                         // Consume term
			}
		}
	}

	// Final check: ensure we didn't end expecting a term (e.g. after AND or NOT)
	if state == StateExpectTerm {
		// This implies the input ended with AND or NOT without a following term
		if len(tokens) > 0 && (strings.ToUpper(tokens[len(tokens)-1]) == "AND" || strings.ToUpper(tokens[len(tokens)-1]) == "NOT") {
			return nil, fmt.Errorf("syntax error: expression cannot end with '%s'", tokens[len(tokens)-1])
		} else if len(tokens) > 0 {
			// It could also mean the *only* token was NOT, handled above, or an internal error
			return nil, fmt.Errorf("internal parser error: ended expecting term unexpectedly")
		}
		// If tokens is empty, it's okay to be in StateExpectTerm initially
	}

	if len(nodes) == 0 {
		return nil, nil
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	// If multiple nodes were parsed, they are implicitly ANDed together by the structure
	return &AndNode{Children: nodes}, nil
}

// parseTerm is a helper to parse a single key:value or -key:value string into a TermNode
func parseTerm(termStr string) (*TermNode, error) {
	negated := false
	originalTermStr := termStr // Keep original for error messages
	if strings.HasPrefix(termStr, "-") {
		if len(termStr) == 1 {
			return nil, fmt.Errorf("invalid term format: '-' must be followed by key:value")
		}
		negated = true
		termStr = termStr[1:]
	}

	parts := strings.SplitN(termStr, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid term format: '%s'. Expected 'key:value' or 'key:*'", originalTermStr)
	}

	key := strings.ToLower(parts[0])
	value := parts[1]
	isTag := (key == "tag")

	termNode := &TermNode{
		Key:      key,
		Value:    value,
		Negated:  negated,
		IsTagKey: isTag,
	}
	return termNode, nil
}

// TODO: Implement parser for OR keyword and precedence (parentheses).
// TODO: Handle quoted values to allow spaces/colons within values.
// TODO: Add proper error handling and reporting (e.g., track token position).
// TODO: Integrate with slog logger (replace fmt.Printf warnings).
