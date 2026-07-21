const { app, BrowserWindow, ipcMain, shell, session } = require("electron");
const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const http = require("http");
const { pathToFileURL } = require("url");

let mainWindow;
let daemon;
let uiServer;
const DAEMON_URL = "http://127.0.0.1:47831";
const UI_URL = "http://127.0.0.1:47832";

const mimeTypes = {
  ".css": "text/css; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".png": "image/png",
  ".svg": "image/svg+xml",
  ".woff2": "font/woff2",
};

function projectRoot() {
  return app.isPackaged ? process.resourcesPath : path.resolve(__dirname, "..");
}

function daemonCommand() {
  if (process.env.RACORE_PYTHON)
    return {
      executable: process.env.RACORE_PYTHON,
      args: ["-m", "racored.server"],
    };
  const root = projectRoot();
  if (app.isPackaged)
    return { executable: path.join(root, "racored", "racored.exe"), args: [] };
  const candidates = [
    path.join(root, ".venv", "Scripts", "python.exe"),
    "python",
  ];
  return {
    executable: candidates.find(
      (candidate) => candidate === "python" || fs.existsSync(candidate),
    ),
    args: ["-m", "racored.server"],
  };
}

function waitForDaemon(timeout = 15000) {
  const started = Date.now();
  return new Promise((resolve, reject) => {
    const check = () => {
      const request = http.get(`${DAEMON_URL}/health`, (response) => {
        response.resume();
        if (response.statusCode === 200) resolve(true);
        else retry();
      });
      request.on("error", retry);
      request.setTimeout(800, () => {
        request.destroy();
        retry();
      });
    };
    const retry = () =>
      Date.now() - started > timeout
        ? reject(new Error("racored did not become ready"))
        : setTimeout(check, 350);
    check();
  });
}

async function startUiServer() {
  if (!app.isPackaged) return;
  const uiRoot = path.join(projectRoot(), "ui");
  const clientRoot = path.join(uiRoot, "client");
  const { default: worker } = await import(
    `${pathToFileURL(path.join(uiRoot, "server", "index.js")).href}?desktop=${Date.now()}`
  );
  const assetResponse = async (request) => {
    const pathname = decodeURIComponent(new URL(request.url).pathname);
    const candidate = path.resolve(clientRoot, `.${pathname}`);
    if (
      !candidate.startsWith(`${path.resolve(clientRoot)}${path.sep}`) ||
      !fs.existsSync(candidate) ||
      !fs.statSync(candidate).isFile()
    ) {
      return new Response("Not found", { status: 404 });
    }
    return new Response(fs.readFileSync(candidate), {
      headers: {
        "content-type":
          mimeTypes[path.extname(candidate)] || "application/octet-stream",
      },
    });
  };
  uiServer = http.createServer(async (request, response) => {
    try {
      const url = new URL(request.url || "/", UI_URL);
      const asset = await assetResponse(new Request(url));
      const result =
        asset.status !== 404
          ? asset
          : await worker.fetch(
              new Request(url, {
                method: request.method,
                headers: request.headers,
              }),
              { ASSETS: { fetch: assetResponse } },
              { waitUntil() {}, passThroughOnException() {} },
            );
      response.writeHead(
        result.status,
        Object.fromEntries(result.headers.entries()),
      );
      response.end(Buffer.from(await result.arrayBuffer()));
    } catch (error) {
      response.writeHead(500, { "content-type": "text/plain; charset=utf-8" });
      response.end(`Racore UI failed: ${error.message}`);
    }
  });
  await new Promise((resolve, reject) => {
    uiServer.once("error", reject);
    uiServer.listen(47832, "127.0.0.1", resolve);
  });
}

