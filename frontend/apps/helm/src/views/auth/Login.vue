<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useSessionState } from "@/composables/useSessionState";
import { loginWithIdentifier } from "@/services/helmApi";

const identifier = ref("superadmin");
const password = ref("");
const loading = ref(false);
const result = ref("等待提交");
const router = useRouter();
const route = useRoute();
const { markAuthenticated } = useSessionState();
const isDev = import.meta.env.DEV;

const redirectPath = computed(() =>
    typeof route.query.redirect === "string" && route.query.redirect.trim() !== ""
        ? route.query.redirect
        : "/",
);

async function submit(): Promise<void> {
    loading.value = true;
    try {
        const session = await loginWithIdentifier(identifier.value, password.value);
        markAuthenticated(true);
        result.value = JSON.stringify(session, null, 2);
        await router.replace(redirectPath.value);
    } catch (error) {
        result.value = `登录失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <section class="w-full max-w-xl space-y-5 rounded-[24px] border border-slate-200 bg-white p-6 shadow-sm">
        <header>
            <div class="flex gap-2 rounded-full bg-slate-100 p-1 text-sm">
                <RouterLink to="/auth/login" class="rounded-full bg-slate-900 px-4 py-2 font-medium text-white">登录
                </RouterLink>
                <RouterLink to="/auth/register" class="rounded-full px-4 py-2 font-medium text-slate-600">注册
                </RouterLink>
            </div>
            <h2 class="mt-5 text-2xl font-semibold text-slate-900">登录 Helm 管理端</h2>
            <p class="mt-2 text-sm text-slate-600">认证页已经独立于后台壳体，登录成功后会自动返回原目标页面。</p>
        </header>

        <div class="grid gap-4">
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">账号或邮箱</span>
                <input v-model="identifier" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="text" placeholder="superadmin 或 admin@example.com" />
            </label>
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">密码</span>
                <input v-model="password" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="password" placeholder="请输入密码" />
            </label>
            <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-slate-500">登录后跳转到：{{ redirectPath }}</div>
                <button
                    class="rounded-xl bg-steel-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="submit">登录</button>
            </div>
        </div>

        <pre v-if="isDev"
            class="max-h-72 overflow-auto rounded-2xl bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
    </section>
</template>
