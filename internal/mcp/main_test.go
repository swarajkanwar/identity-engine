package main

import (
	"testing"
)

func TestQueryRagDatabase(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedMatch string
	}{
		{
			name:          "Exact match identity engine",
			query:         "what is identity-engine?",
			expectedMatch: "Identity Engine is a fast, robust gRPC service for managing user identities.",
		},
		{
			name:          "Partial match authenticate",
			query:         "authenticate",
			expectedMatch: "You authenticate by providing a valid JWT token in the Authorization header.",
		},
		{
			name:          "Partial match mcp",
			query:         "tell me about mcp",
			expectedMatch: "MCP stands for Model Context Protocol, allowing AI models to easily call local tools.",
		},
		{
			name:          "No match",
			query:         "how to cook pasta?",
			expectedMatch: "I could not find any relevant information in the knowledge base.",
		},
		{
			name:          "Case insensitive query",
			query:         "WHAT IS IDENTITY-ENGINE?",
			expectedMatch: "Identity Engine is a fast, robust gRPC service for managing user identities.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := queryRagDatabase(tt.query)
			if result != tt.expectedMatch {
				t.Errorf("queryRagDatabase(%q) = %q, want %q", tt.query, result, tt.expectedMatch)
			}
		})
	}
}
