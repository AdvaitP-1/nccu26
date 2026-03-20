/**
 * Server-only MCP proxy utilities.
 *
 * These run exclusively inside Next.js Route Handlers (server-side).
 * The MCP_BASE_URL env var is never exposed to the browser.
 */

import { NextResponse } from "next/server";

const DEFAULT_TIMEOUT_MS = 15_000;

class McpError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: string,
  ) {
    super(message);
    this.name = "McpError";
  }
}

function getMcpUrl(): string {
  const url = process.env.MCP_BASE_URL;
  if (!url) {
    throw new McpError(
      "MCP_BASE_URL is not configured",
      503,
      "Set the MCP_BASE_URL environment variable to the MCP HTTP API origin (e.g. http://localhost:9091).",
    );
  }
  return url.replace(/\/+$/, "");
}

async function mcpFetch(
  path: string,
  init: RequestInit = {},
  timeoutMs = DEFAULT_TIMEOUT_MS,
): Promise<Response> {
  const base = getMcpUrl();
  const url = `${base}${path}`;

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const res = await fetch(url, {
      ...init,
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
        ...init.headers,
      },
    });
    return res;
  } catch (err: unknown) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new McpError("MCP request timed out", 504);
    }
    throw new McpError(
      "MCP unavailable",
      502,
      err instanceof Error ? err.message : String(err),
    );
  } finally {
    clearTimeout(timer);
  }
}

function jsonError(
  message: string,
  status: number,
  details?: string,
): NextResponse {
  const body: { error: string; details?: string } = { error: message };
  if (details) body.details = details;
  return NextResponse.json(body, { status });
}

async function safeJson(res: Response): Promise<unknown> {
  const text = await res.text();
  try {
    return JSON.parse(text);
  } catch {
    throw new McpError(
      "MCP returned non-JSON response",
      502,
      text.slice(0, 200),
    );
  }
}

function handleMcpError(err: unknown): NextResponse {
  if (err instanceof McpError) {
    return jsonError(err.message, err.status, err.details);
  }
  return jsonError("Internal server error", 500);
}

/**
 * Proxy a GET request to the MCP HTTP API.
 */
export async function proxyMcpGet(mcpPath: string): Promise<NextResponse> {
  try {
    const res = await mcpFetch(mcpPath);
    const data = await safeJson(res);
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return handleMcpError(err);
  }
}

/**
 * Proxy a POST request to the MCP HTTP API, forwarding the incoming body.
 */
export async function proxyMcpPost(
  mcpPath: string,
  request: Request,
): Promise<NextResponse> {
  try {
    const body = await request.text();
    const res = await mcpFetch(mcpPath, { method: "POST", body });
    const data = await safeJson(res);
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return handleMcpError(err);
  }
}

/**
 * Proxy a POST with a raw JSON body string.
 */
export async function proxyMcpPostRaw(
  mcpPath: string,
  body: string,
): Promise<NextResponse> {
  try {
    const res = await mcpFetch(mcpPath, { method: "POST", body });
    const data = await safeJson(res);
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return handleMcpError(err);
  }
}
