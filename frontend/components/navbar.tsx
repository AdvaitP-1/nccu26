"use client"

import { Shell } from "lucide-react"
import { motion } from "framer-motion"
import { ThemeToggle } from "@/components/theme-toggle"

export function Navbar() {
  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
      className="w-full px-4 lg:px-6"
    >
      <nav className="w-full border border-blue-400/30 bg-blue-950/70 backdrop-blur-sm px-6 lg:px-8 h-[72px]">
        <div className="flex h-full items-center justify-between text-blue-50">
          {/* Logo */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2, duration: 0.4 }}
            className="flex items-center gap-3"
          >
            <Shell size={16} strokeWidth={1.5} className="text-blue-200" />
            <span className="text-xs font-mono tracking-[0.15em] uppercase font-bold">
              ORCA
            </span>
          </motion.div>

          {/* Center nav links */}
          <div className="hidden md:flex items-center gap-8">
            {["Platform", "Enterprise", "Resources", "Company"].map((link, i) => (
              <motion.a
                key={link}
                href="#"
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 + i * 0.06, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
                className="text-xs font-mono tracking-widest uppercase text-blue-200/80 hover:text-blue-50 transition-colors duration-200"
              >
                {link}
              </motion.a>
            ))}
          </div>

          {/* Right side: Login + CTA */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5, duration: 0.4 }}
            className="flex items-center gap-4"
          >
            <ThemeToggle />
            <a
              href="#"
              className="hidden sm:block text-xs font-mono tracking-widest uppercase text-blue-200/80 hover:text-blue-50 transition-colors duration-200"
            >
              Log In
            </a>
            <motion.button
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className="bg-blue-400 text-blue-950 px-4 py-2 text-xs font-mono tracking-widest uppercase hover:bg-blue-300 transition-colors"
            >
              Request Demo
            </motion.button>
          </motion.div>
        </div>
      </nav>
    </motion.div>
  )
}
