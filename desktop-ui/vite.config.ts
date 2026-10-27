import react from "@vitejs/plugin-react";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

const projectPath = (path: string) =>
  fileURLToPath(new URL(`../${path}`, import.meta.url));

export default defineConfig({
  root: projectPath("desktop-ui"),
  // Relative asset URLs keep the build loadable from tauri://localhost and from
  // any /ipfs/<cid>/ path on a gateway, not just a domain root.
  base: "./",
  publicDir: projectPath("public"),
  plugins: [react()],
  resolve: {
    alias: {
      "next/image": projectPath("desktop-ui/next-image.tsx"),
    },
  },
  server: {
    host: "127.0.0.1",
    port: 1420,
    strictPort: true,
  },
  build: {
    outDir: projectPath("dist-desktop"),
    emptyOutDir: true,
  },
});
