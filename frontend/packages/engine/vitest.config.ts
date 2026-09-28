import { defineConfig } from "vitest/config";

// TriviaEngine.startTimer usa window.setInterval — sin jsdom, vitest corre
// en Node puro y revienta con "window is not defined". Configuración propia
// del paquete (no heredada del vite.config.ts de la app) porque el motor de
// juego es la única pieza que necesita un DOM simulado para sus pruebas.
export default defineConfig({
  test: {
    environment: "jsdom",
    exclude: ["dist/**", "node_modules/**"],
  },
});
