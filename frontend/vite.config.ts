import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const backendTarget = "http://localhost:8080";
const mlTarget = "http://localhost:8000";
const agentsTarget = "http://localhost:8010";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      "/upload_data": { target: backendTarget, changeOrigin: true },
      "/upload_events": { target: backendTarget, changeOrigin: true },
      "/events": { target: backendTarget, changeOrigin: true },
      "/farmer_data": { target: backendTarget, changeOrigin: true },
      "/farmers": { target: backendTarget, changeOrigin: true },
      "/load_orders": { target: backendTarget, changeOrigin: true },
      "/clients": { target: backendTarget, changeOrigin: true },
      "/client_bonus": { target: backendTarget, changeOrigin: true },
      "/spend_bonus": { target: backendTarget, changeOrigin: true },
      "/recommendations": { target: backendTarget, changeOrigin: true },
      "/vector": { target: backendTarget, changeOrigin: true },
      "/health": { target: backendTarget, changeOrigin: true },
      "/ml": {
        target: mlTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/ml/, ""),
      },
      "/agents": { target: agentsTarget, changeOrigin: true },
    },
  },
});
