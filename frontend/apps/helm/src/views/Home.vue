<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useSessionState } from "@/composables/useSessionState";

const { authenticated, loading, refreshSession, logoutSession, lastCheckedAt } =
    useSessionState();

const status = computed(() => (authenticated.value ? "已登录" : "未登录"));

const checkedText = computed(() => {
    if (!lastCheckedAt.value) {
        return "等待首次检查";
    }

    return new Date(lastCheckedAt.value).toLocaleString();
});
</script>

<template>
    <section class="mx-auto max-w-6xl space-y-4">
        <header class="panel p-6">
            <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
                <div>
                    <div class="text-xs uppercase tracking-[0.24em] text-steel-700">Helm Admin</div>
                    <h2 class="mt-2 text-2xl font-semibold text-slate-900">管理端总览</h2>
                    <p class="mt-2 text-sm text-slate-600">认证中心与接口实验页已经独立，主界面只保留日常管理入口。</p>
                </div>
                <div class="rounded-2xl bg-slate-100 px-4 py-3 text-sm text-slate-600">
                    最近会话检查：{{ checkedText }}
                </div>
            </div>
        </header>

        <div class="grid gap-4 lg:grid-cols-[1.2fr_0.8fr_0.8fr]">
            <article class="panel p-5">
                <h3 class="text-sm font-semibold text-slate-700">会话状态</h3>
                <p class="mt-2 text-3xl font-bold text-steel-700">{{ status }}</p>
                <p class="mt-2 text-sm text-slate-500">状态会在窗口重新聚焦和固定轮询时自动刷新。</p>
                <div class="mt-4 flex gap-2">
                    <button
                        class="rounded-lg bg-steel-600 px-3 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                        :disabled="loading" @click="refreshSession">立即刷新</button>
                    <button
                        class="rounded-lg bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200 disabled:opacity-50"
                        :disabled="loading" @click="logoutSession">退出登录</button>
                </div>
            </article>

            <article class="panel p-5">
                <h3 class="text-sm font-semibold text-slate-700">内容系统</h3>
                <p class="mt-2 text-sm text-slate-600">文章创建、筛选、编辑、分类和评论现在都有正式页面入口。</p>
                <div class="mt-4 flex flex-wrap gap-2">
                    <RouterLink to="/posts"
                        class="inline-flex rounded-lg bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800">
                        文章管理</RouterLink>
                    <RouterLink to="/taxonomy"
                        class="inline-flex rounded-lg bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200">
                        分类标签</RouterLink>
                    <RouterLink to="/comments"
                        class="inline-flex rounded-lg bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200">
                        评论管理</RouterLink>
                </div>
            </article>

            <article class="panel p-5">
                <h3 class="text-sm font-semibold text-slate-700">独立入口</h3>
                <p class="mt-2 text-sm text-slate-600">认证中心和接口实验页不再进入侧边栏，按场景单独打开。</p>
                <div class="mt-4 flex flex-wrap gap-2">
                    <RouterLink to="/auth"
                        class="inline-flex rounded-lg bg-steel-100 px-3 py-2 text-sm font-medium text-steel-800 hover:bg-steel-200">
                        认证中心</RouterLink>
                    <RouterLink to="/lab/api"
                        class="inline-flex rounded-lg bg-slate-100 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-200">
                        接口实验页</RouterLink>
                </div>
            </article>
        </div>
    </section>
</template>
