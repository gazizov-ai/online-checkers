import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const proxyOptions = {
  changeOrigin: true,
  secure: false,
};

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/auth": {
        target: "http://localhost:8081",
        rewrite: (path) => path.replace(/^\/auth/, ""),
        ...proxyOptions,
      },
      "/matchmaking": {
        target: "http://localhost:8082",
        rewrite: (path) => path.replace(/^\/matchmaking/, ""),
        ...proxyOptions,
      },
      "/game": {
        target: "http://localhost:8083",
        rewrite: (path) => path.replace(/^\/game/, ""),
        ws: true,
        ...proxyOptions,
      },
      "/rating": {
        target: "http://localhost:8084",
        rewrite: (path) => path.replace(/^\/rating/, ""),
        ...proxyOptions,
      },
    },
  },
});
