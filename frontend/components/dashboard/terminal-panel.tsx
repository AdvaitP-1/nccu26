"use client"

import { useState, useRef, useEffect, useCallback } from "react"
import { sendMcpCommand, ApiClientError } from "@/lib/api"
import type { CommandHistoryEntry } from "@/lib/types"

const COMMANDS: Record<
  string,
  { description: string; parse: (parts: string[]) => { command: string; args: Record<string, unknown> } | null }
> = {
  "get-vfs-state": {
    description: "Show current VFS state",
    parse: () => ({ command: "get_vfs_state", args: {} }),
  },
  vfs: {
    description: "Alias for get-vfs-state",
    parse: () => ({ command: "get_vfs_state", args: {} }),
  },
  "identify-overlaps": {
    description: "Run overlap analysis on pending VFS changes",
    parse: () => ({ command: "identify_overlaps", args: {} }),
  },
  overlaps: {
    description: "Alias for identify-overlaps",
    parse: () => ({ command: "identify_overlaps", args: {} }),
  },
  "request-micro-commit": {
    description: "Request micro-commit: request-micro-commit <agent_id> [message]",
    parse: (parts) => {
      const agentId = parts[1]
      if (!agentId) return null
      const message = parts.slice(2).join(" ") || undefined
      return { command: "request_micro_commit", args: { agent_id: agentId, message } }
    },
  },
  commit: {
    description: "Alias: commit <agent_id> [message]",
    parse: (parts) => {
      const agentId = parts[1]
      if (!agentId) return null
      const message = parts.slice(2).join(" ") || undefined
      return { command: "request_micro_commit", args: { agent_id: agentId, message } }
    },
  },
  "clear-agent-state": {
    description: "Clear agent VFS: clear-agent-state <agent_id>",
    parse: (parts) => {
      const agentId = parts[1]
      if (!agentId) return null
      return { command: "clear_agent_state", args: { agent_id: agentId } }
    },
  },
  clear: {
    description: "Alias: clear <agent_id>",
    parse: (parts) => {
      const agentId = parts[1]
      if (!agentId) return null
      return { command: "clear_agent_state", args: { agent_id: agentId } }
    },
  },
  "git-health": {
    description: "Check git subsystem health",
    parse: () => ({ command: "git_health", args: {} }),
  },
  health: {
    description: "Alias for git-health",
    parse: () => ({ command: "git_health", args: {} }),
  },
  "seed-demo": {
    description: "Populate VFS with 5 realistic demo agents",
    parse: () => ({ command: "seed_demo", args: {} }),
  },
  seed: {
    description: "Alias for seed-demo",
    parse: () => ({ command: "seed_demo", args: {} }),
  },
  "clear-demo": {
    description: "Remove all demo data (VFS + agent registry)",
    parse: () => ({ command: "clear_demo", args: {} }),
  },
  "register-agent": {
    description: "Register agent: register-agent <id> <type> [display_name]",
    parse: (parts) => {
      const id = parts[1]
      if (!id) return null
      const agentType = parts[2] || "coder"
      const displayName = parts.slice(3).join(" ") || id
      return { command: "register_agent", args: { id, type: agentType, display_name: displayName } }
    },
  },
  "list-agents": {
    description: "List all registered agents with types",
    parse: () => ({ command: "list_agents", args: {} }),
  },
  agents: {
    description: "Alias for list-agents",
    parse: () => ({ command: "list_agents", args: {} }),
  },
}

