<script setup lang="ts">
import { getCurrentWindow } from "@tauri-apps/api/window";
import { X, Minus, Square } from "lucide-vue-next";
const appWindow = getCurrentWindow();

const minimize = () => appWindow.minimize();
const toggleMaximize = async () => {
  const isMaximized = await appWindow.isMaximized();
  isMaximized ? appWindow.unmaximize() : appWindow.maximize();
};
const close = () => appWindow.close();
</script>

<template>
  <header
    data-tauri-drag-region
    class="flex h-10 w-full shrink-0 items-center justify-between border-b bg-background px-3 select-none z-50"
  >
    <div
      class="flex items-center gap-2 text-xs font-mono font-bold tracking-widest text-muted-foreground pointer-events-none"
    >
      HELM <span class="text-primary">///</span> ACCESS_TERMINAL
    </div>

    <div class="flex items-center gap-1 no-drag">
      <button
        @click="minimize"
        class="titlebar-btn p-2 rounded-sm hover-bg-muted hover-text-primary"
        aria-label="Minimize"
      >
        <Minus class="h-3 w-3" />
      </button>
      <button
        @click="toggleMaximize"
        class="titlebar-btn p-2 rounded-sm hover-bg-muted hover-text-primary"
        aria-label="Maximize"
      >
        <Square class="h-3 w-3" />
      </button>
      <button
        @click="close"
        class="titlebar-btn p-2 rounded-sm hover-bg-destructive hover-text-destructive"
        aria-label="Close"
      >
        <X class="h-3 w-3" />
      </button>
    </div>
  </header>
</template>

<style scoped>
/* No extra styles; relies on global utility classes */
.no-drag {
  -webkit-app-region: no-drag;
}
</style>
