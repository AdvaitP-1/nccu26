import { proxyGet } from "@/lib/backend";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ nodeId: string }> },
) {
  const { nodeId } = await params;
  return proxyGet(`/tree/${encodeURIComponent(nodeId)}/siblings`);
}
