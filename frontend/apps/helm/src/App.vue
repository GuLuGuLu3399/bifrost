<script setup lang="ts">
import AppTitleBar from "@/components/layout/AppTitleBar.vue";
import AppNavBar from "@/components/layout/AppNavBar.vue";
import { useSessionState } from "@/composables/useSessionState";
import { computed } from "vue";
import { useRoute } from "vue-router";
import { RouterView } from "vue-router";

const route = useRoute();
const { authenticated } = useSessionState();

const layout = computed(() => {
  const currentLayout = route.meta.layout;
  return typeof currentLayout === "string" ? currentLayout : "shell";
});

const shellVisible = computed(() => layout.value === "shell");
const authLayout = computed(() => layout.value === "auth");
</script>

<template>
  <div class="h-screen w-screen overflow-hidden bg-steel-50 text-slate-900">
    <template v-if="shellVisible">
      <AppTitleBar />
      <div class="flex h-[calc(100vh-38px)]">
        <AppNavBar />
        <main class="flex-1 overflow-auto p-6">
          <RouterView />
        </main>
      </div>
    </template>

    <template v-else-if="authLayout">
      <div
        class="flex min-h-screen items-center justify-center overflow-auto bg-[radial-gradient(circle_at_top,#dbeafe_0%,#eff6ff_38%,#f8fafc_100%)] px-6 py-10">
        <div
          class="w-full max-w-5xl rounded-[28px] border border-white/70 bg-white/85 p-4 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur">
          <div class="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
            <section class="rounded-[24px] bg-slate-950 px-8 py-10 text-white">
              <div
                class="inline-flex rounded-full border border-white/15 px-3 py-1 text-xs tracking-[0.24em] text-slate-300">
                HELM ACCESS
              </div>
              <h1 class="mt-5 text-4xl font-semibold tracking-tight">管理端认证中心</h1>
              <p class="mt-4 max-w-md text-sm leading-6 text-slate-300">
                登录、注册和会话状态已经从管理界面主导航分离，避免未登录场景与后台操作界面混在一起。
              </p>
              <div class="mt-8 grid gap-3 text-sm text-slate-200">
                <div class="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                  当前会话：{{ authenticated ? "已登录" : "未登录" }}
                </div>
                <div class="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                  独立认证页承载登录和注册，不再占用管理端导航栏。
                </div>
              </div>
            </section>

            <main class="flex items-center justify-center px-2 py-2">
              <RouterView />
            </main>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <main class="h-full overflow-auto p-6">
        <RouterView />
      </main>
    </template>
  </div>
</template>
