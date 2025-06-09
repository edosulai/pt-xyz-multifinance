package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	logger *zap.Logger
}

func NewHTTPHandler(logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		logger: logger,
	}
}

// RegisterHTTPRoutes registers additional HTTP routes not covered by gRPC-Gateway
func (h *HTTPHandler) RegisterHTTPRoutes(router *mux.Router) {
	// Captcha routes
	router.HandleFunc("/v1/captcha/new", h.handleNewCaptcha).Methods(http.MethodGet)
	router.HandleFunc("/v1/captcha/{id}.png", h.handleCaptchaImage).Methods(http.MethodGet)

	// Health check
	router.HandleFunc("/health", h.handleHealthCheck).Methods(http.MethodGet)
}

func (h *HTTPHandler) handleNewCaptcha(w http.ResponseWriter, r *http.Request) {
	id := captcha.New()
	response := map[string]string{"captcha_id": id}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode captcha response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) handleCaptchaImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	w.Header().Set("Content-Type", "image/png")
	if err := captcha.WriteImage(w, id, 240, 80); err != nil {
		h.logger.Error("Failed to write captcha image", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
