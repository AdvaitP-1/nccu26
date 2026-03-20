/**
 * Server-only backend proxy utilities.
 *
 * These run exclusively inside Next.js Route Handlers (server-side).
 * The BACKEND_BASE_URL env var is never exposed to the browser.
 */

import { NextResponse } from "next/server";

const DEFAULT_TIMEOUT_MS = 15_000;

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

class BackendError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: string,
  ) {
    super(message);
    this.name = "BackendError";
  }
}

// ---------------------------------------------------------------------------
// Core fetch wrapper (server-only)
// ---------------------------------------------------------------------------

function getBackendUrl(): string {
  const url = process.env.BACKEND_BASE_URL;
  if (!url) {
    throw new BackendError(
      "BACKEND_BASE_URL is not configured",
      503,
      "Set the BACKEND_BASE_URL environment variable to the backend origin (e.g. http://localhost:8000).",
    );
  }
  return url.replace(/\/+$/, "");
}

async function backendFetch(
  path: string,
  init: RequestInit = {},
  timeoutMs = DEFAULT_TIMEOUT_MS,
): Promise<Response> {
  const base = getBackendUrl();
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
      throw new BackendError("Backend request timed out", 504);
    }
    throw new BackendError(
      "Backend unavailable",
      502,
      err instanceof Error ? err.message : String(err),
    );
  } finally {
    clearTimeout(timer);
  }
}

// ---------------------------------------------------------------------------
// JSON error response builder
// ---------------------------------------------------------------------------

function jsonError(
  message: string,
  status: number,
  details?: string,
): NextResponse {
  const body: { error: string; details?: string } = { error: message };
  if (details) body.details = details;
  return NextResponse.json(body, { status });
}

// ---------------------------------------------------------------------------
// Public helpers used by route handlers
// ---------------------------------------------------------------------------

/**
 * Proxy a GET request to the backend and return the response as-is.
 */
export async function proxyGet(backendPath: string): Promise<NextResponse> {
  try {
    const res = await backendFetch(backendPath);
    const data = await safeJson(res);
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return handleProxyError(err);
  }
}

/**
 * Proxy a POST request to the backend.
 * Reads the body from the incoming Next.js Request.
 */
export async function proxyPost(
  backendPath: string,
  request: Request,
): Promise<NextResponse> {
  try {
    const body = await request.text();
    const res = await backendFetch(backendPath, {
      method: "POST",
      body,
    });
    const data = await safeJson(res);
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return handleProxyError(err);
  }
}

/**
 * Proxy a POST request with a raw JSON string body.
 */
export async function proxyPostRaw(
  backendPath: string,
  body: string,
): Promise<NextResponse> {
  try {
    const res = await backendFetch(backendPath, {
      method: "POST",
      body,
    });
    const data = await safeJson(res);
    return NextResponse.json(data, { status: res.status });
  } catch (err) {
    return handleProxyError(err);
  }
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

async function safeJson(res: Response): Promise<unknown> {
  const text = await res.text();
  try {
    return JSON.parse(text);
  } catch {
    throw new BackendError(
      "Backend returned non-JSON response",
      502,
      text.slice(0, 200),
    );
  }
}

function handleProxyError(err: unknown): NextResponse {
  if (err instanceof BackendError) {
    return jsonError(err.message, err.status, err.details);
  }
  return jsonError("Internal server error", 500);
}
