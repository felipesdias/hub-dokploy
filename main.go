package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

type PageData struct {
	Apps      []App
	UpdatedTime string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8007"
	}

	configDir := os.Getenv("DYNAMIC_CONFIG_DIR")
	if configDir == "" {
		configDir = "/etc/dokploy/traefik/dynamic"
	}

	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"slice": func(s string, i, j int) string {
			if len(s) == 0 {
				return ""
			}
			if j > len(s) {
				j = len(s)
			}
			if i >= j {
				return s[:1]
			}
			return s[i:j]
		},
		"lower": strings.ToLower,
	}).ParseFS(templateFS, "templates/index.html")

	if err != nil {
		log.Fatalf("Failed to parse embedded template: %v", err)
	}

	// Handler for dashboard HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		apps, err := ParseDirectory(configDir)
		if err != nil {
			log.Printf("Error scanning directory %s: %v", configDir, err)
		}

		data := PageData{
			Apps:        apps,
			UpdatedTime: time.Now().Format("15:04:05"),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing template: %v", err)
		}
	})

	// Handler for JSON API
	http.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		apps, err := ParseDirectory(configDir)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apps)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	fmt.Printf("🚀 Dokploy Apps Hub listening on port :%s\n", port)
	fmt.Printf("📁 Monitoring Traefik config directory: %s\n", configDir)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
