import { defineConfig } from "vite";

export default defineConfig({
  build: { outDir: "dist", emptyOutDir: true, sourcemap: true },
  server: { port: 34115, strictPort: true },
});
