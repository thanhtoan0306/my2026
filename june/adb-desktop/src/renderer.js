const api = window.adbDesktop;

const $ = (id) => document.getElementById(id);

const hostEl = $("host");
const serialEl = $("serial");
const outputEl = $("output");
const statusEl = $("status-pill");
const deviceListEl = $("device-list");
const cmdDialog = $("cmd-dialog");
const cmdGrid = $("cmd-grid");

const PRESET_COMMANDS = [
  { cmd: "adb devices", desc: "List connected devices" },
  { cmd: "adb devices -l", desc: "List with details" },
  { cmd: "adb connect HOST", desc: "Network connect (uses host field)" },
  { cmd: "adb disconnect HOST", desc: "Disconnect network device" },
  { cmd: "adb tcpip 5555", desc: "Enable TCP (USB connected first)" },
  { cmd: "adb kill-server", desc: "Restart ADB daemon" },
  { cmd: "adb start-server", desc: "Start ADB daemon" },
  { cmd: "adb shell getprop ro.product.model", desc: "Device model" },
  { cmd: "adb shell pm list packages | head -20", desc: "List packages" },
  { cmd: "adb shell input keyevent 3", desc: "Home key" },
  { cmd: "adb shell screencap -p /sdcard/screen.png", desc: "Screenshot to /sdcard" },
  { cmd: "adb logcat -d | tail -50", desc: "Last log lines" },
  { cmd: "adb reboot", desc: "Reboot device" },
];

function setStatus(text, kind = "") {
  statusEl.textContent = text;
  statusEl.className = "status" + (kind ? ` ${kind}` : "");
}

function appendOutput(label, result) {
  const ts = new Date().toLocaleTimeString();
  const prefix = result.ok ? "✓" : "✗";
  const block = `[${ts}] ${prefix} ${label}\n${result.output || "(no output)"}\n\n`;
  outputEl.textContent = block + outputEl.textContent;
}

function hostPlaceholder() {
  const h = (hostEl.value || "").trim();
  return h || "192.168.1.50:5555";
}

function fillPresetCommands() {
  cmdGrid.innerHTML = "";
  for (const item of PRESET_COMMANDS) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "cmd-item";
    let cmd = item.cmd.replace(/HOST/g, hostPlaceholder());
    btn.innerHTML = `${cmd}<small>${item.desc}</small>`;
    btn.onclick = () => {
      $("terminal-cmd").value = cmd;
      cmdDialog.close();
    };
    cmdGrid.appendChild(btn);
  }
}

function renderDevices(devices) {
  deviceListEl.innerHTML = "";
  if (!devices || !devices.length) {
    const li = document.createElement("li");
    li.textContent = "No devices";
    li.style.cursor = "default";
    deviceListEl.appendChild(li);
    return;
  }
  const current = serialEl.value.trim();
  for (const d of devices) {
    const li = document.createElement("li");
    li.innerHTML = `${d.serial}<span class="state">${d.state}${d.info ? " · " + d.info : ""}</span>`;
    if (d.serial === current) li.classList.add("selected");
    li.onclick = () => {
      serialEl.value = d.serial;
      document.querySelectorAll(".device-list li").forEach((el) => el.classList.remove("selected"));
      li.classList.add("selected");
      api.saveSettings({ serial: d.serial });
    };
    deviceListEl.appendChild(li);
  }
}

async function refreshDevices() {
  setStatus("Refreshing…", "busy");
  const res = await api.devices({ serial: serialEl.value.trim() });
  if (res.devices) renderDevices(res.devices);
  setStatus(res.ok ? "Devices updated" : "ADB error", res.ok ? "ok" : "err");
  if (!res.ok) appendOutput("devices", res);
}

async function runKey(keycode) {
  setStatus(`Key ${keycode}…`, "busy");
  const res = await api.key({ keycode, serial: serialEl.value.trim() });
  appendOutput(`keyevent ${keycode}`, res);
  setStatus(res.ok ? "Sent" : "Failed", res.ok ? "ok" : "err");
}

