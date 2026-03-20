"use client"

import { useState, useEffect, useCallback } from "react"
import { Waves, Power, Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"
import { VfsPanel } from "@/components/dashboard/vfs-panel"
import { TerminalPanel } from "@/components/dashboard/terminal-panel"
import { TopologyPanel } from "@/components/dashboard/topology-panel"
import { ThemeToggle } from "@/components/theme-toggle"
import { fetchMcpServerStatus, startMcpServer, type McpServerStatus } from "@/lib/api"
import Link from "next/link"

type Tab = "vfs" | "topology"

const STATUS_COLORS: Record<McpServerStatus["status"], string> = {
  running: "bg-emerald-500",
  starting: "bg-amber-400 animate-pulse",
  stopped: "bg-zinc-400",
  error: "bg-red-500",
}

export function DashboardShell() {
  const [tab, setTab] = useState<Tab>("vfs")
  const [mcpState, setMcpState] = useState<McpServerStatus>({
    status: "stopped",
    pid: null,
    logs: [],
    reachable: false,
  })
  const [starting, setStarting] = useState(false)

  const pollStatus = useCallback(async () => {
    try {
      const s = await fetchMcpServerStatus()
      setMcpState(s)
    } catch {
      setMcpState((prev) => ({ ...prev, reachable: false }))
    }
  }, [])

  useEffect(() => {
    pollStatus()
    const id = setInterval(pollStatus, 5000)
    return () => clearInterval(id)
  }, [pollStatus])

  const handleStart = async () => {
    setStarting(true)
    try {
      const s = await startMcpServer()
      setMcpState(s)
      // Poll aggressively for a few seconds while Go compiles
      for (let i = 0; i < 6; i++) {
        await new Promise((r) => setTimeout(r, 3000))
        const updated = await fetchMcpServerStatus()
        setMcpState(updated)
        if (updated.reachable) break
      }
    } catch {
      setMcpState((prev) => ({ ...prev, status: "error" }))
    } finally {
      setStarting(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col">
      {/* Top bar */}
      <header className="border-b-2 border-foreground px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link href="/" className="flex items-center gap-3">
            <Waves size={16} strokeWidth={1.5} />
            <span className="text-xs font-mono tracking-[0.15em] uppercase font-bold">
              ORCA AI
            </span>
          </Link>
          <span className="text-[10px] font-mono tracking-[0.2em] uppercase text-muted-foreground">
            // PM DASHBOARD
          </span>
        </div>
        <div className="flex items-center gap-4">
          {/* MCP Server Control */}
          <div className="flex items-center gap-2 border border-foreground/20 rounded px-3 py-1.5">
            <div className={cn("w-2 h-2 rounded-full", STATUS_COLORS[mcpState.status])} />
            <span className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
              MCP: {mcpState.status}
            </span>
            {!mcpState.reachable && (
              <button
                type="button"
                onClick={handleStart}
                disabled={starting || mcpState.status === "starting"}
                className={cn(
                  "ml-1 flex items-center gap-1 px-2 py-0.5 text-[10px] font-mono tracking-widest uppercase",
                  "border border-foreground/30 rounded transition-colors",
                  starting || mcpState.status === "starting"
                    ? "opacity-50 cursor-not-allowed"
                    : "hover:bg-foreground hover:text-background",
                )}
              >
                {starting || mcpState.status === "starting" ? (
                  <Loader2 size={10} className="animate-spin" />
                ) : (
                  <Power size={10} />
                )}
                {starting || mcpState.status === "starting" ? "Starting..." : "Start"}
              </button>
            )}
          </div>
          <span className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
            REAL-TIME: SSE + POLL FALLBACK
          </span>
          <ThemeToggle />
          <Link
            href="/"
            className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground hover:text-foreground transition-colors"
          >
            Landing
          </Link>
        </div>
      </header>

      {/* Tab navigation */}
      <div className="border-b-2 border-foreground flex">
        {(
          [
            { id: "vfs" as Tab, label: "VFS / CONTROL" },
            { id: "topology" as Tab, label: "AGENTS / TOPOLOGY" },
          ] as const
        ).map((t) => (
          <button
            key={t.id}
            type="button"
            onClick={() => setTab(t.id)}
            className={cn(
              "px-6 py-3 text-xs font-mono tracking-widest uppercase transition-colors border-r-2 border-foreground",
              tab === t.id
                ? "bg-foreground text-background"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/50",
            )}
          >
            {t.label}
          </button>
        ))}
        <div className="flex-1" />
      </div>

      {/* Content */}
      <div className="flex-1">
        {tab === "vfs" ? <VfsControlTab /> : <TopologyTab />}
      </div>
    </div>
  )
}

function VfsControlTab() {
  return (
    <div className="flex flex-col lg:flex-row h-[calc(100vh-108px)]">
      <div className="flex-1 border-r-2 border-foreground overflow-auto">
        <VfsPanel />
      </div>
      <div className="w-full lg:w-[480px] overflow-auto">
        <TerminalPanel />
      </div>
    </div>
  )
}

function TopologyTab() {
  return (
    <div className="h-[calc(100vh-108px)]">
      <TopologyPanel />
    </div>
  )
}
