# Frontend

The UI uses React and TypeScript under `app/`.

## Build targets

- `npm run build`: Vinext server/client output for the hosted Cloudflare workflow.
- `npm run desktop:ui`: static Vite output in `dist-desktop/` for Tauri.

`desktop-ui/main.tsx` renders the shared application and marks the document as a Tauri desktop environment. Its `next/image` alias supplies a standard static image implementation without requiring a Next runtime in the native bundle.

## Data access

`app/lib/racore-client.ts` is the only daemon API client. In desktop mode it calls `desktopBridge.api`; in hosted mode it uses loopback HTTP directly.

`app/lib/desktop.ts` defines typed wrappers for:

- `daemon_status`
- `daemon_request`
- `platform_info`
- `open_browser`
- `open_external`
- `racore://daemon-exit`

Event listeners resolve to an unsubscribe function. Components should release it during unmount.
