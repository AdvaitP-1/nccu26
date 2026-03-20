"use client"

import { motion } from "framer-motion"

const ease = [0.22, 1, 0.36, 1] as const

const TECH_FEATURES = [
  {
    title: "Multi-Model Agnostic",
    body: "Swap between Claude, GPT-4, or local Llama models depending on the task's complexity.",
  },
  {
    title: "Event-Driven Triggers",
    body: "Automatically trigger a \"Security Review\" agent whenever the Coder agent modifies sensitive files.",
  },
  {
    title: "Human-in-the-Loop",
    body: "A dedicated UI terminal and Kanban board so you can intervene, approve PRs, or pivot the strategy at any time.",
  },
  {
    title: "Integrated MCP Support",
    body: "Uses the Model Context Protocol to give agents deep access to your local files and tools securely.",
  },
]

export function PricingSection() {
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
          Key Technical Features
          <span className="h-px flex-1 bg-blue-400/40" />
        </div>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {TECH_FEATURES.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 18 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-60px" }}
              transition={{ delay: index * 0.08, duration: 0.5, ease }}
              className="rounded-2xl border border-blue-400/30 bg-blue-950/60 p-6 shadow-lg shadow-blue-950/40"
            >
              <h3 className="text-sm font-semibold text-blue-50">{feature.title}</h3>
              <p className="mt-3 text-sm leading-relaxed text-blue-200/80">{feature.body}</p>
            </motion.div>
          ))}
        </div>
      </motion.div>
    </section>
  )
}