async function init() {
  const s = await api.getSettings();
  hostEl.value = s.host || "";
  serialEl.value = s.serial || "";
  fillPresetCommands();
  await refreshDevices();

  const modelRes = await api.prop({
    prop: "ro.product.model",
    serial: s.serial,
  });
  if (modelRes.ok && modelRes.output) {
    setStatus(modelRes.output.trim(), "ok");
  }
}

$("btn-save").onclick = async () => {
  await api.saveSettings({
    host: hostEl.value.trim(),
    serial: serialEl.value.trim(),
  });
  fillPresetCommands();
  setStatus("Settings saved", "ok");
};

$("btn-connect").onclick = async () => {
  setStatus("Connecting…", "busy");
  const res = await api.connect({ host: hostEl.value.trim() });
  appendOutput("connect", res);
  setStatus(res.ok ? "Connected" : "Connect failed", res.ok ? "ok" : "err");
  if (res.ok) await api.saveSettings({ host: hostEl.value.trim() });
  await refreshDevices();
};

$("btn-disconnect").onclick = async () => {
  const res = await api.disconnect({ host: hostEl.value.trim() });
  appendOutput("disconnect", res);
  setStatus(res.ok ? "Disconnected" : "Failed", res.ok ? "ok" : "err");
  await refreshDevices();
};

$("btn-refresh").onclick = refreshDevices;

document.querySelectorAll("[data-key]").forEach((btn) => {
  btn.onclick = () => runKey(btn.dataset.key);
});

$("btn-send-text").onclick = async () => {
  const text = $("input-text").value;
  if (!text) return;
  const res = await api.text({ text, serial: serialEl.value.trim() });
  appendOutput("input text", res);
  setStatus(res.ok ? "Text sent" : "Failed", res.ok ? "ok" : "err");
};

$("btn-tap").onclick = async () => {
  const x = parseInt($("tap-x").value, 10);
  const y = parseInt($("tap-y").value, 10);
  if (Number.isNaN(x) || Number.isNaN(y)) {
    setStatus("Enter X and Y", "err");
    return;
  }
  const res = await api.tap({ x, y, serial: serialEl.value.trim() });
  appendOutput(`tap ${x} ${y}`, res);
  setStatus(res.ok ? "Tap sent" : "Failed", res.ok ? "ok" : "err");
};

$("btn-shell").onclick = async () => {
  const command = $("shell-cmd").value.trim();
  if (!command) return;
  setStatus("Shell…", "busy");
  const res = await api.shell({ command, serial: serialEl.value.trim() });
  appendOutput(`shell: ${command}`, res);
  setStatus(res.ok ? "Done" : "Shell failed", res.ok ? "ok" : "err");
};

$("shell-cmd").addEventListener("keydown", (e) => {
  if (e.key === "Enter") $("btn-shell").click();
});

$("btn-terminal").onclick = async () => {
  const args = $("terminal-cmd").value.trim();
  if (!args) return;
  setStatus("Running…", "busy");
  const res = await api.raw({ args, serial: serialEl.value.trim() });
  appendOutput(args, res);
  setStatus(res.ok ? "Done" : "Command failed", res.ok ? "ok" : "err");
};

$("terminal-cmd").addEventListener("keydown", (e) => {
  if (e.key === "Enter") $("btn-terminal").click();
});

$("btn-cmd-help").onclick = () => {
  fillPresetCommands();
  cmdDialog.showModal();
};

$("btn-reboot").onclick = async () => {
  if (!confirm("Reboot the selected device?")) return;
  const res = await api.reboot({ serial: serialEl.value.trim() });
  appendOutput("reboot", res);
  setStatus(res.ok ? "Reboot sent" : "Failed", res.ok ? "ok" : "err");
};

hostEl.addEventListener("change", fillPresetCommands);

init();
