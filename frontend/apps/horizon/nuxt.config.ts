import { fileURLToPath } from "node:url";
import { defineNuxtConfig } from "nuxt/config";

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
  devtools: { enabled: true },
});
