<script setup lang="ts">
import { computed, ref } from "vue";
import { getCurrentWindow } from "@tauri-apps/api/window";

const isMaximized = ref(false);
const appWindow = getCurrentWindow();

const maximizeLabel = computed(() => (isMaximized.value ? "还原" : "最大化"));

async function minimizeWindow(): Promise<void> {
    try {
        await appWindow.minimize();
    } catch {
        // Ignore when running in plain browser mode.
    }
}

async function toggleMaximizeWindow(): Promise<void> {
    try {
        await appWindow.toggleMaximize();
        isMaximized.value = await appWindow.isMaximized();
    } catch {
        // Ignore when running in plain browser mode.
    }
}

async function closeWindow(): Promise<void> {
    try {
        await appWindow.close();
    } catch {
        // Ignore when running in plain browser mode.
    }
}
</script>

<template>
    <header class="flex h-[38px] select-none items-center justify-between border-b border-slate-200 bg-white/90 pl-3"
        data-tauri-drag-region>
        <div class="flex h-full items-center" data-tauri-drag-region>
            <h1 class="text-sm font-semibold tracking-wide text-slate-700" data-tauri-drag-region>Helm Admin</h1>
        </div>
        <div class="flex h-full items-center" data-tauri-drag-region="false">
            <button
                class="h-full w-11 border-none bg-transparent text-base text-slate-700 transition-colors hover:bg-slate-100"
                title="最小化" @click="minimizeWindow">_</button>
            <button
                class="h-full w-11 border-none bg-transparent text-base text-slate-700 transition-colors hover:bg-slate-100"
                :title="maximizeLabel" @click="toggleMaximizeWindow">▢</button>
            <button
                class="h-full w-11 border-none bg-transparent text-base text-slate-700 transition-colors hover:bg-rose-600 hover:text-white"
                title="关闭" @click="closeWindow">×</button>
        </div>
    </header>
</template>
