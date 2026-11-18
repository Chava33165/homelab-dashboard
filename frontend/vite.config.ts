import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
    // El backend de Go sirve esto embebido (//go:embed) -- assets con hash largo
    // sirven bien detras de cache-control agresivo cuando llegue el service worker.
    assetsDir: "assets",
  },
  server: {
    // Para "vite dev" contra el backend real corriendo en :8080 (no se usa en produccion,
    // ahi el propio binario de Go sirve los estaticos).
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
