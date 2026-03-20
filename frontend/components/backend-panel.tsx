"use client"

import { useCallback, useEffect, useState } from "react"
import { fetchHealth, analyzeOverlaps, ApiClientError } from "@/lib/api"
import type {
  HealthResponse,
  AnalyzeOverlapsResponse,
  AnalyzeOverlapsRequest,
  OverlapSeverity,
} from "@/lib/types"
import { cn } from "@/lib/utils"

// ---------------------------------------------------------------------------
// Sample payload for the overlap analysis demo
// ---------------------------------------------------------------------------

const SAMPLE_PAYLOAD: AnalyzeOverlapsRequest = {
  changesets: [
    {
      agent_id: "agent-alpha",
      files: [
        {
          path: "src/auth.py",
          language: "python",
          content: [
            "class AuthService:",
            "    def login(self, username: str, password: str) -> bool:",
            '        """Authenticate user credentials."""',
            "        return self._check_credentials(username, password)",
            "",
            "    def logout(self, session_id: str) -> None:",
            '        """Invalidate session."""',
            "        self._sessions.pop(session_id, None)",
          ].join("\n"),
        },
      ],
    },
    {
      agent_id: "agent-beta",
      files: [
        {
          path: "src/auth.py",
          language: "python",
          content: [
            "class AuthService:",
            "    def login(self, email: str, pwd: str) -> dict:",
            '        """Login with email and return token."""',
            "        token = self._generate_token(email)",
            '        return {"token": token}',
            "",
            "    def reset_password(self, email: str) -> None:",
            '        """Send reset email."""',
            "        self._send_reset(email)",
          ].join("\n"),
        },
      ],
    },
  ],
}

// ---------------------------------------------------------------------------
// Severity styling
// ---------------------------------------------------------------------------

const SEVERITY_COLORS: Record<OverlapSeverity, string> = {
  critical: "bg-red-500/20 text-red-300 border-red-500/40",
  high: "bg-orange-500/20 text-orange-300 border-orange-500/40",
  medium: "bg-yellow-500/20 text-yellow-300 border-yellow-500/40",
  low: "bg-blue-500/20 text-blue-300 border-blue-500/40",
}

function SeverityBadge({ severity }: { severity: OverlapSeverity }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider",
        SEVERITY_COLORS[severity],
      )}
    >
      {severity}
    </span>
  )
}

// ---------------------------------------------------------------------------
// Health indicator
// ---------------------------------------------------------------------------

type HealthState = "checking" | "ok" | "error"

