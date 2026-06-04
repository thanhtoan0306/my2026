const { app, BrowserWindow, ipcMain } = require("electron");
const path = require("path");
const fs = require("fs");
const adb = require("./adb");

const SETTINGS_FILE = "settings.json";

function settingsPath() {
  return path.join(app.getPath("userData"), SETTINGS_FILE);
}

function loadSettings() {
  try {
    const raw = fs.readFileSync(settingsPath(), "utf8");
    return JSON.parse(raw);
  } catch {
    return { host: "", serial: "", shellHistory: [], terminalHistory: [] };
  }
}

function saveSettings(data) {
  fs.mkdirSync(path.dirname(settingsPath()), { recursive: true });
  fs.writeFileSync(settingsPath(), JSON.stringify(data, null, 2));
}

let settings = loadSettings();
let mainWindow = null;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1100,
    height: 780,
    minWidth: 900,
    minHeight: 600,
    title: "ADB Desktop",
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  mainWindow.loadFile(path.join(__dirname, "..", "src", "index.html"));
}

app.whenReady().then(() => {
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});

ipcMain.handle("settings:get", () => settings);

ipcMain.handle("settings:save", (_e, patch) => {
  settings = { ...settings, ...patch };
  saveSettings(settings);
  return settings;
});

ipcMain.handle("adb:devices", async (_e, { serial } = {}) => {
  if (settings.host) {
    await adb.connect(settings.host);
  }
  return adb.parseDeviceList(serial || settings.serial);
});

ipcMain.handle("adb:connect", async (_e, { host }) => {
  const h = host || settings.host;
  const res = await adb.connect(h);
  if (res.ok && h) {
    settings.host = adb.normalizeHost(h);
    saveSettings(settings);
  }
  return res;
});

ipcMain.handle("adb:disconnect", async (_e, { host }) => {
  return adb.disconnect(host || settings.host);
});

ipcMain.handle("adb:key", async (_e, { keycode, serial }) => {
  if (settings.host) await adb.connect(settings.host);
  return adb.keyEvent(serial || settings.serial, keycode);
});

ipcMain.handle("adb:text", async (_e, { text, serial }) => {
  if (settings.host) await adb.connect(settings.host);
  return adb.inputText(serial || settings.serial, text);
});

ipcMain.handle("adb:tap", async (_e, { x, y, serial }) => {
  if (settings.host) await adb.connect(settings.host);
  return adb.tap(serial || settings.serial, x, y);
});

ipcMain.handle("adb:shell", async (_e, { command, serial }) => {
  if (settings.host) await adb.connect(settings.host);
  return adb.shell(serial || settings.serial, command);
});

ipcMain.handle("adb:raw", async (_e, { args, serial }) => {
  const s = serial || settings.serial;
  const trimmed = (args || "").trim();
  if (trimmed.startsWith("adb ")) {
    return adb.raw(s, trimmed.slice(4));
  }
  return adb.raw(s, trimmed);
});

ipcMain.handle("adb:prop", async (_e, { prop, serial }) => {
  if (settings.host) await adb.connect(settings.host);
  return adb.getProp(serial || settings.serial, prop);
});

ipcMain.handle("adb:reboot", async (_e, { serial }) => {
  if (settings.host) await adb.connect(settings.host);
  return adb.reboot(serial || settings.serial);
});
