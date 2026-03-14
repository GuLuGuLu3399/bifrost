<script setup lang="ts">
import { ref } from "vue";
import { fetchUserProfile, updateUserProfile } from "@/services/helmApi";

const loading = ref(false);
const result = ref("等待查询");
const nickname = ref("");
const bio = ref("");
const toast = ref<{ type: "success" | "error"; text: string } | null>(null);
let toastTimer: ReturnType<typeof setTimeout> | null = null;

function showToast(type: "success" | "error", text: string): void {
    toast.value = { type, text };
    if (toastTimer) {
        clearTimeout(toastTimer);
    }
    toastTimer = setTimeout(() => {
        toast.value = null;
    }, 2200);
}

function applyProfileToForm(data: unknown): void {
    const obj = (data as Record<string, unknown>) ?? {};
    const directNickname = typeof obj.nickname === "string" ? obj.nickname : "";
    const directBio = typeof obj.bio === "string" ? obj.bio : "";

    if (directNickname || directBio) {
        nickname.value = directNickname;
        bio.value = directBio;
        return;
    }

    const profile = obj.profile as Record<string, unknown> | undefined;
    if (profile) {
        nickname.value = typeof profile.nickname === "string" ? profile.nickname : "";
        bio.value = typeof profile.bio === "string" ? profile.bio : "";
    }
}

async function queryProfile(): Promise<void> {
    loading.value = true;
    try {
        const data = await fetchUserProfile();
        applyProfileToForm(data);
        result.value = JSON.stringify(data, null, 2);
        showToast("success", "资料加载成功");
    } catch (error) {
        result.value = `查询失败: ${String(error)}`;
        showToast("error", "资料加载失败");
    } finally {
        loading.value = false;
    }
}

async function saveProfile(): Promise<void> {
    loading.value = true;
    try {
        const data = await updateUserProfile({
            nickname: nickname.value,
            bio: bio.value,
        });
        result.value = JSON.stringify(data, null, 2);
        showToast("success", "资料保存成功");
    } catch (error) {
        result.value = `更新失败: ${String(error)}`;
        showToast("error", "资料保存失败");
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <section class="mx-auto max-w-5xl space-y-4">
        <div v-if="toast" :class="[
            'fixed right-6 top-16 z-50 rounded-lg px-3 py-2 text-sm shadow-panel',
            toast.type === 'success' ? 'bg-emerald-600 text-white' : 'bg-rose-600 text-white',
        ]">
            {{ toast.text }}
        </div>

        <header class="panel p-5">
            <h2 class="text-xl font-semibold text-slate-900">账号设置</h2>
            <p class="mt-2 text-sm text-slate-600">查看并更新当前用户资料（nickname/bio）。</p>
        </header>

        <section class="panel p-5">
            <div class="grid gap-3 md:grid-cols-2">
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">nickname</span>
                    <input v-model="nickname" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                        type="text" placeholder="nickname" />
                </label>
                <label>
                    <span class="mb-1 block text-xs font-medium text-slate-600">bio</span>
                    <input v-model="bio" class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" type="text"
                        placeholder="bio" />
                </label>
            </div>

            <div class="mt-4 flex gap-2">
                <button
                    class="rounded-lg bg-steel-600 px-4 py-2 text-sm font-medium text-white hover:bg-steel-700 disabled:opacity-50"
                    :disabled="loading" @click="queryProfile">读取资料</button>
                <button
                    class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
                    :disabled="loading" @click="saveProfile">保存资料</button>
            </div>
        </section>

        <section class="panel p-5">
            <h3 class="text-sm font-semibold text-slate-700">响应结果</h3>
            <pre
                class="mt-3 max-h-[420px] overflow-auto rounded-lg bg-slate-950 p-4 text-xs text-emerald-200">{{ result }}</pre>
        </section>
    </section>
</template>
