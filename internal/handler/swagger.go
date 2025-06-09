package handler

import (
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pt-xyz-multifinance/pkg/logger"
	"go.uber.org/zap"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

// combineSwaggerFiles combines multiple swagger JSON files into one
func combineSwaggerFiles(files []string) ([]byte, error) {
	log := logger.GetLogger()

	// Base swagger structure
	combined := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":   "PT XYZ Multifinance API",
			"version": "1.0",
		},
		"paths":       map[string]interface{}{},
		"definitions": map[string]interface{}{},
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Error("Failed to read swagger file", zap.String("file", file), zap.Error(err))
			continue
		}

		var swagger map[string]interface{}
		if err := json.Unmarshal(data, &swagger); err != nil {
			log.Error("Failed to parse swagger file", zap.String("file", file), zap.Error(err))
			continue
		}

		// Merge paths
		if paths, ok := swagger["paths"].(map[string]interface{}); ok {
			combinedPaths := combined["paths"].(map[string]interface{})
			for k, v := range paths {
				combinedPaths[k] = v
			}
		}

		// Merge definitions
		if definitions, ok := swagger["definitions"].(map[string]interface{}); ok {
			combinedDefs := combined["definitions"].(map[string]interface{})
			for k, v := range definitions {
				combinedDefs[k] = v
			}
		}
	}

	return json.MarshalIndent(combined, "", "  ")
}

// SwaggerHandler returns a handler that serves the Swagger UI
func SwaggerHandler(swaggerFiles []string) http.Handler {
	// Generate combined swagger on startup
	combinedSwagger, err := combineSwaggerFiles(swaggerFiles)
	if err != nil {
		logger.GetLogger().Error("Failed to combine swagger files", zap.Error(err))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Remove the leading slash and clean the path
		urlPath := strings.TrimPrefix(r.URL.Path, "/")
		if urlPath == "" {
			urlPath = "swagger-ui/index.html"
		}

		// Serve the combined Swagger JSON file
		if urlPath == "swagger.json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write(combinedSwagger)
			return
		}

		// Add swagger-ui prefix if not present and not already there
		if !strings.HasPrefix(urlPath, "swagger-ui/") {
			urlPath = "swagger-ui/" + urlPath
		}

		// Set proper Content-Type headers
		ext := path.Ext(urlPath)
		switch ext {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		}

		log.Info("Serving file", zap.String("path", urlPath))

		content, err := swaggerUI.ReadFile(urlPath)
		if err != nil {
			log.Error("Failed to read swagger file", zap.String("path", urlPath), zap.Error(err))
			http.NotFound(w, r)
			return
		}

		w.Write(content)
	})
}
