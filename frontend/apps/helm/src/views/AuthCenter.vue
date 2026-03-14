<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useSessionState } from "@/composables/useSessionState";
import { loginWithIdentifier, registerUser } from "@/services/helmApi";

type AuthMode = "login" | "register";

const route = useRoute();
const router = useRouter();
const { markAuthenticated } = useSessionState();
const isDev = import.meta.env.DEV;

const loginIdentifier = ref("superadmin");
const loginPassword = ref("");
const registerUsername = ref("");
const registerEmail = ref("");
const registerPassword = ref("");
const loading = ref(false);
const result = ref("等待提交");

const mode = computed<AuthMode>(() =>
    route.query.mode === "register" ? "register" : "login",
);

const redirectPath = computed(() =>
    typeof route.query.redirect === "string" && route.query.redirect.trim() !== ""
        ? route.query.redirect
        : "/",
);

async function switchMode(nextMode: AuthMode): Promise<void> {
    await router.replace({
        path: "/auth",
        query: {
            ...route.query,
            mode: nextMode,
        },
    });
}

async function submitLogin(): Promise<void> {
    loading.value = true;
    try {
        const session = await loginWithIdentifier(
            loginIdentifier.value,
            loginPassword.value,
        );
        markAuthenticated(true);
        result.value = JSON.stringify(session, null, 2);
        await router.replace(redirectPath.value);
    } catch (error) {
        result.value = `登录失败: ${String(error)}`;
    } finally {
        loading.value = false;
    }
}

async function submitRegister(): Promise<void> {
    loading.value = true;
    try {
        await registerUser({
            username: registerUsername.value,
            email: registerEmail.value,
            password: registerPassword.value,
        });
        loginIdentifier.value = registerEmail.value || registerUsername.value;
        loginPassword.value = registerPassword.value;
        result.value = "注册成功，已切换到登录模式";
        await switchMode("login");
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
                <button class="rounded-full px-4 py-2 font-medium"
                    :class="mode === 'login' ? 'bg-slate-900 text-white' : 'text-slate-600'" :disabled="loading"
                    @click="switchMode('login')">登录</button>
                <button class="rounded-full px-4 py-2 font-medium"
                    :class="mode === 'register' ? 'bg-slate-900 text-white' : 'text-slate-600'" :disabled="loading"
                    @click="switchMode('register')">注册</button>
            </div>
            <h2 class="mt-5 text-2xl font-semibold text-slate-900">
                {{ mode === "login" ? "登录 Helm 管理端" : "创建新的管理员账号" }}
            </h2>
            <p class="mt-2 text-sm text-slate-600">
                认证中心已经合并为单页切换模式，未登录时会统一回到这里。
            </p>
        </header>

        <div v-if="mode === 'login'" class="grid gap-4">
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">账号或邮箱</span>
                <input v-model="loginIdentifier" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="text" placeholder="superadmin 或 admin@example.com" />
            </label>
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">密码</span>
                <input v-model="loginPassword" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="password" placeholder="请输入密码" />
            </label>
            <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-slate-500">登录后跳转到：{{ redirectPath }}</div>
                <button
                    class="rounded-xl bg-steel-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="submitLogin">登录</button>
            </div>
        </div>

        <div v-else class="grid gap-4">
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">用户名</span>
                <input v-model="registerUsername" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="text" placeholder="superadmin" />
            </label>
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">邮箱</span>
                <input v-model="registerEmail" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="email" placeholder="admin@example.com" />
            </label>
            <label>
                <span class="mb-1 block text-xs font-medium text-slate-600">密码</span>
                <input v-model="registerPassword" class="w-full rounded-xl border border-slate-300 px-3 py-3 text-sm"
                    type="password" placeholder="请设置密码" />
            </label>
            <div class="flex justify-end">
                <button
                    class="rounded-xl bg-steel-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="submitRegister">注册</button>
            </div>
        </div>

        <pre v-if="isDev"
            class="max-h-72 overflow-auto rounded-2xl bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
        <p v-else class="text-sm text-slate-500">{{ result }}</p>
    </section>
</template>