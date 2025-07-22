import preact from "@preact/preset-vite";
import { defineConfig } from "vite";

const now = new Date();

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [preact()],
  define: {
    __BUILD_DATE__: JSON.stringify(now.toISOString()),
  }
});
