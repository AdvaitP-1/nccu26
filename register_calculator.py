#!/usr/bin/env python3
"""Register the other IDE's calculator.py through the MCP system."""

import json
import queue
import threading
import time
import requests

MCP_URL = "http://localhost:9090"
responses = queue.Queue()
endpoint_url = queue.Queue()


def sse_reader(url):
    resp = requests.get(f"{url}/sse", stream=True, timeout=60)
    buf, ev = "", None
    for chunk in resp.iter_content(chunk_size=1, decode_unicode=True):
        buf += chunk
        while "\n" in buf:
            line, buf = buf.split("\n", 1)
            line = line.rstrip("\r")
            if line.startswith("event:"):
                ev = line[6:].strip()
            elif line.startswith("data:"):
                data = line[5:].strip()
                if ev == "endpoint":
                    endpoint_url.put(data)
                elif ev == "message":
                    try:
                        responses.put(json.loads(data))
                    except json.JSONDecodeError:
                        pass
                ev = None
            elif line == "":
                ev = None


_id = 0


def rpc(method, params=None):
    global _id
    _id += 1
    payload = {"jsonrpc": "2.0", "method": method, "id": _id}
    if params:
        payload["params"] = params
    requests.post(ep, json=payload, timeout=30)
    deadline = time.time() + 30
    while time.time() < deadline:
        try:
            msg = responses.get(timeout=1)
            if msg.get("id") == _id:
                return msg
        except queue.Empty:
            continue
    return {}


def tool(name, args=None):
    r = rpc("tools/call", {"name": name, "arguments": args or {}})
    for b in r.get("result", {}).get("content", []):
        if b.get("type") == "text":
            return b["text"]
    return json.dumps(r)


# Connect
print("Connecting to MCP...")
threading.Thread(target=sse_reader, args=(MCP_URL,), daemon=True).start()
ep = endpoint_url.get(timeout=10)
print(f"Connected: {ep}")

rpc("initialize", {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {"name": "register_calculator", "version": "1.0"},
})
requests.post(ep, json={"jsonrpc": "2.0", "method": "notifications/initialized"}, timeout=10)

# Read the calculator code from the other IDE's workspace
with open("/Users/advaitpandey/Documents/est/calculator.py") as f:
    code = f.read()

print(f"\nRead calculator.py ({len(code)} chars, {code.count(chr(10))} lines)")

# Register as a push
print("\n=== Registering push from other-ide-agent ===")
files_json = json.dumps([{
    "file_path": "calculator.py",
    "language": "python",
    "base_content": "",
    "new_content": code,
}])
result = tool("register_push", {
    "branch_name": "feature/calculator",
    "user_id": "other-ide-agent",
    "files_json": files_json,
    "message": "Add tkinter calculator GUI",
})
print(result)

# Check VFS
print("\n=== VFS State ===")
vfs = json.loads(tool("get_vfs_state"))
print(f"Agents: {vfs['total_agents']}")
print(f"Files:  {vfs['total_files']}")
for pc in vfs["pending_changes"]:
    for f in pc["files"]:
        print(f"  {pc['agent_id']:45s} -> {f['path']}")

# Run overlap analysis
print("\n=== Overlap Analysis ===")
time.sleep(0.5)
analysis = json.loads(tool("identify_overlaps"))
overlaps = analysis.get("overlaps", [])
if overlaps:
    print(f"Found {len(overlaps)} overlap(s):")
    for o in overlaps:
        print(f"  {o['file_path']}: {o['symbol_name']} -- {o['severity'].upper()}")
        print(f"    {o['agent_a']} vs {o['agent_b']}")
else:
    print(f"Result: {analysis.get('note', 'No overlaps')}")

risks = analysis.get("file_risks", [])
if risks:
    print(f"\nFile risks:")
    for r in risks:
        tag = " HOTSPOT" if r["is_hotspot"] else ""
        print(f"  {r['file_path']}: risk={r['risk_score']}/100{tag}")

print("\n=== Done ===")
print("The other IDE's calculator code is now tracked in the MCP system.")
