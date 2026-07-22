# Tauri desktop development

Racore now has a Tauri v2 scaffold alongside the Electron shell during the staged migration. React components remain under `app/` and are consumed by two build pipelines:

- `npm run build` keeps the existing Vinext/Cloudflare hosted workflow.
- `npm run desktop:ui` creates static React assets in `dist-desktop/` for the Tauri webview.

The desktop entry is `desktop-ui/main.tsx`. Its Vite configuration aliases `next/image` to a web-standard image component so the same React UI can run without a Next.js runtime. Public assets continue to come from `public/`.

## Commands

```powershell
npm run desktop:ui
npm run tauri:dev
npm run tauri:build
```

`src-tauri/tauri.conf.json` retains the Electron main-window dimensions and uses a production content security policy. `src-tauri/capabilities/main.json` applies only to the trusted `main` window. Remote browsing windows will use different labels and will not inherit this capability when they are implemented.

The scaffold intentionally does not expose filesystem, operating-system, or shell plugin APIs to React. Desktop-native operations are implemented as application commands in the next migration step.