async function startDaemon() {
  try {
    const response = await fetch(`${DAEMON_URL}/health`);
    if (response.ok) return;
  } catch {}
  const root = projectRoot();
  const moduleRoot = path.join(root, "python");
  const command = daemonCommand();
  daemon = spawn(command.executable, command.args, {
    cwd: moduleRoot,
    env: {
      ...process.env,
      PYTHONPATH: moduleRoot,
      RACORE_KUBO_PATH: app.isPackaged
        ? path.join(root, "kubo", "ipfs.exe")
        : path.join(root, "desktop", "runtime", "kubo", "ipfs.exe"),
    },
    windowsHide: true,
    stdio: app.isPackaged ? "ignore" : "inherit",
  });
  daemon.on("exit", (code) => {
    if (code && mainWindow && !mainWindow.isDestroyed())
      mainWindow.webContents.send("racore:daemon-exit", code);
  });
  await waitForDaemon().catch((error) => console.error(error.message));
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1480,
    height: 940,
    minWidth: 980,
    minHeight: 680,
    backgroundColor: "#090c10",
    title: "Racore Browser",
    autoHideMenuBar: true,
    webPreferences: {
      preload: path.join(__dirname, "preload.cjs"),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
      spellcheck: true,
    },
  });
  const ui =
    process.env.RACORE_UI_URL ||
    (app.isPackaged ? UI_URL : "http://localhost:3000");
  mainWindow.loadURL(`${ui}${ui.includes("?") ? "&" : "?"}desktop=1`);
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    if (/^https?:/.test(url)) openBrowserWindow(url);
    return { action: "deny" };
  });
}

function openBrowserWindow(url) {
  let target = url.trim();
  if (!/^https?:\/\//i.test(target)) target = `https://${target}`;
  const child = new BrowserWindow({
    width: 1280,
    height: 820,
    parent: mainWindow,
    title: "Racore Web",
    backgroundColor: "#090c10",
    autoHideMenuBar: true,
    webPreferences: {
      partition: "persist:racore-web",
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });
  child
    .loadURL(target)
    .catch(() =>
      child.loadURL(
        `https://www.google.com/search?q=${encodeURIComponent(url)}`,
      ),
    );
  child.webContents.setWindowOpenHandler(({ url: next }) => {
    child.loadURL(next);
    return { action: "deny" };
  });
  return true;
}

app.whenReady().then(async () => {
  session.defaultSession.setPermissionRequestHandler(
    (_contents, permission, callback) =>
      callback(
        ["clipboard-sanitized-write", "fullscreen", "notifications"].includes(
          permission,
        ),
      ),
  );
  await startUiServer();
  await startDaemon();
  createWindow();
  app.on(
    "activate",
    () => BrowserWindow.getAllWindows().length === 0 && createWindow(),
  );
});

ipcMain.handle("racore:daemon-status", async () => {
  try {
    const response = await fetch(`${DAEMON_URL}/health`);
    return await response.json();
  } catch (error) {
    return { ok: false, error: error.message };
  }
});
ipcMain.handle("racore:api", async (_event, request) => {
  const method = String(request?.method || "GET").toUpperCase();
  const apiPath = String(request?.path || "/health");
  if (
    !apiPath.startsWith("/") ||
    !["GET", "POST", "PUT", "DELETE"].includes(method)
  )
    throw new Error("Invalid local API request");
  const response = await fetch(`${DAEMON_URL}${apiPath}`, {
    method,
    headers: request?.body ? { "Content-Type": "application/json" } : undefined,
    body: request?.body ? JSON.stringify(request.body) : undefined,
  });
  const text = await response.text();
  let data;
  try {
    data = JSON.parse(text);
  } catch {
    data = { text };
  }
  return { ok: response.ok, status: response.status, data };
});
ipcMain.handle("racore:open-browser", (_event, url) => openBrowserWindow(url));
ipcMain.handle("racore:open-external", (_event, url) =>
  shell.openExternal(url),
);
ipcMain.handle("racore:platform", () => ({
  platform: process.platform,
  version: app.getVersion(),
  packaged: app.isPackaged,
}));

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});
app.on("before-quit", () => {
  if (daemon && !daemon.killed) daemon.kill();
  if (uiServer) uiServer.close();
});
