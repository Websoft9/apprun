package handlers

import (
	"net/http"

	"apprun/pkg/response"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"apprun"`
}

// HealthHandler handles health check requests.
//
//	@Summary		Health Check
//	@Description	Check if the service is running
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:  "ok",
		Service: "apprun",
	}
	response.Success(w, resp)
}
