const API_PORT = import.meta.env.VITE_SIDECAR_PORT || "19527";
const API = `http://127.0.0.1:${API_PORT}`;

const $ = (id) => document.getElementById(id);

function cfg() {
  return {
    adb_host: $("host").value.trim(),
    adb_serial: $("serial").value.trim(),
  };
}

function flash(text, kind = "muted") {
  const el = $("flash");
  el.textContent = text;
  el.className = `flash ${kind}`;
}

async function api(path, opts = {}) {
  const res = await fetch(`${API}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...opts,
  });
  return res.json();
}

async function withConfig(path, body = {}) {
  return api(path, {
    method: "POST",
    body: JSON.stringify({ config: cfg(), ...body }),
  });
}

async function loadConfig() {
  const c = await api("/api/config");
  $("host").value = c.adb_host || "";
  $("serial").value = c.adb_serial || "";
}

async function saveConfig() {
  await api("/api/config", { method: "POST", body: JSON.stringify(cfg()) });
  flash("Settings saved.", "ok");
}

async function refreshDevices() {
  const r = await api("/api/devices");
  $("devices").textContent = r.ok ? r.output || "(empty)" : r.error;
  if (!r.ok) flash(r.error, "err");
}

async function loadApps() {
  const r = await api("/api/apps");
  const ul = $("apps");
  ul.innerHTML = "";
  if (!r.ok) {
    flash(r.error || "Failed to load apps", "err");
    return;
  }
  for (const app of r.apps || []) {
    const li = document.createElement("li");
    const code = document.createElement("code");
    code.textContent = app.package;
    const btn = document.createElement("button");
    btn.textContent = "Stop";
    btn.onclick = async () => {
      const res = await withConfig("/api/apps/stop", { package: app.package });
      flash(res.ok ? res.output || "Stopped" : res.error, res.ok ? "ok" : "err");
      loadApps();
    };
    li.append(code, btn);
    ul.append(li);
  }
}

function bindRemote() {
  const keys = [
    ["▲", "19"], ["◀", "21"], ["●", "23"], ["▶", "22"], ["▼", "20"],
    ["Home", "3"], ["Back", "4"], ["Menu", "82"], ["Vol+", "24"], ["Vol−", "25"],
    ["Power", "26"], ["Recent", "187"], ["Enter", "66"],
  ];
  const row = $("remote-keys");
  for (const [label, code] of keys) {
    const b = document.createElement("button");
    b.textContent = label;
    b.onclick = async () => {
      const r = await withConfig("/api/key", { keycode: code });
      flash(r.ok ? r.output || `key ${code}` : r.error, r.ok ? "ok" : "err");
    };
    row.append(b);
  }
}

async function waitForSidecar(retries = 30) {
  for (let i = 0; i < retries; i++) {
    try {
      const h = await api("/api/health");
      flash(`Sidecar ready · adb: ${h.adb}`, "ok");
      return true;
    } catch {
      await new Promise((r) => setTimeout(r, 300));
    }
  }
  flash(`Cannot reach Go sidecar at ${API}. Run: npm run sidecar`, "err");
  return false;
}

$("btn-save").onclick = saveConfig;
$("btn-devices").onclick = async () => {
  await saveConfig();
  await refreshDevices();
};
$("btn-connect").onclick = async () => {
  const r = await withConfig("/api/connect");
  flash(r.ok ? r.output || "Connected" : r.error, r.ok ? "ok" : "err");
  refreshDevices();
};
$("btn-disconnect").onclick = async () => {
  const r = await withConfig("/api/disconnect");
  flash(r.ok ? r.output || "Disconnected" : r.error, r.ok ? "ok" : "err");
};
$("btn-text").onclick = async () => {
  const r = await withConfig("/api/text", { text: $("text-input").value });
  flash(r.ok ? r.output || "Sent" : r.error, r.ok ? "ok" : "err");
};
$("btn-tap").onclick = async () => {
  const r = await withConfig("/api/tap", { x: $("tap-x").value, y: $("tap-y").value });
  flash(r.ok ? r.output || "Tap sent" : r.error, r.ok ? "ok" : "err");
};
$("btn-shell").onclick = async () => {
  const r = await withConfig("/api/shell", { command: $("shell-cmd").value });
  flash(r.ok ? r.output || "OK" : r.error, r.ok ? "ok" : "err");
};
$("btn-term").onclick = async () => {
  const r = await withConfig("/api/terminal", { command: $("term-cmd").value });
  flash(r.ok ? r.output || "OK" : r.error, r.ok ? "ok" : "err");
};
$("btn-close-all").onclick = async () => {
  if (!confirm("Force-stop ALL third-party apps?")) return;
  const r = await withConfig("/api/apps/close-all");
  flash(r.ok ? r.output : r.error, r.ok ? "ok" : "err");
  loadApps();
};

bindRemote();

(async () => {
  if (await waitForSidecar()) {
    await loadConfig();
    await refreshDevices();
    await loadApps();
  }
})();
