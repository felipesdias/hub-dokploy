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
		// "my-cool-app-12345-router-websecure".split('-router-')[0] -> "my-cool-app-12345".split('-') -> remove "12345" -> "My Cool App"
		{"my-cool-app-12345-router-websecure", "", "app.yml", "My Cool App"},
		{"dokploy-hub-xyz-router-1", "", "hub.yml", "Dokploy Hub"},
		{"api-service-abc-router", "", "api.yaml", "Api Service"},
		{"grafana-router", "", "grafana.yaml", "Grafana"},
	}

	for _, tt := range tests {
		got := formatAppName(tt.routerName, tt.serviceName, tt.fileName)
		if got != tt.expected {
			t.Errorf("formatAppName(%q, %q, %q) = %q, want %q", tt.routerName, tt.serviceName, tt.fileName, got, tt.expected)
		}
	}
}
