import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RacoreProductApp } from "../app/components/RacoreProductApp";
import "../app/globals.css";
import { installLegacyDesktopBridge } from "../app/lib/desktop";

const root = document.getElementById("root");

if (!root) {
  throw new Error("Racore desktop root element is missing");
}

document.documentElement.dataset.desktop = "tauri";
installLegacyDesktopBridge();

createRoot(root).render(
  <StrictMode>
    <RacoreProductApp />
  </StrictMode>,
);
