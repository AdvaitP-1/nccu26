import { proxyMcpGet } from "@/lib/mcp";

/**
 * GET /api/mcp/agents
 *
 * Returns all registered agents from the MCP agent registry.
 */
export async function GET() {
  return proxyMcpGet("/dashboard/agents");
}