function HealthDot({ state }: { state: HealthState }) {
  return (
    <span
      className={cn("inline-block h-2 w-2 rounded-full", {
        "bg-yellow-400 animate-pulse": state === "checking",
        "bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.6)]": state === "ok",
        "bg-red-400 shadow-[0_0_6px_rgba(248,113,113,0.6)]": state === "error",
      })}
    />
  )
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function BackendPanel() {
  const [open, setOpen] = useState(false)
  const [tab, setTab] = useState<"health" | "analysis">("health")

  // Health state
  const [healthState, setHealthState] = useState<HealthState>("checking")
  const [healthData, setHealthData] = useState<HealthResponse | null>(null)
  const [healthError, setHealthError] = useState<string | null>(null)

  // Analysis state
  const [analysisLoading, setAnalysisLoading] = useState(false)
  const [analysisResult, setAnalysisResult] =
    useState<AnalyzeOverlapsResponse | null>(null)
  const [analysisError, setAnalysisError] = useState<string | null>(null)

  const checkHealth = useCallback(async () => {
    setHealthState("checking")
    setHealthError(null)
    try {
      const data = await fetchHealth()
      setHealthData(data)
      setHealthState("ok")
    } catch (err) {
      setHealthState("error")
      setHealthError(
        err instanceof ApiClientError ? err.message : "Connection failed",
      )
    }
  }, [])

  useEffect(() => {
    checkHealth()
    const interval = setInterval(checkHealth, 30_000)
    return () => clearInterval(interval)
  }, [checkHealth])

  async function runAnalysis() {
    setAnalysisLoading(true)
    setAnalysisError(null)
    setAnalysisResult(null)
    try {
      const result = await analyzeOverlaps(SAMPLE_PAYLOAD)
      setAnalysisResult(result)
    } catch (err) {
      setAnalysisError(
        err instanceof ApiClientError
          ? `${err.message}${err.details ? ` — ${err.details}` : ""}`
          : "Analysis request failed",
      )
    } finally {
      setAnalysisLoading(false)
    }
  }

  return (
    <>
      {/* Floating toggle */}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={cn(
          "fixed bottom-5 left-5 z-50 flex items-center gap-2 rounded-full border px-3 py-2 text-xs font-mono uppercase tracking-wider backdrop-blur transition-all duration-200",
          open
            ? "border-blue-400/60 bg-blue-950/90 text-blue-200"
            : "border-blue-400/30 bg-blue-950/70 text-blue-300 hover:border-blue-400/50 hover:bg-blue-950/80",
        )}
      >
        <HealthDot state={healthState} />
        <span>{open ? "Close Panel" : "Backend"}</span>
      </button>

      {/* Slide-up panel */}
      <div
        className={cn(
          "fixed bottom-16 left-5 z-40 w-[420px] max-h-[70vh] overflow-hidden rounded-2xl border border-blue-400/30 bg-blue-950/90 shadow-2xl shadow-blue-950/60 backdrop-blur-md transition-all duration-300",
          open
            ? "translate-y-0 opacity-100"
            : "pointer-events-none translate-y-4 opacity-0",
        )}
      >
        {/* Tab bar */}
        <div className="flex border-b border-blue-400/20">
          {(["health", "analysis"] as const).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTab(t)}
              className={cn(
                "flex-1 px-4 py-3 text-xs font-mono uppercase tracking-widest transition-colors",
                tab === t
                  ? "border-b-2 border-blue-400 text-blue-100"
                  : "text-blue-400/60 hover:text-blue-300",
              )}
            >
              {t === "health" ? "Health" : "Overlap Analysis"}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="overflow-y-auto p-5 max-h-[calc(70vh-48px)]">
          {tab === "health" ? (
            <HealthTab
              state={healthState}
              data={healthData}
              error={healthError}
              onRetry={checkHealth}
            />
          ) : (
            <AnalysisTab
              loading={analysisLoading}
              result={analysisResult}
              error={analysisError}
              onRun={runAnalysis}
            />
          )}
        </div>
      </div>
    </>
  )
}

// ---------------------------------------------------------------------------
// Health tab
// ---------------------------------------------------------------------------

function HealthTab({
  state,
  data,
  error,
  onRetry,
}: {
  state: HealthState
  data: HealthResponse | null
  error: string | null
  onRetry: () => void
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-blue-100">
          Backend Connection
        </h3>
        <button
          type="button"
          onClick={onRetry}
          className="text-[10px] font-mono uppercase tracking-wider text-blue-400 hover:text-blue-300 transition-colors"
        >
          Retry
        </button>
      </div>

      <div className="rounded-xl border border-blue-400/20 bg-blue-950/40 p-4 space-y-3">
        <div className="flex items-center gap-3">
          <HealthDot state={state} />
          <span className="text-sm font-mono text-blue-200">
            {state === "checking" && "Checking…"}
            {state === "ok" && `Connected — status: ${data?.status ?? "ok"}`}
            {state === "error" && "Unreachable"}
          </span>
        </div>

        {error && (
          <p className="text-xs text-red-300/80 font-mono">{error}</p>
        )}

        <div className="text-[10px] text-blue-400/50 font-mono space-y-1">
          <p>Route: /api/health → BACKEND_BASE_URL/health</p>
          <p>Auto-refreshes every 30 s</p>
        </div>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Analysis tab
// ---------------------------------------------------------------------------

function AnalysisTab({
  loading,
  result,
  error,
  onRun,
}: {
  loading: boolean
  result: AnalyzeOverlapsResponse | null
  error: string | null
  onRun: () => void
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-blue-100">
          Overlap Analysis
        </h3>
        <button
          type="button"
          onClick={onRun}
          disabled={loading}
          className={cn(
            "rounded-full border px-3 py-1.5 text-[10px] font-mono uppercase tracking-wider transition-all",
            loading
              ? "border-blue-400/20 text-blue-400/40 cursor-not-allowed"
              : "border-blue-400/40 text-blue-300 hover:border-blue-300/60 hover:bg-blue-900/40",
          )}
        >
          {loading ? "Running…" : "Run Sample"}
        </button>
      </div>

      {/* Sample payload preview */}
      <details className="group">
        <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-wider text-blue-400/60 hover:text-blue-300 transition-colors">
          Sample Payload (2 agents × src/auth.py)
        </summary>
        <pre className="mt-2 max-h-40 overflow-auto rounded-lg border border-blue-400/10 bg-slate-950/60 p-3 text-[10px] text-blue-300/70 font-mono leading-relaxed">
          {JSON.stringify(SAMPLE_PAYLOAD, null, 2)}
        </pre>
      </details>

      {/* Loading */}
      {loading && (
        <div className="flex items-center gap-2 text-xs text-blue-300/70 font-mono">
          <span className="inline-block h-3 w-3 animate-spin rounded-full border-2 border-blue-400/30 border-t-blue-400" />
          Analyzing changesets…
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="rounded-xl border border-red-500/30 bg-red-950/30 p-3 text-xs text-red-300 font-mono">
          {error}
        </div>
      )}

      {/* Results */}
      {result && (
        <div className="space-y-4">
          {/* Overlaps */}
          <div>
            <h4 className="text-xs font-semibold text-blue-200 mb-2">
              Overlaps ({result.overlaps.length})
            </h4>
            {result.overlaps.length === 0 ? (
              <p className="text-xs text-blue-400/50 font-mono">
                No overlaps detected.
              </p>
            ) : (
              <div className="space-y-2">
                {result.overlaps.map((o, i) => (
                  <div
                    key={`overlap-${i}`}
                    className="rounded-lg border border-blue-400/15 bg-blue-950/30 p-3 space-y-1.5"
                  >
                    <div className="flex items-center justify-between">
                      <code className="text-xs text-blue-200 font-mono">
                        {o.file_path}
                      </code>
                      <SeverityBadge severity={o.severity} />
                    </div>
                    <p className="text-[11px] text-blue-300/80 font-mono">
                      {o.symbol_kind}{" "}
                      <span className="text-blue-200">{o.symbol_name}</span>
                      {" — "}
                      {o.agent_a} vs {o.agent_b}
                    </p>
                    <p className="text-[10px] text-blue-400/50 font-mono">
                      {o.reason}
                    </p>
                    <div className="flex gap-4 text-[10px] text-blue-400/40 font-mono">
                      <span>
                        {o.agent_a}: L{o.start_line_a}–{o.end_line_a}
                      </span>
                      <span>
                        {o.agent_b}: L{o.start_line_b}–{o.end_line_b}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* File Risks */}
          <div>
            <h4 className="text-xs font-semibold text-blue-200 mb-2">
              File Risks ({result.file_risks.length})
            </h4>
            {result.file_risks.length === 0 ? (
              <p className="text-xs text-blue-400/50 font-mono">
                No file risks computed.
              </p>
            ) : (
              <div className="space-y-2">
                {result.file_risks.map((fr) => (
                  <div
                    key={fr.file_path}
                    className="rounded-lg border border-blue-400/15 bg-blue-950/30 p-3 space-y-2"
                  >
                    <div className="flex items-center justify-between">
                      <code className="text-xs text-blue-200 font-mono">
                        {fr.file_path}
                      </code>
                      {fr.is_hotspot && (
                        <span className="inline-flex items-center rounded-full border border-red-500/40 bg-red-500/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-red-300">
                          Hotspot
                        </span>
                      )}
                    </div>
                    {/* Risk bar */}
                    <div className="flex items-center gap-3">
                      <span className="text-[10px] text-blue-400/60 font-mono w-10">
                        Risk
                      </span>
                      <div className="flex-1 h-1.5 rounded-full bg-blue-950/60 overflow-hidden">
                        <div
                          className={cn(
                            "h-full rounded-full transition-all duration-500",
                            fr.risk_score >= 70
                              ? "bg-red-400"
                              : fr.risk_score >= 40
                                ? "bg-orange-400"
                                : "bg-emerald-400",
                          )}
                          style={{ width: `${fr.risk_score}%` }}
                        />
                      </div>
                      <span className="text-[10px] text-blue-200 font-mono w-6 text-right">
                        {fr.risk_score}
                      </span>
                    </div>
                    <p className="text-[10px] text-blue-400/50 font-mono">
                      {fr.summary}
                    </p>
                    <div className="flex gap-4 text-[10px] text-blue-400/40 font-mono">
                      <span>{fr.overlap_count} overlaps</span>
                      <span>{fr.contributors_count} contributors</span>
                      {fr.max_severity && (
                        <span>max: {fr.max_severity}</span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
