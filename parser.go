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
	RouterName  string   `json:"routerName"`
	ServiceName string   `json:"serviceName"`
	FileName    string   `json:"fileName"`
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

		urls := make([]string, len(domains))
		for i, d := range domains {
			urls[i] = fmt.Sprintf("https://%s", d)
		}

		appName := formatAppName(routerName, router.Service, fileName)

		apps = append(apps, App{
			Name:        appName,
			Domains:     domains,
			URLs:        urls,
			RouterName:  routerName,
			ServiceName: router.Service,
			FileName:    fileName,
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

// formatAppName parses the app name matching the requested logic:
// parts = name.split('-router-')[0].split('-') -> remove last element if len > 1 -> join
func formatAppName(routerName, serviceName, fileName string) string {
	raw := routerName
	if raw == "" {
		raw = serviceName
	}
	if raw == "" {
		raw = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	// 1. split('-router-')[0]
	if idx := strings.Index(raw, "-router-"); idx != -1 {
		raw = raw[:idx]
	} else if idx := strings.Index(raw, "-router"); idx != -1 {
		raw = raw[:idx]
	}

	// 2. split('-')
	parts := strings.Split(raw, "-")

	// 3. Remove last part if there are multiple parts
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
	}

	// 4. Title case and join remaining parts
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.Title(part)
		}
	}

	return strings.Join(parts, " ")
}
