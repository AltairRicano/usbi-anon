import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Greenfield para F10 (plan/04_Rediseno_identidad_gustos.md §5): sin Tauri
// todavía — esta fase solo cubre registro/login/paneles de admin, no los
// minijuegos que justificarían el empaquetado de escritorio. El wrapper
// Tauri se añade cuando se porte el resto de la app de juego (dashboard,
// niveles, sync offline), no antes.
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
  preview: {
    proxy: {
      "/api": {
        target: process.env.VITE_BACKEND_URL ?? "http://localhost:8088",
        changeOrigin: true,
      },
    },
  },
});
