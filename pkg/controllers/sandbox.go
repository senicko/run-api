package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/senicko/run-api/pkg/sandbox"
	"github.com/senicko/run-api/pkg/services"
	"go.uber.org/zap"
)

type SandboxController struct {
	sandboxService *services.SandboxService
	logger         *zap.Logger
}

// NewSandboxController creates a new sandbox controller.
func NewSandboxController(logger *zap.Logger, sandboxService *services.SandboxService) *SandboxController {
	return &SandboxController{
		logger:         logger,
		sandboxService: sandboxService,
	}
}

func (c *SandboxController) Run(w http.ResponseWriter, r *http.Request) {
	var execRequest sandbox.ExecRequest

	if err := json.NewDecoder(r.Body).Decode(&execRequest); err != nil {
		http.Error(w, "Failed to decode the exec request", http.StatusBadRequest)
		return
	}

	result, err := c.sandboxService.ProcessRequest(r.Context(), execRequest)
	if err != nil {
		http.Error(w, "Failed to execute", http.StatusInternalServerError)
		return
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJson)
}
