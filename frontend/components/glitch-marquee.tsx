"use client"

import { motion } from "framer-motion"

const ease = [0.22, 1, 0.36, 1] as const

const STEPS = [
  {
    title: "Define (The Spec)",
    body: "Start with a spec.md. The Orchestrator analyzes your requirements and builds a task roadmap.",
  },
  {
    title: "Execute (The Swarm)",
    body: "Specialized agents pick up tasks from an autonomous Kanban board. The Coder writes, the Tester validates, and the Reviewer audits.",
  },
  {
    title: "Refine (Context Compaction)",
    body: "As the project grows, the system automatically \"compacts\" context, keeping the AI's memory sharp and costs low.",
  },
]

export function GlitchMarquee() {
  return (
    <section className="w-full px-6 py-16 lg:px-12">
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 0.6, ease }}
        className="mx-auto flex max-w-5xl flex-col gap-8"
      >
        <div className="flex items-center gap-4 text-xs font-mono uppercase tracking-[0.3em] text-blue-200/80">
          Core Workflow
          <span className="h-px flex-1 bg-blue-400/40" />
        </div>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {STEPS.map((step, index) => (
            <motion.div
              key={step.title}
              initial={{ opacity: 0, y: 18 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-60px" }}
              transition={{ delay: index * 0.1, duration: 0.5, ease }}
              className="rounded-2xl border border-blue-400/30 bg-blue-950/60 p-6"
            >
              <h3 className="text-sm font-semibold text-blue-50">{step.title}</h3>
              <p className="mt-3 text-sm leading-relaxed text-blue-200/80">{step.body}</p>
            </motion.div>
          ))}
        </div>
      </motion.div>
    </section>
  )
}
