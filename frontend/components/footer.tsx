"use client"

import { motion } from "framer-motion"

const ease = [0.22, 1, 0.36, 1] as const

export function Footer() {
  return (
    <motion.footer
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true, margin: "-40px" }}
      transition={{ duration: 0.6, ease }}
      className="w-full border-t border-blue-400/30 px-6 py-10 lg:px-12"
    >
      <div className="mx-auto flex max-w-5xl flex-col gap-6 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-col gap-2">
          <span className="text-xs font-mono uppercase tracking-[0.3em] text-blue-200/80">
            Aero-Orchestrate
          </span>
          <span className="text-sm font-semibold text-blue-50">
            Build faster, plan smarter, orchestrate better.
          </span>
        </div>
        <div className="flex flex-wrap items-center gap-6 text-xs font-mono uppercase tracking-[0.2em] text-blue-200/80">
          {"GitHub Repo, Discord Community, API Reference".split(", ").map((link) => (
            <a key={link} href="#" className="hover:text-blue-50 transition-colors">
              {link}
            </a>
          ))}
        </div>
      </div>
    </motion.footer>
  )
}
