<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useSessionState } from "@/composables/useSessionState";

const { authenticated, loading, refreshSession, logoutSession, lastCheckedAt } =
    useSessionState();

const authText = computed(() => (authenticated.value ? "已登录" : "未登录"));

const checkedText = computed(() => {
    if (!lastCheckedAt.value) {
        return "等待首次检查";
    }

    return new Date(lastCheckedAt.value).toLocaleTimeString();
});

const links = [
    { to: "/", label: "概览" },
    { to: "/posts", label: "文章管理" },
    { to: "/taxonomy", label: "分类标签" },
    { to: "/comments", label: "评论管理" },
    { to: "/profile", label: "账号设置" },
];

async function doLogout(): Promise<void> {
    await logoutSession();
}
</script>

<template>
    <aside class="w-64 shrink-0 border-r border-slate-200 bg-white p-4">
        <div class="mb-4 rounded-2xl bg-steel-100 p-4">
            <div class="text-xs font-medium uppercase tracking-[0.18em] text-steel-700">会话状态</div>
            <div class="mt-2 text-sm font-semibold text-steel-900">{{ authText }}</div>
            <div class="mt-1 text-xs text-steel-700/80">自动刷新，最近检查于 {{ checkedText }}</div>
            <div class="mt-3 flex gap-2">
                <button
                    class="rounded-md bg-steel-600 px-2 py-1 text-xs font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="refreshSession">刷新</button>
                <button
                    class="rounded-md bg-slate-200 px-2 py-1 text-xs font-medium text-slate-700 hover:bg-slate-300 disabled:opacity-50"
                    :disabled="loading" @click="doLogout">退出</button>
            </div>
        </div>

        <nav class="space-y-1" data-tauri-drag-region="false">
            <RouterLink v-for="item in links" :key="item.to" :to="item.to"
                class="block rounded-xl px-3 py-2 text-sm text-slate-700 transition-colors hover:bg-slate-100">
                {{ item.label }}
            </RouterLink>
        </nav>

        <div class="mt-6 rounded-2xl border border-dashed border-slate-200 p-3 text-xs leading-5 text-slate-500">
            认证中心和接口实验页已经独立，不再出现在主导航中。
        </div>
    </aside>
</template>

<style scoped>
.router-link-active {
    @apply bg-steel-600 text-white hover:bg-steel-600;
}
</style>