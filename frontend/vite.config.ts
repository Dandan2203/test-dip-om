import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

/**
 * Конфігурація Vite для проєкту FinAgent (frontend).
 *
 * Proxy-правило перенаправляє всі запити /ping (та майбутні /api/*)
 * з dev-сервера (порт 5173) на Go-бекенд (порт 8080),
 * що вирішує проблему CORS під час локальної розробки.
 */
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // Всі запити, що починаються з /api або /ping → Go backend
      "/ping": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
