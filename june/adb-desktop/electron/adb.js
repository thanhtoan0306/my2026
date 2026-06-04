const { execFile } = require("child_process");
const { promisify } = require("util");

const execFileAsync = promisify(execFile);

function normalizeHost(host) {
  const h = (host || "").trim();
  if (!h) return "";
  if (h.includes(":")) return h;
  return `${h}:5555`;
}

async function runAdb(args, serial = "") {
  const base = serial ? ["-s", serial, ...args] : args;
  try {
    const { stdout, stderr } = await execFileAsync("adb", base, {
      maxBuffer: 10 * 1024 * 1024,
      timeout: 120000,
    });
    const out = [stdout, stderr].filter(Boolean).join("").trim();
    return { ok: true, output: out };
  } catch (err) {
    const out = [err.stdout, err.stderr].filter(Boolean).join("").trim();
    return { ok: false, output: out || err.message };
  }
}

function parseDevices(text) {
  const lines = text.split("\n").slice(1);
  const devices = [];
  for (const line of lines) {
    const m = line.trim().match(/^(\S+)\s+(\S+)(?:\s+(.*))?$/);
    if (!m || m[2] === "List") continue;
    devices.push({
      serial: m[1],
      state: m[2],
      info: (m[3] || "").trim(),
    });
  }
  return devices;
}

module.exports = {
  normalizeHost,

  async devices(serial) {
    return runAdb(["devices", "-l"], serial);
  },

  async parseDeviceList(serial) {
    const res = await runAdb(["devices", "-l"], "");
    if (!res.ok) return res;
    return { ok: true, output: res.output, devices: parseDevices(res.output) };
  },

  async connect(host) {
    const h = normalizeHost(host);
    if (!h) return { ok: false, output: "Enter device IP or host:port" };
    return runAdb(["connect", h]);
  },

  async disconnect(host) {
    const h = normalizeHost(host);
    if (!h) return { ok: false, output: "Enter device IP first" };
    return runAdb(["disconnect", h]);
  },

  async keyEvent(serial, keycode) {
    return runAdb(["shell", "input", "keyevent", String(keycode)], serial);
  },

  async inputText(serial, text) {
    const escaped = (text || "").replace(/ /g, "%s");
    return runAdb(["shell", "input", "text", escaped], serial);
  },

  async tap(serial, x, y) {
    return runAdb(["shell", "input", "tap", String(x), String(y)], serial);
  },

  async swipe(serial, x1, y1, x2, y2, durationMs = 300) {
    return runAdb(
      [
        "shell",
        "input",
        "swipe",
        String(x1),
        String(y1),
        String(x2),
        String(y2),
        String(durationMs),
      ],
      serial
    );
  },

  async shell(serial, command) {
    const cmd = command.trim();
    if (!cmd) return { ok: false, output: "Empty shell command" };
    return runAdb(["shell", "sh", "-c", cmd], serial);
  },

  async raw(serial, args) {
    const parts = args.trim().split(/\s+/).filter(Boolean);
    if (!parts.length) return { ok: false, output: "Empty command" };
    return runAdb(parts, serial);
  },

  async getProp(serial, prop) {
    return runAdb(["shell", "getprop", prop], serial);
  },

  async reboot(serial) {
    return runAdb(["reboot"], serial);
  },
};
