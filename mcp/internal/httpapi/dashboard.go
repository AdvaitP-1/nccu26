package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/nccuhacks/nccu26/mcp/internal/agents"
	"github.com/nccuhacks/nccu26/mcp/internal/analysisclient"
	"github.com/nccuhacks/nccu26/mcp/internal/commits"
	"github.com/nccuhacks/nccu26/mcp/internal/events"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
	"github.com/nccuhacks/nccu26/mcp/internal/policy"
	"github.com/nccuhacks/nccu26/mcp/internal/service"
	"github.com/nccuhacks/nccu26/mcp/internal/vfs"
)

// DashboardHandler exposes VFS, command, SSE, and agent-registry
// endpoints for the PM dashboard.
type DashboardHandler struct {
	vfs      *vfs.Manager
	analysis *analysisclient.Client
	policy   *policy.Evaluator
	commits  *commits.Coordinator
	gitSvc   *service.GitService
	registry *agents.Registry
	bus      *events.Bus
	logger   *slog.Logger
}

// NewDashboardHandler creates a handler with all dependencies.
func NewDashboardHandler(
	v *vfs.Manager,
	a *analysisclient.Client,
	p *policy.Evaluator,
	c *commits.Coordinator,
	g *service.GitService,
	r *agents.Registry,
	b *events.Bus,
) *DashboardHandler {
	return &DashboardHandler{
		vfs:      v,
		analysis: a,
		policy:   p,
		commits:  c,
		gitSvc:   g,
		registry: r,
		bus:      b,
		logger:   slog.Default().With("component", "dashboard-api"),
	}
}

// RegisterRoutes adds all dashboard endpoints to the mux.
func (h *DashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /dashboard/vfs", h.GetVFS)
	mux.HandleFunc("POST /dashboard/command", h.RunCommand)
	mux.HandleFunc("GET /dashboard/events", h.SSE)
	mux.HandleFunc("GET /dashboard/agents", h.GetAgents)
}

// ---------------------------------------------------------------------------
// GET /dashboard/vfs
// ---------------------------------------------------------------------------

func (h *DashboardHandler) GetVFS(w http.ResponseWriter, _ *http.Request) {
	state := h.vfs.State()
	writeJSON(w, http.StatusOK, state)
}

// ---------------------------------------------------------------------------
// GET /dashboard/agents
// ---------------------------------------------------------------------------

func (h *DashboardHandler) GetAgents(w http.ResponseWriter, _ *http.Request) {
	all := h.registry.All()
	writeJSON(w, http.StatusOK, all)
}

// ---------------------------------------------------------------------------
// GET /dashboard/events  (SSE)
// ---------------------------------------------------------------------------

func (h *DashboardHandler) SSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sub := h.bus.Subscribe()
	defer h.bus.Unsubscribe(sub)

	h.logger.Info("SSE client connected", "subscribers", h.bus.SubscriberCount())

	// Send initial keepalive
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Keepalive ticker so proxies don't close idle connections
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			h.logger.Info("SSE client disconnected")
			return
		case evt := <-sub.Ch:
			data := evt.JSON()
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		case <-keepalive.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// ---------------------------------------------------------------------------
// POST /dashboard/command
// ---------------------------------------------------------------------------

type commandRequest struct {
	Command string         `json:"command"`
	Args    map[string]any `json:"args"`
}

type commandResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *DashboardHandler) RunCommand(w http.ResponseWriter, r *http.Request) {
	var req commandRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, commandResponse{
			Error: "invalid request body: " + err.Error(),
		})
		return
	}

	h.logger.Info("dashboard command", "command", req.Command)

	switch req.Command {
	case "get_vfs_state":
		h.cmdGetVFS(w)
	case "identify_overlaps":
		h.cmdIdentifyOverlaps(w)
	case "request_micro_commit":
		h.cmdMicroCommit(w, req.Args)
	case "clear_agent_state":
		h.cmdClearAgent(w, req.Args)
	case "git_health":
		h.cmdGitHealth(w, r)
	case "register_push":
		h.cmdRegisterPush(w, req.Args)
	case "seed_demo":
		h.cmdSeedDemo(w)
	case "clear_demo":
		h.cmdClearDemo(w)
	case "register_agent":
		h.cmdRegisterAgent(w, req.Args)
	case "list_agents":
		h.cmdListAgents(w)
	case "propose_files":
		h.cmdProposeFiles(w, req.Args)
	default:
		writeJSON(w, http.StatusBadRequest, commandResponse{
			Error: fmt.Sprintf("unknown command: %q", req.Command),
		})
	}

	h.bus.Publish(events.TypeCommandResult, map[string]string{
		"command": req.Command,
	})
}

