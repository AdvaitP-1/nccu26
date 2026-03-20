/**
 * GET /api/mcp/events
 *
 * Proxies the MCP SSE stream to the browser.  The browser connects to
 * this endpoint via EventSource, and we relay events from the upstream
 * MCP server's /dashboard/events endpoint.  This keeps MCP_BASE_URL
 * server-side only.
 */
export const dynamic = "force-dynamic";

export async function GET() {
  const mcpUrl = process.env.MCP_BASE_URL?.replace(/\/+$/, "");
  if (!mcpUrl) {
    return new Response("MCP_BASE_URL not configured", { status: 503 });
  }

  try {
    const upstream = await fetch(`${mcpUrl}/dashboard/events`, {
      headers: { Accept: "text/event-stream" },
      cache: "no-store",
    });

    if (!upstream.ok || !upstream.body) {
      return new Response("MCP SSE connection failed", { status: 502 });
    }

    return new Response(upstream.body, {
      headers: {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache, no-transform",
        Connection: "keep-alive",
      },
    });
  } catch {
    return new Response("Failed to connect to MCP", { status: 502 });
  }
}
