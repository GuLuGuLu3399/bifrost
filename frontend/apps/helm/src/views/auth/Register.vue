<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { registerUser } from "@/services/helmApi";

const username = ref("");
const email = ref("");
const password = ref("");
const loading = ref(false);
const result = ref("等待提交");
const router = useRouter();
const isDev = import.meta.env.DEV;

async function submit(): Promise<void> {
    loading.value = true;
    try {
        await registerUser({
            username: username.value,
            email: email.value,
            password: password.value,
        });
        result.value = "注册成功，正在跳转到登录页";
        window.setTimeout(() => {
            void router.replace("/auth/login");
        }, 600);
    } catch (error) {
        result.value = `注册失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <section class="w-full max-w-xl space-y-5 rounded-[24px] border border-slate-200 bg-white p-6 shadow-sm">
        <header>
            <div class="flex gap-2 rounded-full bg-slate-100 p-1 text-sm">
                <RouterLink to="/auth/login" class="rounded-full px-4 py-2 font-medium text-slate-600">登录</RouterLink>
                <RouterLink to="/auth/register" class="rounded-full bg-slate-900 px-4 py-2 font-medium text-white">注册
                </RouterLink>
            </div>
            <h2 class="mt-5 text-2xl font-semibold text-slate-900">创建新的管理员账号</h2>
            <p class="mt-2 text-sm text-slate-600">注册仍然保留在独立认证中心，不再占用后台导航入口。</p>
        </header>

        <div class="grid gap-4">
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">用户名</span>
                <input v-model="username" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="text" placeholder="superadmin" />
            </label>
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">邮箱</span>
                <input v-model="email" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm" type="email"
                    placeholder="admin@example.com" />
            </label>
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">密码</span>
                <input v-model="password" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="password" placeholder="请设置密码" />
            </label>
            <div class="flex justify-end">
                <button
                    class="rounded-xl bg-steel-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="submit">注册</button>
            </div>
        </div>

        <pre v-if="isDev"
            class="max-h-72 overflow-auto rounded-2xl bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
        <p v-else class="text-sm text-slate-500">{{ result }}</p>
    </section>
</template>
