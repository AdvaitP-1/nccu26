"use client"

import { useState } from "react"
import { motion } from "framer-motion"
import { Navbar } from "@/components/navbar"
import { Footer } from "@/components/footer"
import { cn } from "@/lib/utils"
import Link from "next/link"

const ease = [0.22, 1, 0.36, 1] as const

type DocSection =
  | "getting-started"
  | "architecture"
  | "mcp-tools"
  | "dashboard-api"
  | "configuration"

const SECTIONS: { id: DocSection; label: string }[] = [
  { id: "getting-started", label: "Getting Started" },
  { id: "architecture", label: "Architecture" },
  { id: "mcp-tools", label: "MCP Tools" },
  { id: "dashboard-api", label: "Dashboard API" },
  { id: "configuration", label: "Configuration" },
]

function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-foreground text-background text-[11px] font-mono p-4 overflow-x-auto leading-relaxed border-2 border-foreground">
      {children}
    </pre>
  )
}

function SectionHeading({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="text-sm font-mono font-bold tracking-wide uppercase mt-8 mb-3 flex items-center gap-2">
      <span className="h-1.5 w-1.5 bg-[#ea580c]" />
      {children}
    </h3>
  )
}

function Paragraph({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-xs font-mono text-muted-foreground leading-relaxed mb-4">
      {children}
    </p>
  )
}

function GettingStarted() {
  return (
    <div>
      <h2 className="text-xl font-mono font-bold tracking-tight uppercase mb-6">
        Getting Started
      </h2>

      <Paragraph>
        Orca AI is a monorepo with three services: a Next.js frontend, a Python
        analysis backend, and a Go MCP server. Follow these steps to get
        everything running locally.
      </Paragraph>

      <SectionHeading>Prerequisites</SectionHeading>
      <ul className="space-y-2 mb-6">
        {[
          "Node.js 18+ and npm",
          "Python 3.11+ and pip",
          "Go 1.22+",
          "Git",
        ].map((item) => (
          <li key={item} className="flex items-start gap-2">
            <span className="h-1.5 w-1.5 bg-foreground/40 mt-1.5 shrink-0" />
            <span className="text-xs font-mono text-muted-foreground">{item}</span>
          </li>
        ))}
      </ul>

      <SectionHeading>1. Start the Python backend</SectionHeading>
      <CodeBlock>{`cd backend
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000`}</CodeBlock>

      <SectionHeading>2. Start the frontend</SectionHeading>
      <CodeBlock>{`cd frontend
cp .env.example .env
npm install
npm run dev`}</CodeBlock>

      <SectionHeading>3. Start the MCP server</SectionHeading>
      <CodeBlock>{`cd mcp
go run ./cmd/server`}</CodeBlock>

      <Paragraph>
        The MCP server starts an SSE transport on :3000 for MCP clients and an
        HTTP API on :9091 for dashboard endpoints. The frontend connects to the
        HTTP API via the MCP_BASE_URL environment variable.
      </Paragraph>

      <SectionHeading>4. Seed demo data</SectionHeading>
      <Paragraph>
        Once all services are running, open the dashboard at{" "}
        <Link href="/dashboard" className="text-[#ea580c] hover:underline">
          /dashboard
        </Link>{" "}
        and click &quot;Seed Demo Data&quot; or type <code className="text-foreground bg-muted px-1">seed</code> in the
        command console. This populates the VFS with 5 realistic agents.
      </Paragraph>
    </div>
  )
}

