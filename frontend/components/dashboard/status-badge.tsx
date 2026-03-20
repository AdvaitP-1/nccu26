"use client"

import { cn } from "@/lib/utils"
import type { TopologyNodeStatus } from "@/lib/types"

const STATUS_STYLES: Record<TopologyNodeStatus, string> = {
  online: "bg-emerald-500",
  active: "bg-[#ea580c]",
  idle: "bg-muted-foreground",
  offline: "bg-red-500",
  unknown: "bg-muted-foreground/50",
}

const STATUS_LABELS: Record<TopologyNodeStatus, string> = {
  online: "ONLINE",
  active: "ACTIVE",
  idle: "IDLE",
  offline: "OFFLINE",
  unknown: "UNKNOWN",
}

export function StatusDot({ status }: { status: TopologyNodeStatus }) {
  return (
    <span
      className={cn(
        "inline-block h-1.5 w-1.5 shrink-0",
        STATUS_STYLES[status],
        (status === "online" || status === "active") && "animate-pulse",
      )}
    />
  )
}

export function StatusBadge({ status }: { status: TopologyNodeStatus }) {
  return (
    <span className="inline-flex items-center gap-1.5 text-[10px] font-mono tracking-widest uppercase text-muted-foreground">
      <StatusDot status={status} />
      {STATUS_LABELS[status]}
    </span>
  )
}
