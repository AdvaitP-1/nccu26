// Package httpapi provides REST API handlers for the git state and
// execution service.
//
// These endpoints mirror the MCP tools but are exposed as plain HTTP
// for direct integration and debugging.
//
// Routes:
//
//	GET  /git/health                          → subsystem health
//	POST /git/pushes                          → register a push
//	GET  /git/branches/{branch}/files/{path}  → file state on a branch
//	POST /git/merge/context                   → prepare merge context
//	POST /git/merge/apply                     → apply merge result
//	POST /git/commit/prepare                  → prepare commit (dry run)
//	POST /git/commit                          → create grouped commit
//	POST /git/push                            → push commit to remote
//	GET  /git/commits/{id}                    → commit status
package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
	"github.com/nccuhacks/nccu26/mcp/internal/service"
)

// Handler holds the HTTP handlers for the git API.
type Handler struct {
	svc    *service.GitService
	logger *slog.Logger
}

// NewHandler creates a Handler backed by the given GitService.
func NewHandler(svc *service.GitService) *Handler {
	return &Handler{
		svc:    svc,
		logger: slog.Default().With("component", "httpapi"),
	}
}

// RegisterRoutes adds all git API routes to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /git/health", h.Health)
	mux.HandleFunc("POST /git/pushes", h.RegisterPush)
	mux.HandleFunc("GET /git/branches/{branch}/files/{path...}", h.GetBranchFileState)
	mux.HandleFunc("POST /git/merge/context", h.PrepareMergeContext)
	mux.HandleFunc("POST /git/merge/apply", h.ApplyMergeResult)
	mux.HandleFunc("POST /git/commit/prepare", h.PrepareCommit)
	mux.HandleFunc("POST /git/commit", h.CreateCommit)
	mux.HandleFunc("POST /git/push", h.PushCommit)
	mux.HandleFunc("GET /git/commits/{id}", h.GetCommitStatus)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	status := h.svc.Health(r.Context())
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) RegisterPush(w http.ResponseWriter, r *http.Request) {
	var req models.IngestPushRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.IngestPush(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetBranchFileState(w http.ResponseWriter, r *http.Request) {
	branch := r.PathValue("branch")
	filePath := r.PathValue("path")

	if branch == "" || filePath == "" {
		writeError(w, http.StatusBadRequest, "branch and file path are required")
		return
	}

	resp, err := h.svc.GetFileState(branch, filePath)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) PrepareMergeContext(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BranchName string `json:"branch_name"`
		FilePath   string `json:"file_path"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.PrepareMergeContext(req.BranchName, req.FilePath)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ApplyMergeResult(w http.ResponseWriter, r *http.Request) {
	var req models.ApplyMergeResultRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.ApplyMergeResult(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) PrepareCommit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PushID string `json:"push_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.PrepareCommit(req.PushID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateCommit(w http.ResponseWriter, r *http.Request) {
	var req models.GroupedCommitRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.CreateGroupedCommit(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) PushCommit(w http.ResponseWriter, r *http.Request) {
	var req models.GitPushRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.PushCommit(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetCommitStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "commit id is required")
		return
	}

	record, err := h.svc.GetCommitRecord(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

func readJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
