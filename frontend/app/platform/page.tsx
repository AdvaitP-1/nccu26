"use client"

import { motion } from "framer-motion"
import { Navbar } from "@/components/navbar"
import { Footer } from "@/components/footer"
import {
  Layers,
  GitBranch,
  Shield,
  Activity,
  Network,
  Terminal,
  Waves,
} from "lucide-react"
import Link from "next/link"

const ease = [0.22, 1, 0.36, 1] as const

const CAPABILITIES = [
  {
    icon: Layers,
    title: "Virtual File System",
    description:
      "Every pending file change from every agent is tracked in an in-memory VFS before anything touches your repo. The VFS is the single source of truth for what agents intend to write, enabling pre-commit analysis and conflict detection across all active workstreams.",
    details: [
      "Per-agent shadow workspaces",
      "File-level change tracking with timestamps",
      "Session and task ID correlation",
      "Atomic propose / commit / clear lifecycle",
    ],
  },
  {
    icon: Shield,
    title: "Overlap Detection & Risk Analysis",
    description:
      "When two or more agents touch the same file or function, Orca's analysis backend detects the structural overlap and scores the risk. The policy engine then decides whether to allow, warn, or block a micro-commit based on configurable thresholds.",
    details: [
      "AST-aware structural comparison",
      "Configurable risk thresholds per project",
      "Block-on-critical policy enforcement",
      "Overlap heatmaps across the full file tree",
    ],
  },
  {
    icon: GitBranch,
    title: "Micro-Commit Coordination",
    description:
      "Agents don't push directly to your branch. Instead, Orca coordinates micro-commits: small, verified changesets that pass overlap analysis and policy evaluation before being applied. This eliminates merge conflicts at the source.",
    details: [
      "Pre-commit policy gate with allow/warn/block",
      "Automatic VFS cleanup after successful commit",
      "Agent-scoped commit messages",
      "Full audit trail of every commit decision",
    ],
  },
  {
    icon: Network,
    title: "Task Tree Decomposition",
    description:
      "Complex features are broken into a tree of isolated tasks, each assigned to a specific agent. The tree structure ensures agents work on non-overlapping units while maintaining a clear dependency graph that Orca uses to order merges.",
    details: [
      "Hierarchical task breakdown",
      "Agent-to-task assignment tracking",
      "Dependency-aware merge ordering",
      "Sibling and parent node queries",
    ],
  },
  {
    icon: Activity,
    title: "Real-Time Dashboard",
    description:
      "The PM dashboard gives product managers live visibility into the orchestration layer. See VFS state, run structured commands, and visualize the agent topology—all through a controlled interface backed by SSE with polling fallback.",
    details: [
      "SSE-powered live updates with auto-reconnect",
      "Structured command console (not a raw shell)",
      "Agent topology graph with type classification",
      "One-click demo seeding for stakeholder demos",
    ],
  },
  {
    icon: Terminal,
    title: "MCP Tool Protocol",
    description:
      "Orca exposes its capabilities as MCP tools over SSE, allowing any MCP-compatible client—including IBM watsonx Orchestrate and Cursor—to invoke VFS operations, overlap analysis, and merge coordination programmatically.",
    details: [
      "12+ registered MCP tools",
      "SSE transport for remote invocation",
      "Structured JSON tool schemas",
      "Compatible with watsonx ADK and Cursor",
    ],
  },
]

const ARCHITECTURE_LAYERS = [
  {
    label: "EXTERNAL ORCHESTRATORS",
    items: ["IBM watsonx Orchestrate", "Cursor IDE", "Custom ADK Clients"],
    color: "border-muted-foreground/40",
  },
  {
    label: "MCP SERVER",
    items: ["SSE Transport", "Tool Registry", "Dashboard API", "Event Bus"],
    color: "border-foreground",
  },
  {
    label: "CORE ENGINE",
    items: ["VFS Manager", "Agent Registry", "Policy Evaluator", "Commit Coordinator"],
    color: "border-[#ea580c]",
  },
  {
    label: "ANALYSIS BACKEND",
    items: ["Overlap Detection", "Risk Scoring", "AST Parsing", "Merge Planning"],
    color: "border-foreground",
  },
]

