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
	IsWebsecure bool     `json:"isWebsecure"`
}

var (
	// Matches Host(...) in Traefik rules, e.g. Host(`example.com`), Host("a.com", "b.com")
	hostRegex = regexp.MustCompile(`Host\s*\(\s*([^)]+)\s*\)`)
	// Cleans quotes and backticks
	quoteRegex = regexp.MustCompile("[\"`']")
)

// ParseDirectory scans the given directory for Traefik yaml/json files and returns deduplicated Apps.
func ParseDirectory(dirPath string) ([]App, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []App{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var rawApps []App

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

		rawApps = append(rawApps, parsedApps...)
	}

	// Deduplicate apps, prioritizing websecure routers and grouping by domain/app name
	apps := deduplicateApps(rawApps)

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

		isWebsecure := hasWebsecureEntryPoint(router.EntryPoints) || strings.Contains(strings.ToLower(routerName), "websecure")

		// App name extracted from the first domain's first part before '.'
		appName := extractAppNameFromDomain(domains[0])

		apps = append(apps, App{
			Name:        appName,
			Domains:     domains,
			URLs:        urls,
			RouterName:  routerName,
			ServiceName: router.Service,
			FileName:    fileName,
			IsWebsecure: isWebsecure,
		})
	}

	return apps, nil
}

// extractAppNameFromDomain splits the domain by '.' and takes the first part (subdomain)
// Example: "n8n.example.com" -> "N8n"
// Example: "finance-api.example.com" -> "Finance Api"
func extractAppNameFromDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) == 0 || parts[0] == "" {
		return domain
	}
	subdomain := parts[0]

	// Split by hyphens or underscores for human readable title formatting
	subParts := strings.FieldsFunc(subdomain, func(r rune) bool {
		return r == '-' || r == '_'
	})

	for i, p := range subParts {
		if len(p) > 0 {
			subParts[i] = strings.Title(p)
		}
	}

	if len(subParts) > 0 {
		return strings.Join(subParts, " ")
	}

	return strings.Title(subdomain)
}

// deduplicateApps prioritizes websecure routers and merges duplicate domains
func deduplicateApps(rawApps []App) []App {
	domainMap := make(map[string]App) // domain -> App
	var domainOrder []string

	for _, app := range rawApps {
		for _, d := range app.Domains {
			dLower := strings.ToLower(d)
			existing, found := domainMap[dLower]
			if !found {
				domainMap[dLower] = app
				domainOrder = append(domainOrder, dLower)
			} else {
				// If current router is websecure and existing is not, prioritize websecure!
				if !existing.IsWebsecure && app.IsWebsecure {
					domainMap[dLower] = app
				}
			}
		}
	}

	// Group apps sharing the same App Name into a single card
	appGroup := make(map[string]*App)
	var finalOrder []string

	for _, dLower := range domainOrder {
		app := domainMap[dLower]
		key := strings.ToLower(app.Name)

		if existing, found := appGroup[key]; found {
			for _, d := range app.Domains {
				if !containsDomain(existing.Domains, d) {
					existing.Domains = append(existing.Domains, d)
					existing.URLs = append(existing.URLs, "https://"+d)
				}
			}
			if app.IsWebsecure {
				existing.IsWebsecure = true
			}
		} else {
			appCopy := app
			appGroup[key] = &appCopy
			finalOrder = append(finalOrder, key)
		}
	}

	var result []App
	for _, key := range finalOrder {
		result = append(result, *appGroup[key])
	}

	return result
}

func containsDomain(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

func hasWebsecureEntryPoint(entryPoints []string) bool {
	for _, ep := range entryPoints {
		epLower := strings.ToLower(ep)
		if epLower == "websecure" || epLower == "https" {
			return true
		}
	}
	return false
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
