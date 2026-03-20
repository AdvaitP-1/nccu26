"use client";

import { useEffect, useRef, useState, useCallback } from "react";

export interface SSEEvent {
  type: string;
  timestamp: string;
  data: unknown;
}

interface UseSSEOptions {
  url: string;
  onEvent: (event: SSEEvent) => void;
  enabled?: boolean;
  reconnectMs?: number;
}

/**
 * React hook that connects to an SSE endpoint and dispatches parsed
 * events to the caller.  Auto-reconnects on disconnect with exponential
 * backoff capped at `reconnectMs`.  Falls back gracefully if the SSE
 * endpoint is unavailable.
 */
export function useSSE({
  url,
  onEvent,
  enabled = true,
  reconnectMs = 3000,
}: UseSSEOptions) {
  const [connected, setConnected] = useState(false);
  const [retries, setRetries] = useState(0);
  const sourceRef = useRef<EventSource | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  const connect = useCallback(() => {
    if (sourceRef.current) {
      sourceRef.current.close();
      sourceRef.current = null;
    }

    const source = new EventSource(url);
    sourceRef.current = source;

    source.onopen = () => {
      setConnected(true);
      setRetries(0);
    };

    source.onmessage = (e) => {
      try {
        const parsed: SSEEvent = JSON.parse(e.data);
        onEventRef.current(parsed);
      } catch {
        // non-JSON keepalive or malformed — ignore
      }
    };

    // Named event types from MCP
    const eventTypes = [
      "vfs_update",
      "agent_registered",
      "agent_removed",
      "command_result",
    ];
    for (const t of eventTypes) {
      source.addEventListener(t, (e: MessageEvent) => {
        try {
          const parsed: SSEEvent = JSON.parse(e.data);
          onEventRef.current(parsed);
        } catch {
          // ignore
        }
      });
    }

    source.onerror = () => {
      setConnected(false);
      source.close();
      sourceRef.current = null;
      setRetries((r) => r + 1);
      const delay = Math.min(reconnectMs * Math.pow(2, retries), 30000);
      reconnectRef.current = setTimeout(connect, delay);
    };
  }, [url, reconnectMs, retries]);

  useEffect(() => {
    if (!enabled) {
      if (sourceRef.current) {
        sourceRef.current.close();
        sourceRef.current = null;
      }
      setConnected(false);
      return;
    }

    connect();

    return () => {
      if (sourceRef.current) {
        sourceRef.current.close();
        sourceRef.current = null;
      }
      if (reconnectRef.current) clearTimeout(reconnectRef.current);
    };
  }, [connect, enabled]);

  return { connected };
}
