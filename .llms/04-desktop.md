# Desktop Shell

## Overview

The desktop shell is an Electron application that wraps the frontend and Go daemon into a native desktop application. It supports Windows, macOS, and Linux.

## Files

```
desktop/
  main.cjs            Electron main process
  preload.cjs         Context bridge (preload script)
  README.md           Desktop-specific documentation
  runtime/kubo/       Embedded Kubo IPFS binary (Windows .exe)
```

## Main Process (`desktop/main.cjs`)

### Startup Sequence

```
app.whenReady()
  -> startUiServer()          (packaged mode only: serves built UI on port 47832)
  -> startDaemon()            (spawns racored, waits for /health)
  -> createWindow()           (creates BrowserWindow, loads UI)
```

### Daemon Discovery (`daemonCommand()`)

The daemon binary is resolved in this priority order:

1. **Packaged mode**: `<resources>/racored/racored.exe` (bundled via `electron-builder`)
2. **Development mode**: `<project>/god/build/racored`
3. **Python fallback**: If `RACORE_PYTHON` env var is set, run `python -m racored.server`
4. **Legacy Python**: `python -m racored.server` (original Python backend)

The function first checks if the Go binary exists (`fs.existsSync`). If found, it uses Go. Otherwise, it falls back to Python for backward compatibility.

### Daemon Wait (`waitForDaemon()`)

- Polls `GET http://127.0.0.1:47831/health` every 350ms
- Timeout after 15 seconds
- Resolves when the daemon responds with HTTP 200
- Rejects with "racored did not become ready" on timeout

### UI Server (`startUiServer()`)

- Only starts in packaged mode (`app.isPackaged`)
- Serves the pre-built UI from `<resources>/ui/`
- Uses a simple HTTP server on port 47832
- Static assets served directly; all other requests go through the Vinext SSR worker

### Window Management

- **Main window**: 1480x940, dark background (`#090c10`), auto-hide menu bar
- **Child windows**: For external URLs opened from the browser, uses a persistent session partition (`persist:racore-web`)
- **Navigation**: URLs are loaded in the main window; external links open in the default browser

### IPC Handlers

| Channel | Handler | Description |
|---------|---------|-------------|
| `racore:daemon-status` | `GET /health` | Returns daemon health JSON |
| `racore:api` | Generic HTTP to daemon | Proxies any API request to the daemon |
| `racore:open-browser` | Creates child window | Opens URL in a new Electron window |
| `racore:open-external` | `shell.openExternal` | Opens URL in the system browser |
| `racore:platform` | Returns platform info | `{ platform, version, packaged }` |

The `racore:api` handler validates that the path starts with `/` and the method is one of `GET|POST|PUT|DELETE`, then forwards the request to `http://127.0.0.1:47831{path}`.

### Daemon Lifecycle

- **Start**: Spawned via `child_process.spawn()` on app startup
- **Environment**: `RACORE_KUBO_PATH` is set to the bundled Kubo binary path
- **Stdio**: Inherited in dev mode, ignored in packaged mode
- **Exit handling**: If daemon exits with non-zero code, sends `racore:daemon-exit` event to renderer
- **Shutdown**: Daemon is killed via `daemon.kill()` on `before-quit`

## Preload Script (`desktop/preload.cjs`)

Injects `window.racoreDesktop` into the renderer process via `contextBridge.exposeInMainWorld`:

```javascript
contextBridge.exposeInMainWorld('racoreDesktop', {
  isDesktop: true,
  daemonUrl: 'http://127.0.0.1:47831',
  status: () => ipcRenderer.invoke('racore:daemon-status'),
  api: (request) => ipcRenderer.invoke('racore:api', request),
  platform: () => ipcRenderer.invoke('racore:platform'),
  openBrowser: (url) => ipcRenderer.invoke('racore:open-browser', url),
  openExternal: (url) => ipcRenderer.invoke('racore:open-external', url),
  onDaemonExit: (callback) => {
    ipcRenderer.on('racore:daemon-exit', (_event, code) => callback(code));
  },
});
```

## Packaging (`electron-builder`)

Configured in `package.json` under the `"build"` key:

- **App ID**: `xyz.racore.browser`
- **Output directory**: `desktop-dist/`
- **Included files**: `desktop/**/*`, `package.json`
- **Extra resources**:
  - `dist/` -> `ui/` (built frontend)
  - `god/build/` -> `racored/` (Go daemon binaries)
  - `desktop/runtime/kubo/` -> `kubo/` (IPFS binary)
- **Windows target**: NSIS installer (`Racore-Browser-${version}-Setup.exe`)
- **Code signing**: Disabled (`signAndEditExecutable: false`)

## npm Scripts

| Script | Command | Description |
|--------|---------|-------------|
| `desktop` | `electron .` | Launch Electron with current build |
| `desktop:dev` | `concurrently -k -s first "npm run dev" "wait-on http://localhost:3000 && electron ."` | Dev mode with hot reload |
| `desktop:package` | `npm run build && electron-builder --win nsis` | Build + package for distribution |

## Security

- **Context isolation**: Enabled (`contextIsolation: true`)
- **Node integration**: Disabled (`nodeIntegration: false`)
- **Sandbox**: Enabled (`sandbox: true`)
- **Spellcheck**: Enabled
- **Permission handling**: Only allows `clipboard-sanitized-write`, `fullscreen`, `notifications`
- **Window creation**: All external URLs are opened in separate child windows with identical security settings
