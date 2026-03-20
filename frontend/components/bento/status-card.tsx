"use client"

import { useEffect, useState } from "react"

const AGENTS = [
  { name: "ARCH-PLANNER", status: "ACTIVE", tasks: "12/12" },
  { name: "CODER-ALPHA", status: "ACTIVE", tasks: "8/10" },
  { name: "CODER-BETA", status: "ACTIVE", tasks: "6/10" },
  { name: "REVIEWER-01", status: "IDLE", tasks: "0/0" },
]

export function StatusCard() {
  const [tick, setTick] = useState(0)

  useEffect(() => {
    const interval = setInterval(() => {
      setTick((t) => t + 1)
    }, 2000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between border-b-2 border-foreground px-4 py-2">
        <span className="text-[10px] tracking-widest text-muted-foreground uppercase">
          agent_pool.status
        </span>
        <span className="text-[10px] tracking-widest text-muted-foreground">
          {`TICK:${String(tick).padStart(4, "0")}`}
        </span>
      </div>
      <div className="flex-1 flex flex-col p-4 gap-0">
        <div className="grid grid-cols-3 gap-2 border-b border-border pb-2 mb-2">
          <span className="text-[9px] tracking-[0.15em] uppercase text-muted-foreground">Agent</span>
          <span className="text-[9px] tracking-[0.15em] uppercase text-muted-foreground">Status</span>
          <span className="text-[9px] tracking-[0.15em] uppercase text-muted-foreground text-right">Tasks</span>
        </div>
        {AGENTS.map((agent) => (
          <div
            key={agent.name}
            className="grid grid-cols-3 gap-2 py-2 border-b border-border last:border-none"
          >
            <span className="text-xs font-mono text-foreground">{agent.name}</span>
            <div className="flex items-center gap-2">
              <span
                className="h-1.5 w-1.5"
                style={{
                  backgroundColor: agent.status === "ACTIVE" ? "#ea580c" : "hsl(var(--muted-foreground))",
                }}
              />
              <span className="text-xs font-mono text-muted-foreground">{agent.status}</span>
            </div>
            <span className="text-xs font-mono text-foreground text-right">{agent.tasks}</span>
          </div>
        ))}
        <div className="mt-auto pt-4">
          <div className="flex items-center justify-between mb-1">
            <span className="text-[9px] tracking-[0.15em] uppercase text-muted-foreground">
              Merge Success Rate
            </span>
            <span className="text-[9px] font-mono text-foreground">99%</span>
          </div>
          <div className="h-2 w-full border border-foreground">
            <div className="h-full bg-foreground" style={{ width: "99%" }} />
          </div>
        </div>
      </div>
    </div>
  )
}
