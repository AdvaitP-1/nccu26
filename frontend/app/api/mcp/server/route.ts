import { NextResponse } from "next/server";
import { spawn } from "child_process";
import path from "path";

let mcpProcess: ReturnType<typeof spawn> | null = null;
let mcpStatus: "stopped" | "starting" | "running" | "error" = "stopped";
let mcpLogs: string[] = [];
const MAX_LOGS = 50;

function addLog(line: string) {
  mcpLogs.push(line);
  if (mcpLogs.length > MAX_LOGS) mcpLogs.shift();
}

async function probeMcp(): Promise<boolean> {
  const url = process.env.MCP_BASE_URL?.replace(/\/+$/, "");
  if (!url) return false;
  try {
    const res = await fetch(`${url}/git/health`, {
      signal: AbortSignal.timeout(2000),
    });
    return res.ok;
  } catch {
    return false;
  }
}

/**
 * GET /api/mcp/server
 *
 * Returns the current MCP server status and recent logs.
 */
export async function GET() {
  const live = await probeMcp();
  if (live && mcpStatus !== "running") mcpStatus = "running";
  if (!live && mcpStatus === "running" && !mcpProcess) mcpStatus = "stopped";

  return NextResponse.json({
    status: mcpStatus,
    pid: mcpProcess?.pid ?? null,
    logs: mcpLogs.slice(-20),
    reachable: live,
  });
}

/**
 * POST /api/mcp/server
 *
 * Starts the MCP server as a child process.  If already running, returns
 * the current status without spawning again.
 */
export async function POST() {
  const alreadyRunning = await probeMcp();
  if (alreadyRunning) {
    mcpStatus = "running";
    return NextResponse.json({
      status: "running",
      message: "MCP server is already running",
      reachable: true,
    });
  }

  if (mcpProcess && mcpStatus === "starting") {
    return NextResponse.json({
      status: "starting",
      message: "MCP server is currently starting up",
    });
  }

  const repoRoot = path.resolve(process.cwd(), "..");
  const mcpDir = path.join(repoRoot, "mcp");

  mcpStatus = "starting";
  mcpLogs = [];
  addLog(`[dashboard] Starting MCP server from ${mcpDir}`);

  try {
    mcpProcess = spawn("go", ["run", "./cmd/server"], {
      cwd: mcpDir,
      env: {
        ...process.env,
        MCP_GIT_REPO_PATH: repoRoot,
        MCP_SERVER_ADDR: ":9090",
        MCP_HTTP_ADDR: ":9091",
        MCP_BACKEND_URL: "http://localhost:8000",
      },
      stdio: ["ignore", "pipe", "pipe"],
      detached: false,
    });

    mcpProcess.stdout?.on("data", (data: Buffer) => {
      const line = data.toString().trim();
      if (line) addLog(line);
      if (line.includes("starting MCP SSE server") || line.includes("starting HTTP API server")) {
        mcpStatus = "running";
      }
    });

    mcpProcess.stderr?.on("data", (data: Buffer) => {
      const line = data.toString().trim();
      if (line) addLog(`[stderr] ${line}`);
    });

    mcpProcess.on("error", (err) => {
      addLog(`[error] ${err.message}`);
      mcpStatus = "error";
      mcpProcess = null;
    });

    mcpProcess.on("exit", (code) => {
      addLog(`[exit] MCP process exited with code ${code}`);
      mcpStatus = "stopped";
      mcpProcess = null;
    });

    // Wait briefly for it to start
    await new Promise((r) => setTimeout(r, 3000));
    const reachable = await probeMcp();
    if (reachable) mcpStatus = "running";

    return NextResponse.json({
      status: mcpStatus,
      pid: mcpProcess?.pid ?? null,
      message: reachable ? "MCP server started" : "MCP server starting (may take a few seconds for Go to compile)",
      reachable,
    });
  } catch (err) {
    mcpStatus = "error";
    mcpProcess = null;
    const msg = err instanceof Error ? err.message : String(err);
    addLog(`[error] Failed to spawn: ${msg}`);
    return NextResponse.json(
      { status: "error", message: msg },
      { status: 500 },
    );
  }
}