// ---------------------------------------------------------------------------
// Command implementations
// ---------------------------------------------------------------------------

func (h *DashboardHandler) cmdGetVFS(w http.ResponseWriter) {
	state := h.vfs.State()
	writeJSON(w, http.StatusOK, commandResponse{Success: true, Data: state})
}

func (h *DashboardHandler) cmdIdentifyOverlaps(w http.ResponseWriter) {
	changesets := h.vfs.ChangeSetsForAnalysis()
	if len(changesets) < 2 {
		writeJSON(w, http.StatusOK, commandResponse{
			Success: true,
			Data: map[string]any{
				"overlaps":   []any{},
				"file_risks": []any{},
				"note":       "fewer than 2 agents — nothing to compare",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := h.analysis.AnalyzeOverlaps(ctx, changesets)
	if err != nil {
		writeJSON(w, http.StatusOK, commandResponse{
			Error: "analysis failed: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, commandResponse{Success: true, Data: resp})
}

func (h *DashboardHandler) cmdMicroCommit(w http.ResponseWriter, args map[string]any) {
	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		writeJSON(w, http.StatusOK, commandResponse{Error: "agent_id is required"})
		return
	}

	message, _ := args["message"].(string)
	if message == "" {
		message = fmt.Sprintf("micro-commit by %s", agentID)
	}

	files, err := h.vfs.FilesForAgent(agentID)
	if err != nil {
		writeJSON(w, http.StatusOK, commandResponse{Error: err.Error()})
		return
	}

	changesets := h.vfs.ChangeSetsForAnalysis()
	var analysisResp *models.AnalyzeOverlapsResponse
	if len(changesets) >= 2 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		analysisResp, err = h.analysis.AnalyzeOverlaps(ctx, changesets)
		if err != nil {
			writeJSON(w, http.StatusOK, commandResponse{Error: "analysis failed: " + err.Error()})
			return
		}
	}

	decision := h.policy.Evaluate(analysisResp)
	if !decision.Allowed {
		writeJSON(w, http.StatusOK, commandResponse{Success: false, Data: decision})
		return
	}

	result := h.commits.Commit(models.CommitRequest{
		AgentID: agentID,
		Files:   files,
		Message: message,
	})

	if result.Allowed {
		h.vfs.Clear(agentID)
		h.registry.SetStatus(agentID, agents.StatusIdle)
		h.bus.Publish(events.TypeVFSUpdate, h.vfs.State())
	}

	writeJSON(w, http.StatusOK, commandResponse{Success: true, Data: result})
}

func (h *DashboardHandler) cmdClearAgent(w http.ResponseWriter, args map[string]any) {
	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		writeJSON(w, http.StatusOK, commandResponse{Error: "agent_id is required"})
		return
	}

	h.vfs.Clear(agentID)
	h.registry.SetStatus(agentID, agents.StatusIdle)
	h.bus.Publish(events.TypeVFSUpdate, h.vfs.State())

	writeJSON(w, http.StatusOK, commandResponse{
		Success: true,
		Data:    map[string]string{"cleared": agentID},
	})
}

func (h *DashboardHandler) cmdGitHealth(w http.ResponseWriter, r *http.Request) {
	if h.gitSvc == nil {
		writeJSON(w, http.StatusOK, commandResponse{
			Success: true,
			Data:    map[string]string{"status": "no git service configured"},
		})
		return
	}

	status := h.gitSvc.Health(r.Context())
	writeJSON(w, http.StatusOK, commandResponse{Success: true, Data: status})
}

func (h *DashboardHandler) cmdRegisterPush(w http.ResponseWriter, args map[string]any) {
	branchName, _ := args["branch_name"].(string)
	userID, _ := args["user_id"].(string)
	filesJSON, _ := args["files_json"].(string)
	message, _ := args["message"].(string)

	if branchName == "" || userID == "" || filesJSON == "" {
		writeJSON(w, http.StatusOK, commandResponse{
			Error: "branch_name, user_id, and files_json are required",
		})
		return
	}

	var files []models.PushFileChange
	if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
		writeJSON(w, http.StatusOK, commandResponse{Error: "invalid files_json: " + err.Error()})
		return
	}

	req := models.IngestPushRequest{
		BranchName: branchName,
		UserID:     userID,
		Files:      files,
		Message:    message,
	}

	resp, err := h.gitSvc.IngestPush(context.Background(), req)
	if err != nil {
		writeJSON(w, http.StatusOK, commandResponse{Error: err.Error()})
		return
	}

	h.registry.Touch(userID)
	h.bus.Publish(events.TypeVFSUpdate, h.vfs.State())

	writeJSON(w, http.StatusOK, commandResponse{Success: true, Data: resp})
}

// ---------------------------------------------------------------------------
// Agent registry commands
// ---------------------------------------------------------------------------

func (h *DashboardHandler) cmdRegisterAgent(w http.ResponseWriter, args map[string]any) {
	id, _ := args["id"].(string)
	agentType, _ := args["type"].(string)
	displayName, _ := args["display_name"].(string)

	if id == "" {
		writeJSON(w, http.StatusOK, commandResponse{Error: "id is required"})
		return
	}
	if agentType == "" {
		agentType = "coder"
	}
	if displayName == "" {
		displayName = id
	}

	h.registry.Register(id, agents.Type(agentType), displayName, nil)
	h.bus.Publish(events.TypeAgentRegistered, map[string]string{"id": id, "type": agentType})

	writeJSON(w, http.StatusOK, commandResponse{
		Success: true,
		Data:    h.registry.Get(id),
	})
}

func (h *DashboardHandler) cmdListAgents(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, commandResponse{
		Success: true,
		Data:    h.registry.All(),
	})
}

func (h *DashboardHandler) cmdProposeFiles(w http.ResponseWriter, args map[string]any) {
	agentID, _ := args["agent_id"].(string)
	sessionID, _ := args["session_id"].(string)
	taskID, _ := args["task_id"].(string)
	filesRaw, _ := args["files"].([]any)

	if agentID == "" || len(filesRaw) == 0 {
		writeJSON(w, http.StatusOK, commandResponse{Error: "agent_id and files are required"})
		return
	}
	if sessionID == "" {
		sessionID = "swarm-session"
	}
	if taskID == "" {
		taskID = "swarm-task"
	}

	var files []models.FileSnapshot
	for _, raw := range filesRaw {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		path, _ := m["path"].(string)
		lang, _ := m["language"].(string)
		content, _ := m["content"].(string)
		if path == "" {
			continue
		}
		if lang == "" {
			lang = "python"
		}
		files = append(files, models.FileSnapshot{Path: path, Language: lang, Content: content})
	}

	if len(files) == 0 {
		writeJSON(w, http.StatusOK, commandResponse{Error: "no valid files provided"})
		return
	}

	h.vfs.Propose(agentID, sessionID, taskID, files)
	h.bus.Publish(events.TypeVFSUpdate, h.vfs.State())

	writeJSON(w, http.StatusOK, commandResponse{
		Success: true,
		Data: map[string]any{
			"agent_id":    agentID,
			"files_count": len(files),
		},
	})
}

// ---------------------------------------------------------------------------
// Demo seed / clear
// ---------------------------------------------------------------------------

func (h *DashboardHandler) cmdSeedDemo(w http.ResponseWriter) {
	h.seedDemoData()

	writeJSON(w, http.StatusOK, commandResponse{
		Success: true,
		Data: map[string]any{
			"seeded_agents": 5,
			"message":       "Demo data populated: 5 agents with realistic file changes",
		},
	})
}

func (h *DashboardHandler) cmdClearDemo(w http.ResponseWriter) {
	h.vfs.ClearAll()
	h.registry.Clear()
	h.bus.Publish(events.TypeVFSUpdate, h.vfs.State())
	h.bus.Publish(events.TypeAgentRemoved, map[string]string{"scope": "all"})

	writeJSON(w, http.StatusOK, commandResponse{
		Success: true,
		Data:    map[string]string{"message": "All demo data cleared"},
	})
}

// seedDemoData populates the VFS and agent registry with realistic
// multi-agent orchestration data suitable for a PM dashboard demo.
func (h *DashboardHandler) seedDemoData() {
	type seedAgent struct {
		id          string
		agentType   agents.Type
		displayName string
		sessionID   string
		taskID      string
		files       []models.FileSnapshot
	}

	demoAgents := []seedAgent{
		{
			id:          "arch-planner-01",
			agentType:   agents.TypeManager,
			displayName: "Architecture Planner",
			sessionID:   "sess-demo-001",
			taskID:      "TASK-100",
			files: []models.FileSnapshot{
				{Path: "docs/architecture/system-design.md", Language: "markdown", Content: "# System Architecture\n\n## Overview\nMicroservices-based API gateway with event-driven processing.\n\n## Components\n- API Gateway (Go)\n- Auth Service (Go)\n- Event Bus (Kafka)\n- Data Store (PostgreSQL)"},
				{Path: "docs/architecture/task-decomposition.yaml", Language: "yaml", Content: "tasks:\n  - id: TASK-101\n    title: Implement auth middleware\n    assignee: coder-auth\n    priority: high\n  - id: TASK-102\n    title: Build REST API handlers\n    assignee: coder-api\n    priority: high\n  - id: TASK-103\n    title: Add integration tests\n    assignee: coder-api\n    priority: medium"},
				{Path: "docs/architecture/api-contracts.json", Language: "json", Content: "{\n  \"openapi\": \"3.0.3\",\n  \"paths\": {\n    \"/api/v1/users\": {\"get\": {}, \"post\": {}},\n    \"/api/v1/auth/login\": {\"post\": {}},\n    \"/api/v1/auth/refresh\": {\"post\": {}}\n  }\n}"},
			},
		},
		{
			id:          "coder-auth",
			agentType:   agents.TypeCoder,
			displayName: "Auth Service Coder",
			sessionID:   "sess-demo-002",
			taskID:      "TASK-101",
			files: []models.FileSnapshot{
				{Path: "backend/auth/middleware.go", Language: "go", Content: "package auth\n\nimport (\n\t\"net/http\"\n\t\"strings\"\n)\n\nfunc JWTMiddleware(next http.Handler) http.Handler {\n\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\ttoken := strings.TrimPrefix(r.Header.Get(\"Authorization\"), \"Bearer \")\n\t\tif token == \"\" {\n\t\t\thttp.Error(w, \"unauthorized\", http.StatusUnauthorized)\n\t\t\treturn\n\t\t}\n\t\tnext.ServeHTTP(w, r)\n\t})\n}"},
				{Path: "backend/auth/tokens.go", Language: "go", Content: "package auth\n\nimport (\n\t\"crypto/rand\"\n\t\"encoding/hex\"\n\t\"time\"\n)\n\ntype TokenPair struct {\n\tAccessToken  string    `json:\"access_token\"`\n\tRefreshToken string    `json:\"refresh_token\"`\n\tExpiresAt    time.Time `json:\"expires_at\"`\n}\n\nfunc GenerateTokenPair(userID string) (*TokenPair, error) {\n\tbytes := make([]byte, 32)\n\t_, err := rand.Read(bytes)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn &TokenPair{\n\t\tAccessToken:  hex.EncodeToString(bytes),\n\t\tExpiresAt:    time.Now().Add(15 * time.Minute),\n\t}, nil\n}"},
				{Path: "backend/auth/models.go", Language: "go", Content: "package auth\n\nimport \"time\"\n\ntype User struct {\n\tID        string    `json:\"id\" db:\"id\"`\n\tEmail     string    `json:\"email\" db:\"email\"`\n\tPassHash  string    `json:\"-\" db:\"pass_hash\"`\n\tCreatedAt time.Time `json:\"created_at\" db:\"created_at\"`\n}"},
				{Path: "backend/auth/handler.go", Language: "go", Content: "package auth\n\nimport (\n\t\"encoding/json\"\n\t\"net/http\"\n)\n\ntype Handler struct {\n\t// service dependencies\n}\n\nfunc (h *Handler) Login(w http.ResponseWriter, r *http.Request) {\n\tvar req struct {\n\t\tEmail    string `json:\"email\"`\n\t\tPassword string `json:\"password\"`\n\t}\n\tif err := json.NewDecoder(r.Body).Decode(&req); err != nil {\n\t\thttp.Error(w, \"bad request\", http.StatusBadRequest)\n\t\treturn\n\t}\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(map[string]string{\"status\": \"ok\"})\n}"},
			{Path: "backend/shared/config.go", Language: "go", Content: "package shared\n\nimport (\n\t\"os\"\n\t\"strconv\"\n\t\"time\"\n)\n\ntype AppConfig struct {\n\tPort         int\n\tDatabaseURL  string\n\tJWTSecret    string\n\tTokenExpiry  time.Duration\n}\n\nfunc LoadConfig() *AppConfig {\n\tport, _ := strconv.Atoi(os.Getenv(\"PORT\"))\n\tif port == 0 {\n\t\tport = 8080\n\t}\n\treturn &AppConfig{\n\t\tPort:        port,\n\t\tDatabaseURL: os.Getenv(\"DATABASE_URL\"),\n\t\tJWTSecret:   os.Getenv(\"JWT_SECRET\"),\n\t\tTokenExpiry: 15 * time.Minute,\n\t}\n}\n\nfunc ValidateToken(token string) (string, error) {\n\tif token == \"\" {\n\t\treturn \"\", fmt.Errorf(\"empty token\")\n\t}\n\treturn \"user-id\", nil\n}"},
			},
		},
		{
			id:          "coder-api",
			agentType:   agents.TypeCoder,
			displayName: "API Endpoints Coder",
			sessionID:   "sess-demo-003",
			taskID:      "TASK-102",
			files: []models.FileSnapshot{
				{Path: "backend/api/routes.go", Language: "go", Content: "package api\n\nimport \"net/http\"\n\nfunc RegisterRoutes(mux *http.ServeMux) {\n\tmux.HandleFunc(\"GET /api/v1/users\", ListUsers)\n\tmux.HandleFunc(\"POST /api/v1/users\", CreateUser)\n\tmux.HandleFunc(\"GET /api/v1/users/{id}\", GetUser)\n\tmux.HandleFunc(\"PUT /api/v1/users/{id}\", UpdateUser)\n\tmux.HandleFunc(\"DELETE /api/v1/users/{id}\", DeleteUser)\n}"},
				{Path: "backend/api/users.go", Language: "go", Content: "package api\n\nimport (\n\t\"encoding/json\"\n\t\"net/http\"\n)\n\nfunc ListUsers(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode([]map[string]string{})\n}\n\nfunc CreateUser(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusCreated)\n}\n\nfunc GetUser(w http.ResponseWriter, r *http.Request) {\n\tid := r.PathValue(\"id\")\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(map[string]string{\"id\": id})\n}\n\nfunc UpdateUser(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusNoContent)\n}\n\nfunc DeleteUser(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusNoContent)\n}"},
				{Path: "backend/api/middleware.go", Language: "go", Content: "package api\n\nimport (\n\t\"log/slog\"\n\t\"net/http\"\n\t\"time\"\n)\n\nfunc LoggingMiddleware(next http.Handler) http.Handler {\n\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\tstart := time.Now()\n\t\tnext.ServeHTTP(w, r)\n\t\tslog.Info(\"request\", \"method\", r.Method, \"path\", r.URL.Path, \"duration\", time.Since(start))\n\t})\n}"},
			{Path: "backend/shared/config.go", Language: "go", Content: "package shared\n\nimport (\n\t\"os\"\n\t\"strconv\"\n\t\"time\"\n)\n\ntype AppConfig struct {\n\tPort         int\n\tDatabaseURL  string\n\tAPIKey       string\n\tRateLimit    int\n}\n\nfunc LoadConfig() *AppConfig {\n\tport, _ := strconv.Atoi(os.Getenv(\"PORT\"))\n\tif port == 0 {\n\t\tport = 3000\n\t}\n\treturn &AppConfig{\n\t\tPort:        port,\n\t\tDatabaseURL: os.Getenv(\"DATABASE_URL\"),\n\t\tAPIKey:      os.Getenv(\"API_KEY\"),\n\t\tRateLimit:   100,\n\t}\n}\n\nfunc ValidateToken(token string) (string, error) {\n\tif len(token) < 10 {\n\t\treturn \"\", fmt.Errorf(\"token too short\")\n\t}\n\treturn \"api-user\", nil\n}"},
			},
		},
		{
			id:          "merge-coordinator",
			agentType:   agents.TypeMerge,
			displayName: "Merge Coordinator",
			sessionID:   "sess-demo-004",
			taskID:      "TASK-200",
			files: []models.FileSnapshot{
				{Path: ".orca/merge-plan.yaml", Language: "yaml", Content: "merge_plan:\n  strategy: sequential\n  order:\n    - agent: coder-auth\n      branch: feat/auth-middleware\n      risk: low\n    - agent: coder-api\n      branch: feat/api-endpoints\n      risk: medium\n  conflicts_detected: 1\n  auto_resolvable: true"},
				{Path: ".orca/conflict-resolution.json", Language: "json", Content: "{\n  \"conflicts\": [\n    {\n      \"file\": \"backend/api/middleware.go\",\n      \"agents\": [\"coder-auth\", \"coder-api\"],\n      \"type\": \"structural_overlap\",\n      \"resolution\": \"merge_both\",\n      \"confidence\": 0.94\n    }\n  ]\n}"},
			},
		},
		{
			id:          "reviewer-01",
			agentType:   agents.TypeReviewer,
			displayName: "Code Reviewer",
			sessionID:   "sess-demo-005",
			taskID:      "TASK-300",
			files: []models.FileSnapshot{
				{Path: ".orca/review/coder-auth-review.md", Language: "markdown", Content: "# Code Review: coder-auth\n\n## Summary\nAuth middleware and token generation implementation.\n\n## Findings\n- [PASS] JWT middleware correctly validates bearer tokens\n- [WARN] Token generation uses crypto/rand (good) but no expiry validation on refresh\n- [PASS] Models follow Go conventions\n- [INFO] Consider adding rate limiting to login handler\n\n## Risk: LOW\n## Recommendation: APPROVE with minor suggestions"},
				{Path: ".orca/review/coder-api-review.md", Language: "markdown", Content: "# Code Review: coder-api\n\n## Summary\nREST API endpoints for user management.\n\n## Findings\n- [PASS] RESTful route structure\n- [WARN] No input validation on CreateUser/UpdateUser\n- [WARN] Missing pagination on ListUsers\n- [PASS] Logging middleware is clean\n\n## Risk: MEDIUM\n## Recommendation: APPROVE with required changes"},
			},
		},
	}

	for _, da := range demoAgents {
		h.registry.Register(da.id, da.agentType, da.displayName, map[string]string{
			"session": da.sessionID,
			"task":    da.taskID,
		})

		h.vfs.Propose(da.id, da.sessionID, da.taskID, da.files)
	}

	h.bus.Publish(events.TypeVFSUpdate, h.vfs.State())
	for _, da := range demoAgents {
		h.bus.Publish(events.TypeAgentRegistered, map[string]string{
			"id":   da.id,
			"type": string(da.agentType),
		})
	}
}
