import { type NextRequest } from "next/server";
import { proxyPost } from "@/lib/backend";

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ nodeId: string }> },
) {
  const { nodeId } = await params;
  return proxyPost(`/tree/${encodeURIComponent(nodeId)}/status`, request);
}
