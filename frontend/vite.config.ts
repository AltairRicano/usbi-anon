import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Sin empaquetado de escritorio (Tauri) en ningún plan de este proyecto —
// decisión de F10.8 (plan/05_Contenido_maker_y_juego.md §5), no una omisión
// temporal: USBI-Anon es web, y el maker local exporta niveles como descarga
// de navegador (Blob + <a download>), igual que el fallback web que ya tenía
// ../usbi. No reintroducir la rama Tauri en MakerPage.tsx.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.VITE_BACKEND_URL ?? "http://localhost:8088",
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    // packages/* tiene su propio vitest.config.ts (ver frontend/packages/
    // engine/vitest.config.ts) y corre por separado vía
    // `npm run test --workspace=@usbi/engine` — excluido aquí para no
    // duplicar su ejecución.
    exclude: ["**/node_modules/**", "packages/**"],
  },
  preview: {
    // Sin esto, `vite preview` responde 403 "Blocked request" a cualquier
    // request cuyo header Host no sea localhost/IP — bloquea justo el tráfico
    // que llega vía el túnel de Cloudflare con Host: usbi.heimdall-lab.com.
    allowedHosts: ["usbi.heimdall-lab.com"],
    proxy: {
      "/api": {
        target: process.env.VITE_BACKEND_URL ?? "http://localhost:8088",
        changeOrigin: true,
      },
    },
  },
});
