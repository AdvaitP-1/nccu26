"use client"

import { WorkflowDiagram } from "@/components/workflow-diagram"
import { motion } from "framer-motion"
import Link from "next/link"

const ease = [0.22, 1, 0.36, 1] as const

export function HeroSection() {
  return (
    <section className="relative w-full px-12 pt-6 pb-12 lg:px-24 lg:pt-10 lg:pb-16">
      <div className="flex flex-col items-center text-center">
        <motion.h1
          initial={{ opacity: 0, y: 30, filter: "blur(8px)" }}
          animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
          transition={{ duration: 0.7, ease }}
          className="font-pixel text-4xl sm:text-6xl lg:text-7xl xl:text-8xl tracking-tight text-foreground mb-2 select-none"
        >
          ORCHESTRATE. ANALYZE.
        </motion.h1>

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.6, delay: 0.15, ease }}
          className="w-full max-w-2xl my-4 lg:my-6"
        >
          <WorkflowDiagram />
        </motion.div>

        <motion.h1
          initial={{ opacity: 0, y: 30, filter: "blur(8px)" }}
          animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
          transition={{ duration: 0.7, delay: 0.25, ease }}
          className="font-pixel text-4xl sm:text-6xl lg:text-7xl xl:text-8xl tracking-tight text-foreground mb-4 select-none"
          aria-hidden="true"
        >
          MERGE.
        </motion.h1>

        <motion.p
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.45, ease }}
          className="text-xs lg:text-sm text-muted-foreground max-w-md mb-6 leading-relaxed font-mono"
        >
          Orca AI is the multi-agent orchestration layer that prevents code conflicts before they happen. Real-time overlap detection. Structural risk analysis. Full merge control.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.6, ease }}
        >
          <Link
            href="/dashboard"
            className="bg-foreground text-background px-6 py-2.5 text-xs font-mono tracking-widest uppercase hover:bg-[#ea580c] transition-colors"
          >
            Open Dashboard
          </Link>
        </motion.div>
      </div>
    </section>
  )
}
