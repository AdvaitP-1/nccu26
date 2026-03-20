"use client"

import { motion } from "framer-motion"

const ease = [0.22, 1, 0.36, 1] as const

export function AboutSection() {
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
          The Why
          <span className="h-px flex-1 bg-blue-400/40" />
        </div>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div className="rounded-2xl border border-blue-400/30 bg-blue-950/70 p-6">
            <h3 className="text-sm font-semibold text-blue-50">The Problem</h3>
            <p className="mt-3 text-sm leading-relaxed text-blue-200/80">
              Single-prompt AI tools lose context, hallucinate on large codebases, and can't manage complex workflows autonomously.
            </p>
          </div>
          <div className="rounded-2xl border border-blue-400/30 bg-blue-950/60 p-6">
            <h3 className="text-sm font-semibold text-blue-50">The Solution: An Orchestration Layer</h3>
            <p className="mt-3 text-sm leading-relaxed text-blue-200/80">
              By breaking tasks down into a "Spec-First" workflow, our system ensures that every agent (Planner, Coder, Reviewer) stays aligned with the project's architectural requirements.
            </p>
          </div>
        </div>
      </motion.div>
    </section>
  )
}
