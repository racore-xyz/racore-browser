import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "desktop/**",
    "desktop-dist/**",
    "dist/**",
    "dist-desktop/**",
    "website/dist/**",
    "website/.vinext/**",
    "src-tauri/target/**",
    "src-tauri/gen/**",
    "src-tauri/binaries/**",
    "src-tauri/resources/**",
    "work/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
