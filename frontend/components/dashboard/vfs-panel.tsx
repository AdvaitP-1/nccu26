"use client"

import { useCallback, useEffect, useState } from "react"
import { fetchVfsState, seedDemo, clearDemo, ApiClientError } from "@/lib/api"
import { useSSE } from "@/hooks/use-sse"
import type { SSEEvent } from "@/hooks/use-sse"
import type { VfsState, VfsPendingChange } from "@/lib/types"
import { StatusDot } from "@/components/dashboard/status-badge"
import { cn } from "@/lib/utils"

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const secs = Math.floor(diff / 1000)
  if (secs < 60) return `${secs}s ago`
  const mins = Math.floor(secs / 60)
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  return `${hrs}h ago`
}

export function VfsPanel() {
  const [vfs, setVfs] = useState<VfsState | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [seeding, setSeeding] = useState(false)

  const load = useCallback(async () => {
    try {
      const state = await fetchVfsState()
      setVfs(state)
      setError(null)
    } catch (err) {
      setError(
        err instanceof ApiClientError ? err.message : "Failed to fetch VFS",
      )
    } finally {
      setLoading(false)
    }
  }, [])

  // SSE: push updates from MCP arrive here — replaces most polling
  const handleSSE = useCallback((evt: SSEEvent) => {
    if (evt.type === "vfs_update" && evt.data) {
      setVfs(evt.data as VfsState)
      setError(null)
    }
  }, [])

  const { connected } = useSSE({
    url: "/api/mcp/events",
    onEvent: handleSSE,
  })

  // Fallback polling when SSE is disconnected
  useEffect(() => {
    load()
    if (!connected) {
      const id = setInterval(load, 3000)
      return () => clearInterval(id)
    }
  }, [load, connected])

  const handleSeed = useCallback(async () => {
    setSeeding(true)
    try {
      await seedDemo()
      await load()
    } catch {
      // terminal panel shows errors for command failures
    } finally {
      setSeeding(false)
    }
  }, [load])

  const handleClear = useCallback(async () => {
    try {
      await clearDemo()
      await load()
    } catch {
      // ignore
    }
  }, [load])

  const isEmpty = !error && vfs && vfs.pending_changes.length === 0

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-3 border-b-2 border-foreground">
        <div className="flex items-center gap-3">
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            VIRTUAL FILE SYSTEM
          </span>
          <StatusDot status={error ? "offline" : "online"} />
        </div>
        <div className="flex items-center gap-3">
          {vfs && vfs.pending_changes.length > 0 && (
            <button
              type="button"
              onClick={handleClear}
              className="text-[10px] font-mono tracking-widest uppercase text-red-400 hover:text-red-300 transition-colors"
            >
              Clear All
            </button>
          )}
          <button
            type="button"
            onClick={load}
            className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground hover:text-foreground transition-colors"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Summary bar */}
      <div className="flex items-center gap-6 px-5 py-2 border-b border-border bg-muted/30">
        <span className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
          AGENTS: {vfs?.total_agents ?? "—"}
        </span>
        <span className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
          FILES: {vfs?.total_files ?? "—"}
        </span>
        <span
          className={cn(
            "text-[10px] font-mono tracking-widest uppercase",
            connected ? "text-green-500" : "text-[#ea580c]",
          )}
        >
          {loading ? "LOADING…" : connected ? "SSE LIVE" : "POLLING"}
        </span>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-5">
        {error && (
          <div className="border-2 border-red-500/30 p-4 mb-4">
            <p className="text-xs font-mono text-red-400">{error}</p>
            <p className="text-[10px] font-mono text-muted-foreground mt-1">
              Ensure MCP server is running and MCP_BASE_URL is set.
            </p>
          </div>
        )}

        {/* Empty state with seed prompt */}
        {isEmpty && (
          <div className="border-2 border-border p-8 text-center">
            <p className="text-xs font-mono text-muted-foreground">
              No pending changes in VFS
            </p>
            <p className="text-[10px] font-mono text-muted-foreground/60 mt-2 mb-6">
              Agents have not pushed any file changes yet.
            </p>
            <button
              type="button"
              onClick={handleSeed}
              disabled={seeding}
              className={cn(
                "px-6 py-3 border-2 border-foreground text-xs font-mono tracking-widest uppercase",
                "hover:bg-foreground hover:text-background transition-colors",
                seeding && "opacity-50 cursor-not-allowed",
              )}
            >
              {seeding ? "SEEDING…" : "SEED DEMO DATA"}
            </button>
            <p className="text-[10px] font-mono text-muted-foreground/40 mt-3">
              Populates 5 agents with realistic multi-file changesets
            </p>
          </div>
        )}

        {vfs &&
          vfs.pending_changes.map((pc) => (
            <AgentCard key={pc.agent_id} change={pc} />
          ))}
      </div>
    </div>
  )
}

function AgentCard({ change }: { change: VfsPendingChange }) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="border-2 border-foreground mb-4">
      {/* Agent header */}
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-muted/30 transition-colors"
      >
        <div className="flex items-center gap-3">
          <StatusDot status="active" />
          <span className="text-xs font-mono font-bold tracking-wide uppercase">
            {change.agent_id}
          </span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[10px] font-mono text-muted-foreground">
            {change.files.length} file{change.files.length !== 1 ? "s" : ""}
          </span>
          <span className="text-[10px] font-mono text-muted-foreground">
            {timeAgo(change.updated_at)}
          </span>
          <span className={cn("text-xs font-mono transition-transform", expanded && "rotate-90")}>
            {">"}
          </span>
        </div>
      </button>

      {/* Metadata */}
      <div className="px-4 py-2 border-t border-border flex gap-6">
        {change.session_id && (
          <span className="text-[10px] font-mono text-muted-foreground">
            SESSION: {change.session_id}
          </span>
        )}
        {change.task_id && (
          <span className="text-[10px] font-mono text-muted-foreground">
            TASK: {change.task_id}
          </span>
        )}
      </div>

      {/* File list */}
      {expanded && (
        <div className="border-t-2 border-foreground">
          {change.files.map((f, i) => (
            <div
              key={`${f.path}-${i}`}
              className="flex items-center justify-between px-4 py-2 border-b border-border last:border-none"
            >
              <span className="text-xs font-mono text-foreground truncate">
                {f.path}
              </span>
              <span className="text-[10px] font-mono text-muted-foreground shrink-0 ml-4">
                {f.language || "—"}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
