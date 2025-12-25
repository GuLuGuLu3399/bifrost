import type { Config } from "tailwindcss";
import uiPreset from "@bifrost/ui/tailwind.config";

const config: Config = {
  presets: [uiPreset as Config],
  content: [
    "./app/**/*.{vue,ts}",
    "./components/**/*.{vue,ts}",
    "./layouts/**/*.{vue,ts}",
    "./pages/**/*.{vue,ts}",
    "./plugins/**/*.{js,ts}",
    "../../packages/ui/**/*.{vue,ts}",
  ],
  theme: {
    extend: {},
  },
};

export default config;
