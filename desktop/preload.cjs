const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("racoreDesktop", {
  isDesktop: true,
  daemonUrl: "http://127.0.0.1:47831",
  status: () => ipcRenderer.invoke("racore:daemon-status"),
  api: (request) => ipcRenderer.invoke("racore:api", request),
  platform: () => ipcRenderer.invoke("racore:platform"),
  openBrowser: (url) => ipcRenderer.invoke("racore:open-browser", url),
  openExternal: (url) => ipcRenderer.invoke("racore:open-external", url),
  onDaemonExit: (callback) =>
    ipcRenderer.on("racore:daemon-exit", (_event, code) => callback(code)),
});