function CapabilityCard({
  capability,
  index,
}: {
  capability: (typeof CAPABILITIES)[number]
  index: number
}) {
  const Icon = capability.icon

  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-40px" }}
      transition={{ delay: index * 0.08, duration: 0.6, ease }}
      className="border-2 border-foreground flex flex-col"
    >
      <div className="flex items-center gap-3 px-5 py-4 border-b-2 border-foreground">
        <Icon size={16} strokeWidth={1.5} className="text-[#ea580c] shrink-0" />
        <h3 className="text-sm font-mono font-bold tracking-wide uppercase">
          {capability.title}
        </h3>
      </div>
      <div className="px-5 py-4 flex-1">
        <p className="text-xs font-mono text-muted-foreground leading-relaxed mb-4">
          {capability.description}
        </p>
        <ul className="space-y-2">
          {capability.details.map((detail) => (
            <li key={detail} className="flex items-start gap-2">
              <span className="h-1.5 w-1.5 bg-[#ea580c] mt-1.5 shrink-0" />
              <span className="text-[11px] font-mono text-foreground/80">
                {detail}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </motion.div>
  )
}

export default function PlatformPage() {
  return (
    <div className="min-h-screen dot-grid-bg">
      <Navbar />
      <main>
        {/* Hero */}
        <section className="w-full px-6 pt-12 pb-16 lg:px-12 lg:pt-16 lg:pb-24">
          <div className="max-w-4xl mx-auto text-center">
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, ease }}
              className="flex items-center justify-center gap-3 mb-6"
            >
              <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
                {"// PLATFORM"}
              </span>
              <div className="w-16 border-t border-border" />
              <span className="h-2 w-2 bg-[#ea580c]" />
            </motion.div>

            <motion.h1
              initial={{ opacity: 0, y: 20, filter: "blur(6px)" }}
              animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
              transition={{ duration: 0.7, ease }}
              className="font-pixel text-3xl sm:text-5xl lg:text-6xl tracking-tight text-foreground mb-6"
            >
              THE MULTI-AGENT
              <br />
              <span className="text-[#ea580c]">ORCHESTRATION LAYER</span>
            </motion.h1>

            <motion.p
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2, ease }}
              className="text-xs lg:text-sm text-muted-foreground max-w-2xl mx-auto leading-relaxed font-mono"
            >
              Orca AI sits between your AI agents and your repository. It tracks
              every pending change in a virtual file system, detects structural
              overlaps in real-time, and coordinates micro-commits so conflicts
              never reach your main branch.
            </motion.p>
          </div>
        </section>

        {/* Architecture diagram */}
        <section className="w-full px-6 pb-16 lg:px-12 lg:pb-24">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true, margin: "-60px" }}
            transition={{ duration: 0.5, ease }}
            className="flex items-center gap-4 mb-8"
          >
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              {"// ARCHITECTURE"}
            </span>
            <div className="flex-1 border-t border-border" />
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              001
            </span>
          </motion.div>

          <div className="max-w-3xl mx-auto space-y-0">
            {ARCHITECTURE_LAYERS.map((layer, i) => (
              <motion.div
                key={layer.label}
                initial={{ opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1, duration: 0.5, ease }}
                className={`border-2 ${layer.color} ${i > 0 ? "-mt-[2px]" : ""}`}
              >
                <div className="flex items-center gap-3 px-5 py-2 border-b border-border">
                  <span className="h-1.5 w-1.5 bg-[#ea580c]" />
                  <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
                    {layer.label}
                  </span>
                </div>
                <div className="flex flex-wrap gap-3 px-5 py-3">
                  {layer.items.map((item) => (
                    <span
                      key={item}
                      className="text-[11px] font-mono text-foreground/80 border border-border px-3 py-1"
                    >
                      {item}
                    </span>
                  ))}
                </div>
              </motion.div>
            ))}
            <motion.div
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ delay: 0.5, duration: 0.4 }}
              className="flex items-center justify-center gap-2 py-4"
            >
              <Waves size={12} className="text-muted-foreground" />
              <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
                DATA FLOWS TOP → BOTTOM
              </span>
            </motion.div>
          </div>
        </section>

        {/* Capabilities grid */}
        <section className="w-full px-6 pb-20 lg:px-12 lg:pb-28">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true, margin: "-60px" }}
            transition={{ duration: 0.5, ease }}
            className="flex items-center gap-4 mb-8"
          >
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              {"// CAPABILITIES"}
            </span>
            <div className="flex-1 border-t border-border" />
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              002
            </span>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-0">
            {CAPABILITIES.map((cap, i) => (
              <CapabilityCard key={cap.title} capability={cap} index={i} />
            ))}
          </div>
        </section>

        {/* CTA */}
        <section className="w-full px-6 pb-20 lg:px-12 lg:pb-28">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, ease }}
            className="border-2 border-foreground p-8 lg:p-12 text-center"
          >
            <h2 className="text-xl lg:text-2xl font-mono font-bold tracking-tight uppercase mb-4">
              See it in action
            </h2>
            <p className="text-xs font-mono text-muted-foreground max-w-lg mx-auto mb-8 leading-relaxed">
              The PM Dashboard provides live visibility into the orchestration
              layer. Explore the VFS, run commands, and visualize the agent
              topology.
            </p>
            <Link
              href="/dashboard"
              className="inline-block bg-foreground text-background px-8 py-3 text-xs font-mono tracking-widest uppercase hover:bg-[#ea580c] transition-colors"
            >
              Open Dashboard
            </Link>
          </motion.div>
        </section>
      </main>
      <Footer />
    </div>
  )
}
