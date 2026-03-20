"use client"

import { motion } from "framer-motion"

const ease = [0.22, 1, 0.36, 1] as const

export function HeroSection() {
  return (
    <section className="relative w-full px-6 pb-20 pt-20 lg:px-12">
      <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
        <motion.p
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease }}
          className="mb-4 text-xs font-mono uppercase tracking-[0.3em] text-blue-200/80"
        >
          Aero-Orchestrate
        </motion.p>
        <motion.h1
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, ease }}
          className="text-3xl font-semibold text-blue-50 sm:text-5xl lg:text-6xl"
        >
          Orchestrate Your AI Engineering Team.
        </motion.h1>
        <motion.p
          initial={{ opacity: 0, y: 14 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.1, ease }}
          className="mt-6 max-w-2xl text-sm leading-relaxed text-blue-200/90 sm:text-base"
        >
          Move beyond basic completion. Deploy a coordinated swarm of specialized AI agents that architect, code, test, and review—all governed by a single source of truth.
        </motion.p>
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2, ease }}
          className="mt-8 flex flex-wrap items-center justify-center gap-3"
        >
          <button className="rounded-full bg-blue-400 px-6 py-3 text-xs font-mono uppercase tracking-[0.2em] text-blue-950 transition hover:bg-blue-300">
            View Documentation
          </button>
          <button className="rounded-full border border-blue-300/50 px-6 py-3 text-xs font-mono uppercase tracking-[0.2em] text-blue-100 transition hover:border-blue-200 hover:text-blue-50">
            Explore the Multi-Agent Framework
          </button>
        </motion.div>
      </div>
    </section>
  )
}
