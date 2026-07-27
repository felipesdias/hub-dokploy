package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// TraefikConfig structure matching dynamic configuration files
type TraefikConfig struct {
	HTTP struct {
		Routers map[string]struct {
			EntryPoints []string    `yaml:"entryPoints" json:"entryPoints"`
			Rule        string      `yaml:"rule" json:"rule"`
			Service     string      `yaml:"service" json:"service"`
			TLS         interface{} `yaml:"tls" json:"tls"`
		} `yaml:"routers" json:"routers"`
		Services map[string]interface{} `yaml:"services" json:"services"`
	} `yaml:"http" json:"http"`
}

// App represents a deployed application extracted from Traefik configs
type App struct {
	Name        string   `json:"name"`
	Domains     []string `json:"domains"`
	URLs        []string `json:"urls"`
	Protocol    string   `json:"protocol"`
	RouterName  string   `json:"routerName"`
	ServiceName string   `json:"serviceName"`
	FileName    string   `json:"fileName"`
	HasTLS      bool     `json:"hasTls"`
}

var (
	// Matches Host(...) in Traefik rules, e.g. Host(`example.com`), Host("a.com", "b.com")
	hostRegex = regexp.MustCompile(`Host\s*\(\s*([^)]+)\s*\)`)
	// Cleans quotes and backticks
	quoteRegex = regexp.MustCompile("[\"`']")
)

// ParseDirectory scans the given directory for Traefik yaml/json files and returns all detected Apps with domains.
func ParseDirectory(dirPath string) ([]App, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []App{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var apps []App

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yml" && ext != ".yaml" && ext != ".json" {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		parsedApps, err := parseConfigFile(filePath, entry.Name())
		if err != nil {
			// Log error or skip unparseable file silently
			fmt.Printf("Warning: failed to parse %s: %v\n", filePath, err)
			continue
		}

		apps = append(apps, parsedApps...)
	}

	// Sort apps alphabetically by name
	sort.Slice(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})

	return apps, nil
}

// parseConfigFile reads a single config file and extracts applications
func parseConfigFile(filePath, fileName string) ([]App, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config TraefikConfig
	ext := strings.ToLower(filepath.Ext(fileName))

	if ext == ".json" {
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}

	var apps []App

	for routerName, router := range config.HTTP.Routers {
		domains := extractDomainsFromRule(router.Rule)
		if len(domains) == 0 {
			continue
		}

		hasTLS := router.TLS != nil || hasWebsecureEntryPoint(router.EntryPoints)
		protocol := "http"
		if hasTLS {
			protocol = "https"
		}

		urls := make([]string, len(domains))
		for i, d := range domains {
			urls[i] = fmt.Sprintf("%s://%s", protocol, d)
		}

		appName := formatAppName(routerName, router.Service, fileName)

		apps = append(apps, App{
			Name:        appName,
			Domains:     domains,
			URLs:        urls,
			Protocol:    protocol,
			RouterName:  routerName,
			ServiceName: router.Service,
			FileName:    fileName,
			HasTLS:      hasTLS,
		})
	}

	return apps, nil
}

// extractDomainsFromRule extracts domain names from Traefik Host(...) rule strings
func extractDomainsFromRule(rule string) []string {
	matches := hostRegex.FindAllStringSubmatch(rule, -1)
	if len(matches) == 0 {
		return nil
	}

	domainMap := make(map[string]bool)
	var domains []string

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		// raw inside Host(...) e.g. `domain1.com`, `domain2.com`
		rawContent := match[1]
		parts := strings.Split(rawContent, ",")
		for _, p := range parts {
			cleanDomain := strings.TrimSpace(quoteRegex.ReplaceAllString(p, ""))
			if cleanDomain != "" && !domainMap[cleanDomain] {
				domainMap[cleanDomain] = true
				domains = append(domains, cleanDomain)
			}
		}
	}

	return domains
}

// hasWebsecureEntryPoint checks if websecure or https entrypoint is set
func hasWebsecureEntryPoint(entryPoints []string) bool {
	for _, ep := range entryPoints {
		epLower := strings.ToLower(ep)
		if epLower == "websecure" || epLower == "https" {
			return true
		}
	}
	return false
}

// formatAppName generates a clean human-readable application name
func formatAppName(routerName, serviceName, fileName string) string {
	raw := routerName
	if raw == "" {
		raw = serviceName
	}
	if raw == "" {
		raw = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	// Remove common traefik suffixes
	raw = strings.TrimSuffix(raw, "-router")
	raw = strings.TrimSuffix(raw, "-rtr")
	raw = strings.TrimSuffix(raw, "-secure")
	raw = strings.TrimSuffix(raw, "-web")

	// Format dashes/underscores to spaces or clean name
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.Title(part)
		}
	}

	formatted := strings.Join(parts, " ")
	if formatted == "" {
		return raw
	}
	return formatted
}
