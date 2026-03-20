import { proxyMcpGet } from "@/lib/mcp";

export async function GET() {
  return proxyMcpGet("/dashboard/vfs");
}
