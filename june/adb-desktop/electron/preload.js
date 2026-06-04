const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("adbDesktop", {
  getSettings: () => ipcRenderer.invoke("settings:get"),
  saveSettings: (patch) => ipcRenderer.invoke("settings:save", patch),
  devices: (opts) => ipcRenderer.invoke("adb:devices", opts || {}),
  connect: (opts) => ipcRenderer.invoke("adb:connect", opts || {}),
  disconnect: (opts) => ipcRenderer.invoke("adb:disconnect", opts || {}),
  key: (opts) => ipcRenderer.invoke("adb:key", opts),
  text: (opts) => ipcRenderer.invoke("adb:text", opts),
  tap: (opts) => ipcRenderer.invoke("adb:tap", opts),
  shell: (opts) => ipcRenderer.invoke("adb:shell", opts),
  raw: (opts) => ipcRenderer.invoke("adb:raw", opts),
  prop: (opts) => ipcRenderer.invoke("adb:prop", opts),
  reboot: (opts) => ipcRenderer.invoke("adb:reboot", opts || {}),
});
