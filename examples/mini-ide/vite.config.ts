import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  base: process.env.BASE_PATH || "/",
  plugins: [react()],
  server: {
    port: 3001,
  },
  build: {
    outDir: "dist",
    sourcemap: true,
  },
});
