package main

import (
	"reflect"
	"testing"
)

func TestExtractDomainsFromRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		expected []string
	}{
		{
			name:     "Single host with backticks",
			rule:     "Host(`app.example.com`)",
			expected: []string{"app.example.com"},
		},
		{
			name:     "Single host with double quotes",
			rule:     `Host("dashboard.dokploy.com")`,
			expected: []string{"dashboard.dokploy.com"},
		},
		{
			name:     "Multiple hosts in one Host clause",
			rule:     "Host(`a.example.com`, `b.example.com`)",
			expected: []string{"a.example.com", "b.example.com"},
		},
		{
			name:     "OR condition with multiple Host clauses",
			rule:     "Host(`app.example.com`) || Host(`app.internal.net`)",
			expected: []string{"app.example.com", "app.internal.net"},
		},
		{
			name:     "Host with PathPrefix AND condition",
			rule:     "Host(`api.domain.com`) && PathPrefix(`/v1`)",
			expected: []string{"api.domain.com"},
		},
		{
			name:     "PathPrefix only (no Host)",
			rule:     "PathPrefix(`/metrics`)",
			expected: nil,
		},
		{
			name:     "Empty rule",
			rule:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDomainsFromRule(tt.rule)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractDomainsFromRule(%q) = %v, want %v", tt.rule, got, tt.expected)
			}
		})
	}
}

func TestFormatAppName(t *testing.T) {
	tests := []struct {
		routerName  string
		serviceName string
		fileName    string
		expected    string
	}{
		{"my-app-router", "my-app-service", "my-app.yml", "My App"},
		{"dokploy-hub-secure", "", "hub.yml", "Dokploy Hub"},
		{"api_v1_service", "", "api.yaml", "Api V1 Service"},
	}

	for _, tt := range tests {
		got := formatAppName(tt.routerName, tt.serviceName, tt.fileName)
		if got != tt.expected {
			t.Errorf("formatAppName(%q, %q, %q) = %q, want %q", tt.routerName, tt.serviceName, tt.fileName, got, tt.expected)
		}
	}
}