function Architecture() {
  return (
    <div>
      <h2 className="text-xl font-mono font-bold tracking-tight uppercase mb-6">
        Architecture
      </h2>

      <SectionHeading>Data Flow</SectionHeading>
      <CodeBlock>{`Browser → Frontend UI → Next.js Route Handlers (/api/*) → Python Backend (:8000)
                                                            → MCP Go Server (:9091)

IBM watsonx Orchestrate → MCP Go Server (:3000/SSE) → Python Backend (:8000)`}</CodeBlock>

      <SectionHeading>Layer Boundaries</SectionHeading>
      <div className="border-2 border-foreground mb-6">
        <div className="grid grid-cols-4 border-b-2 border-foreground bg-muted/30">
          {["Layer", "Runs", "Talks To", "Env Var"].map((h) => (
            <div key={h} className="px-3 py-2 text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
              {h}
            </div>
          ))}
        </div>
        {[
          ["Frontend UI", "Browser", "Next.js /api/*", "—"],
          ["Route Handlers", "Server-side", "Backend, MCP", "BACKEND_BASE_URL, MCP_BASE_URL"],
          ["Python Backend", "Server :8000", "—", "—"],
          ["MCP Go Server", "Server :3000 + :9091", "Backend", "MCP_BACKEND_URL"],
          ["watsonx Agents", "IBM SaaS", "MCP (via ngrok)", "MCP_BASE_URL"],
        ].map((row) => (
          <div key={row[0]} className="grid grid-cols-4 border-b border-border last:border-none">
            {row.map((cell, i) => (
              <div key={`${row[0]}-${i}`} className="px-3 py-2 text-[11px] font-mono text-muted-foreground">
                {cell}
              </div>
            ))}
          </div>
        ))}
      </div>

      <SectionHeading>Key Rules</SectionHeading>
      <ul className="space-y-2 mb-6">
        {[
          "The frontend never calls the backend or MCP directly from browser code.",
          "BACKEND_BASE_URL and MCP_BASE_URL are server-only env vars (no NEXT_PUBLIC_ prefix).",
          "ngrok is only used to expose MCP to IBM watsonx during demos.",
          "The agent registry is the source of truth for agent types (not name heuristics).",
          "SSE events are best-effort; the dashboard falls back to polling if SSE disconnects.",
        ].map((rule) => (
          <li key={rule} className="flex items-start gap-2">
            <span className="h-1.5 w-1.5 bg-[#ea580c] mt-1.5 shrink-0" />
            <span className="text-xs font-mono text-muted-foreground">{rule}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function McpTools() {
  const tools = [
    {
      name: "get_vfs_state",
      description: "Returns the full VFS snapshot: all pending changes, file counts, and agent counts.",
      args: "None",
    },
    {
      name: "identify_overlaps",
      description: "Sends all VFS changesets to the analysis backend for structural overlap detection and risk scoring.",
      args: "None",
    },
    {
      name: "request_micro_commit",
      description: "Requests a micro-commit for a specific agent. Runs overlap analysis, evaluates policy, and commits if allowed.",
      args: "agent_id (string), message (string, optional)",
    },
    {
      name: "register_push",
      description: "Ingests a git push event with file changes for a branch/user. Used by external systems to feed data into Orca.",
      args: "branch_name, user_id, files_json, message",
    },
    {
      name: "git_health",
      description: "Returns the health status of the git subsystem.",
      args: "None",
    },
    {
      name: "get_branch_file_state",
      description: "Returns the file tree state for a specific branch in the storage layer.",
      args: "branch_name (string)",
    },
    {
      name: "prepare_merge_context",
      description: "Gathers the context needed to merge two branches, including file diffs and conflict markers.",
      args: "source_branch, target_branch",
    },
    {
      name: "apply_merge_result",
      description: "Applies a resolved merge result to the storage layer.",
      args: "merge_id, resolutions",
    },
    {
      name: "prepare_commit",
      description: "Stages files for a commit in the git execution layer.",
      args: "branch_name, files",
    },
    {
      name: "create_commit",
      description: "Creates a git commit with the staged changes.",
      args: "branch_name, message, author",
    },
    {
      name: "push_commit",
      description: "Pushes committed changes to the remote repository.",
      args: "branch_name, remote",
    },
    {
      name: "get_commit_status",
      description: "Returns the status of a previously initiated commit operation.",
      args: "commit_id (string)",
    },
  ]

  return (
    <div>
      <h2 className="text-xl font-mono font-bold tracking-tight uppercase mb-6">
        MCP Tools Reference
      </h2>

      <Paragraph>
        Orca AI exposes {tools.length} tools via the MCP protocol over SSE. These
        tools can be invoked by any MCP-compatible client, including IBM watsonx
        Orchestrate, Cursor IDE, and custom ADK integrations.
      </Paragraph>

      <SectionHeading>Tool Inventory</SectionHeading>
      <div className="space-y-0 mb-6">
        {tools.map((tool, i) => (
          <div key={tool.name} className={cn("border-2 border-foreground p-4", i > 0 && "-mt-[2px]")}>
            <div className="flex items-center gap-2 mb-2">
              <code className="text-xs font-mono font-bold text-[#ea580c]">
                {tool.name}
              </code>
            </div>
            <p className="text-[11px] font-mono text-muted-foreground mb-2">
              {tool.description}
            </p>
            <div className="flex items-center gap-2">
              <span className="text-[10px] font-mono tracking-widest uppercase text-muted-foreground/60">
                ARGS:
              </span>
              <span className="text-[11px] font-mono text-foreground/70">
                {tool.args}
              </span>
            </div>
          </div>
        ))}
      </div>

      <SectionHeading>Invocation via SSE</SectionHeading>
      <Paragraph>
        Connect to the MCP SSE endpoint at <code className="text-foreground bg-muted px-1">http://localhost:3000/sse</code>.
        The server advertises all tools during the handshake. Send JSON-RPC
        tool-call messages to invoke any tool.
      </Paragraph>
      <CodeBlock>{`// Example: calling get_vfs_state via MCP protocol
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "get_vfs_state",
    "arguments": {}
  },
  "id": 1
}`}</CodeBlock>
    </div>
  )
}

function DashboardApi() {
  const endpoints = [
    { method: "GET", path: "/dashboard/vfs", description: "Returns VFS state snapshot (all pending changes)" },
    { method: "GET", path: "/dashboard/agents", description: "Returns all registered agents from the agent registry" },
    { method: "GET", path: "/dashboard/events", description: "SSE stream of real-time events (vfs_update, agent_registered, agent_removed, command_result)" },
    { method: "POST", path: "/dashboard/command", description: "Executes a structured command and returns the result" },
  ]

  const commands = [
    { name: "get_vfs_state", args: "—", description: "Get VFS snapshot" },
    { name: "identify_overlaps", args: "—", description: "Run overlap analysis" },
    { name: "request_micro_commit", args: "agent_id, message?", description: "Commit an agent's changes" },
    { name: "clear_agent_state", args: "agent_id", description: "Clear an agent's VFS entries" },
    { name: "git_health", args: "—", description: "Check git subsystem health" },
    { name: "register_push", args: "branch_name, user_id, files_json, message?", description: "Ingest a push event" },
    { name: "seed_demo", args: "—", description: "Populate VFS with 5 demo agents" },
    { name: "clear_demo", args: "—", description: "Wipe all demo data" },
    { name: "register_agent", args: "id, type?, display_name?", description: "Register an agent with explicit type" },
    { name: "list_agents", args: "—", description: "List all registered agents" },
  ]

  return (
    <div>
      <h2 className="text-xl font-mono font-bold tracking-tight uppercase mb-6">
        Dashboard API Reference
      </h2>

      <Paragraph>
        The MCP HTTP API (default port 9091) exposes REST endpoints used by the
        Next.js frontend route handlers. The frontend proxies these through
        /api/mcp/* to keep MCP_BASE_URL server-side only.
      </Paragraph>

      <SectionHeading>REST Endpoints</SectionHeading>
      <div className="border-2 border-foreground mb-6">
        {endpoints.map((ep, i) => (
          <div key={ep.path} className={cn("flex items-start gap-4 px-4 py-3", i > 0 && "border-t border-border")}>
            <span className={cn(
              "text-[10px] font-mono font-bold tracking-widest shrink-0 px-2 py-0.5",
              ep.method === "GET" ? "bg-green-500/20 text-green-400" : "bg-blue-500/20 text-blue-400",
            )}>
              {ep.method}
            </span>
            <div>
              <code className="text-xs font-mono font-bold text-foreground">{ep.path}</code>
              <p className="text-[11px] font-mono text-muted-foreground mt-1">{ep.description}</p>
            </div>
          </div>
        ))}
      </div>

      <SectionHeading>Command Reference</SectionHeading>
      <Paragraph>
        Send commands via <code className="text-foreground bg-muted px-1">POST /dashboard/command</code> with
        a JSON body: <code className="text-foreground bg-muted px-1">{`{"command": "...", "args": {...}}`}</code>
      </Paragraph>
      <div className="border-2 border-foreground mb-6">
        <div className="grid grid-cols-3 border-b-2 border-foreground bg-muted/30">
          {["Command", "Args", "Description"].map((h) => (
            <div key={h} className="px-3 py-2 text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
              {h}
            </div>
          ))}
        </div>
        {commands.map((cmd) => (
          <div key={cmd.name} className="grid grid-cols-3 border-b border-border last:border-none">
            <div className="px-3 py-2">
              <code className="text-[11px] font-mono text-[#ea580c]">{cmd.name}</code>
            </div>
            <div className="px-3 py-2 text-[11px] font-mono text-muted-foreground">{cmd.args}</div>
            <div className="px-3 py-2 text-[11px] font-mono text-muted-foreground">{cmd.description}</div>
          </div>
        ))}
      </div>

      <SectionHeading>SSE Event Types</SectionHeading>
      <CodeBlock>{`event: vfs_update
data: {"type":"vfs_update","timestamp":"...","data":{...VFSState}}

event: agent_registered
data: {"type":"agent_registered","timestamp":"...","data":{"id":"...","type":"coder"}}

event: agent_removed
data: {"type":"agent_removed","timestamp":"...","data":{"scope":"all"}}

event: command_result
data: {"type":"command_result","timestamp":"...","data":{"command":"..."}}`}</CodeBlock>
    </div>
  )
}

function Configuration() {
  return (
    <div>
      <h2 className="text-xl font-mono font-bold tracking-tight uppercase mb-6">
        Configuration
      </h2>

      <SectionHeading>Frontend Environment Variables</SectionHeading>
      <div className="border-2 border-foreground mb-6">
        {[
          { name: "BACKEND_BASE_URL", default: "http://localhost:8000", description: "Python analysis backend URL (server-side only)" },
          { name: "MCP_BASE_URL", default: "http://localhost:9091", description: "MCP HTTP API URL (server-side only)" },
        ].map((v, i) => (
          <div key={v.name} className={cn("px-4 py-3", i > 0 && "border-t border-border")}>
            <code className="text-xs font-mono font-bold text-[#ea580c]">{v.name}</code>
            <p className="text-[11px] font-mono text-muted-foreground mt-1">{v.description}</p>
            <p className="text-[11px] font-mono text-foreground/50 mt-1">
              Default: <code className="bg-muted px-1">{v.default}</code>
            </p>
          </div>
        ))}
      </div>

      <SectionHeading>MCP Server Configuration</SectionHeading>
      <Paragraph>
        The MCP server is configured via environment variables or CLI flags.
        Key settings:
      </Paragraph>
      <div className="border-2 border-foreground mb-6">
        {[
          { name: "SERVER_ADDR", default: ":3000", description: "SSE transport listen address" },
          { name: "HTTP_ADDR", default: ":9091", description: "HTTP API listen address (dashboard + git endpoints)" },
          { name: "PUBLIC_URL", default: "http://localhost:3000", description: "Public URL for SSE base (used by MCP clients)" },
          { name: "BACKEND_BASE_URL", default: "http://localhost:8000", description: "Analysis backend URL" },
          { name: "GIT_REPO_PATH", default: "", description: "Path to git repo for execution mode (empty = state-tracking only)" },
          { name: "RISK_THRESHOLD", default: "0.7", description: "Risk score threshold for policy evaluation" },
          { name: "BLOCK_ON_CRITICAL", default: "true", description: "Block micro-commits when critical overlaps are detected" },
        ].map((v, i) => (
          <div key={v.name} className={cn("px-4 py-3", i > 0 && "border-t border-border")}>
            <code className="text-xs font-mono font-bold text-[#ea580c]">{v.name}</code>
            <p className="text-[11px] font-mono text-muted-foreground mt-1">{v.description}</p>
            <p className="text-[11px] font-mono text-foreground/50 mt-1">
              Default: <code className="bg-muted px-1">{v.default}</code>
            </p>
          </div>
        ))}
      </div>

      <SectionHeading>Agent Types</SectionHeading>
      <Paragraph>
        Agents are classified by type in the registry. Each type receives
        distinct styling in the topology graph and can be targeted by policy rules.
      </Paragraph>
      <div className="border-2 border-foreground mb-6">
        <div className="grid grid-cols-3 border-b-2 border-foreground bg-muted/30">
          {["Type", "Color", "Description"].map((h) => (
            <div key={h} className="px-3 py-2 text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
              {h}
            </div>
          ))}
        </div>
        {[
          { type: "manager", color: "bg-purple-500", description: "Architecture planners and task decomposers" },
          { type: "coder", color: "bg-blue-500", description: "Code-writing agents assigned to specific tasks" },
          { type: "merge", color: "bg-amber-500", description: "Merge coordinators and conflict resolvers" },
          { type: "reviewer", color: "bg-green-500", description: "Code review and quality assurance agents" },
          { type: "external", color: "bg-gray-500", description: "External orchestrators (e.g. watsonx)" },
        ].map((row) => (
          <div key={row.type} className="grid grid-cols-3 border-b border-border last:border-none items-center">
            <div className="px-3 py-2">
              <code className="text-[11px] font-mono font-bold text-foreground">{row.type}</code>
            </div>
            <div className="px-3 py-2">
              <span className={cn("inline-block h-3 w-3 rounded-full", row.color)} />
            </div>
            <div className="px-3 py-2 text-[11px] font-mono text-muted-foreground">{row.description}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

const SECTION_CONTENT: Record<DocSection, () => React.JSX.Element> = {
  "getting-started": GettingStarted,
  architecture: Architecture,
  "mcp-tools": McpTools,
  "dashboard-api": DashboardApi,
  configuration: Configuration,
}

export default function DocsPage() {
  const [activeSection, setActiveSection] = useState<DocSection>("getting-started")
  const Content = SECTION_CONTENT[activeSection]

  return (
    <div className="min-h-screen dot-grid-bg">
      <Navbar />
      <main>
        <section className="w-full px-6 pt-12 pb-6 lg:px-12 lg:pt-16 lg:pb-8">
          <div className="max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, ease }}
              className="flex items-center gap-3 mb-6"
            >
              <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
                {"// DOCUMENTATION"}
              </span>
              <div className="w-16 border-t border-border" />
              <span className="h-2 w-2 bg-[#ea580c]" />
            </motion.div>

            <motion.h1
              initial={{ opacity: 0, y: 20, filter: "blur(6px)" }}
              animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
              transition={{ duration: 0.7, ease }}
              className="font-pixel text-3xl sm:text-4xl lg:text-5xl tracking-tight text-foreground mb-4"
            >
              DOCS
            </motion.h1>

            <motion.p
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2, ease }}
              className="text-xs lg:text-sm text-muted-foreground max-w-2xl leading-relaxed font-mono mb-8"
            >
              Everything you need to set up, configure, and integrate with Orca
              AI&apos;s multi-agent orchestration platform.
            </motion.p>
          </div>
        </section>

        <section className="w-full px-6 pb-20 lg:px-12 lg:pb-28">
          <div className="max-w-5xl mx-auto flex flex-col lg:flex-row gap-0">
            {/* Sidebar nav */}
            <motion.aside
              initial={{ opacity: 0, x: -16 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.3, ease }}
              className="w-full lg:w-[220px] shrink-0 border-2 border-foreground lg:border-r-0 mb-0"
            >
              <div className="px-4 py-3 border-b-2 border-foreground bg-muted/30">
                <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
                  SECTIONS
                </span>
              </div>
              {SECTIONS.map((section) => (
                <button
                  key={section.id}
                  type="button"
                  onClick={() => setActiveSection(section.id)}
                  className={cn(
                    "w-full text-left px-4 py-3 text-xs font-mono tracking-wide uppercase border-b border-border last:border-none transition-colors",
                    activeSection === section.id
                      ? "bg-foreground text-background"
                      : "text-muted-foreground hover:text-foreground hover:bg-muted/50",
                  )}
                >
                  {section.label}
                </button>
              ))}
            </motion.aside>

            {/* Content area */}
            <motion.div
              key={activeSection}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, ease }}
              className="flex-1 border-2 border-foreground p-6 lg:p-8"
            >
              <Content />
            </motion.div>
          </div>
        </section>
      </main>
      <Footer />
    </div>
  )
}
