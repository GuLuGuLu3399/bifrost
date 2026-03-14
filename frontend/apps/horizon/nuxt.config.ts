import { fileURLToPath } from "node:url";
import { defineNuxtConfig } from "nuxt/config";
import tailwindcss from "@tailwindcss/vite";

const uiSrcPath = fileURLToPath(
  new URL("../../packages/ui/src", import.meta.url)
);

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  alias: {
    "@bifrost/ui": uiSrcPath,
  },
  build: {
    transpile: ["@bifrost/ui"],
  },
  css: ["~/assets/css/main.css"],
  vite: {
    plugins: [tailwindcss()],
  },
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE || "http://localhost:8080",
      cdnUrl: process.env.CDN_URL || "https://cdn.example.com",
    },
  },
  imports: {
    autoImport: true,
  },
  components: {
    dirs: ["~/components"],
  },
});
