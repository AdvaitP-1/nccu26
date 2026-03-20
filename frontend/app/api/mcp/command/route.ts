import { type NextRequest } from "next/server";
import { proxyMcpPost } from "@/lib/mcp";

export async function POST(request: NextRequest) {
  return proxyMcpPost("/dashboard/command", request);
}