function formatOutput(data: unknown): string {
  if (typeof data === "string") return data
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

let entryCounter = 0

export function TerminalPanel() {
  const [input, setInput] = useState("")
  const [history, setHistory] = useState<CommandHistoryEntry[]>([])
  const [running, setRunning] = useState(false)
  const [cmdHistory, setCmdHistory] = useState<string[]>([])
  const [histIdx, setHistIdx] = useState(-1)
  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    scrollRef.current?.scrollTo(0, scrollRef.current.scrollHeight)
  }, [history])

  const addEntry = useCallback(
    (input: string, output: string, success: boolean) => {
      entryCounter++
      setHistory((h) => [
        ...h,
        {
          id: `cmd-${entryCounter}`,
          input,
          output,
          success,
          timestamp: Date.now(),
        },
      ])
    },
    [],
  )

  const handleSubmit = useCallback(async () => {
    const raw = input.trim()
    if (!raw) return
    setInput("")
    setCmdHistory((h) => [...h, raw])
    setHistIdx(-1)

    if (raw === "help") {
      const lines = ["Available commands:", ""]
      const seen = new Set<string>()
      for (const [name, info] of Object.entries(COMMANDS)) {
        if (seen.has(info.description)) continue
        seen.add(info.description)
        lines.push(`  ${name.padEnd(26)} ${info.description}`)
      }
      lines.push("", "  help                       Show this help")
      lines.push("  clear-screen               Clear terminal output")
      addEntry(raw, lines.join("\n"), true)
      return
    }

    if (raw === "clear-screen") {
      setHistory([])
      return
    }

    const parts = raw.split(/\s+/)
    const cmdName = parts[0]
    const cmdDef = COMMANDS[cmdName]

    if (!cmdDef) {
      addEntry(raw, `Unknown command: "${cmdName}". Type 'help' for available commands.`, false)
      return
    }

    const parsed = cmdDef.parse(parts)
    if (!parsed) {
      addEntry(raw, `Invalid arguments. Usage: ${cmdDef.description}`, false)
      return
    }

    setRunning(true)
    try {
      const result = await sendMcpCommand(parsed)
      if (result.error) {
        addEntry(raw, `ERROR: ${result.error}`, false)
      } else {
        addEntry(raw, formatOutput(result.data), result.success)
      }
    } catch (err) {
      addEntry(
        raw,
        err instanceof ApiClientError
          ? `ERROR [${err.status}]: ${err.message}`
          : "Request failed",
        false,
      )
    } finally {
      setRunning(false)
    }
  }, [input, addEntry])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        e.preventDefault()
        handleSubmit()
      } else if (e.key === "ArrowUp") {
        e.preventDefault()
        if (cmdHistory.length === 0) return
        const next = histIdx < 0 ? cmdHistory.length - 1 : Math.max(0, histIdx - 1)
        setHistIdx(next)
        setInput(cmdHistory[next])
      } else if (e.key === "ArrowDown") {
        e.preventDefault()
        if (histIdx < 0) return
        const next = histIdx + 1
        if (next >= cmdHistory.length) {
          setHistIdx(-1)
          setInput("")
        } else {
          setHistIdx(next)
          setInput(cmdHistory[next])
        }
      }
    },
    [handleSubmit, cmdHistory, histIdx],
  )

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-3 border-b-2 border-foreground">
        <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
          COMMAND CONSOLE
        </span>
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 bg-[#ea580c]" />
          <span className="h-2 w-2 bg-foreground" />
          <span className="h-2 w-2 border border-foreground" />
        </div>
      </div>

      {/* Output area */}
      <div
        ref={scrollRef}
        onClick={() => inputRef.current?.focus()}
        className="flex-1 bg-foreground overflow-auto p-4 cursor-text"
      >
        {/* Welcome message */}
        {history.length === 0 && (
          <div className="text-xs font-mono text-background/60 space-y-1">
            <p>ORCA AI Command Console v2.0.0</p>
            <p>Type &apos;help&apos; for available commands.</p>
            <p>Type &apos;seed&apos; to populate demo data.</p>
            <p className="text-background/30">
              ─────────────────────────────────────
            </p>
          </div>
        )}

        {/* History entries */}
        {history.map((entry) => (
          <div key={entry.id} className="mb-3">
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono text-[#ea580c]">{">"}</span>
              <span className="text-xs font-mono text-background">
                {entry.input}
              </span>
            </div>
            <pre className={`text-xs font-mono mt-1 whitespace-pre-wrap break-all ${
              entry.success ? "text-background/70" : "text-red-400"
            }`}>
              {entry.output}
            </pre>
          </div>
        ))}

        {running && (
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-[#ea580c] animate-blink">
              {"_"}
            </span>
            <span className="text-xs font-mono text-background/50">
              Running...
            </span>
          </div>
        )}
      </div>

      {/* Input */}
      <div className="flex items-center border-t-2 border-foreground bg-foreground px-4 py-3">
        <span className="text-xs font-mono text-[#ea580c] mr-2">{">"}</span>
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={running}
          placeholder={running ? "Running..." : "Enter command..."}
          className="flex-1 bg-transparent text-xs font-mono text-background placeholder:text-background/30 outline-none"
          autoFocus
        />
      </div>
    </div>
  )
}
