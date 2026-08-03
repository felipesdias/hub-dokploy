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

func TestExtractAppNameFromDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"n8n.felipesdias.com.br", "N8n"},
		{"finance-api.felipesdias.com.br", "Finance Api"},
		{"dokploy.felipesdias.com.br", "Dokploy"},
		{"nutri.felipesdias.com.br", "Nutri"},
	}

	for _, tt := range tests {
		got := extractAppNameFromDomain(tt.domain)
		if got != tt.expected {
			t.Errorf("extractAppNameFromDomain(%q) = %q, want %q", tt.domain, got, tt.expected)
		}
	}
}

func TestDeduplicateAppsWebsecurePriority(t *testing.T) {
	rawApps := []App{
		{
			Name:        "N8n",
			Domains:     []string{"n8n.felipesdias.com.br"},
			RouterName:  "n8n-jwutte-router-1",
			ServiceName: "n8n-jwutte-service-1",
			IsWebsecure: false, // web router
		},
		{
			Name:        "N8n",
			Domains:     []string{"n8n.felipesdias.com.br"},
			RouterName:  "n8n-jwutte-router-websecure-1",
			ServiceName: "n8n-jwutte-service-1",
			IsWebsecure: true, // websecure router
		},
	}

	deduped := deduplicateApps(rawApps)

	if len(deduped) != 1 {
		t.Fatalf("expected 1 deduplicated app, got %d", len(deduped))
	}

	if !deduped[0].IsWebsecure {
		t.Errorf("expected websecure router to be prioritized, but got IsWebsecure=false")
	}

	if deduped[0].Name != "N8n" || deduped[0].Domains[0] != "n8n.felipesdias.com.br" {
		t.Errorf("unexpected app content: %+v", deduped[0])
	}
}
