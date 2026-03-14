<script setup lang="ts">
import { ref } from "vue";
import { useCookie, navigateTo, useRuntimeConfig } from "#app";

definePageMeta({
  layout: "blank",
});

const config = useRuntimeConfig();
const identifier = ref("");
const password = ref("");
const isLoading = ref(false);
const errorMsg = ref("");

const handleLogin = async (e: Event) => {
  e.preventDefault();
  isLoading.value = true;
  errorMsg.value = "";

  try {
    const data = await $fetch<{ access_token?: string; accessToken?: string }>(
      "/v1/auth/login",
      {
        baseURL: config.public.apiBase as string,
        method: "POST",
        body: {
          identifier: identifier.value,
          password: password.value,
        },
      }
    );

    // 保存 Token
    const authToken = useCookie("auth_token");
    authToken.value = data.access_token || data.accessToken || "";

    // 重定向到首页
    await navigateTo("/");
  } catch (err: any) {
    const msg =
      err?.data?.message || err?.message || "登录失败，请重试";
    errorMsg.value = msg;
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center px-6">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div
          class="w-16 h-16 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center text-white font-bold text-2xl mx-auto mb-4">
          H
        </div>
        <h1 class="text-3xl font-bold text-gray-900">Horizon</h1>
        <p class="text-gray-600 mt-2">登录你的账户</p>
      </div>

      <!-- Login form -->
      <form @submit="handleLogin" class="bg-white rounded-lg shadow-lg p-8 space-y-6">
        <!-- Error message -->
        <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700 text-sm">
          {{ errorMsg }}
        </div>

        <!-- 账号 -->
        <div>
          <label for="identifier" class="block text-sm font-semibold text-gray-900 mb-2">
            账号（用户名或邮箱）
          </label>
          <input id="identifier" v-model="identifier" type="text" required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="username 或 your@email.com" />
        </div>

        <!-- Password -->
        <div>
          <label for="password" class="block text-sm font-semibold text-gray-900 mb-2">
            密码
          </label>
          <input id="password" v-model="password" type="password" required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="••••••••" />
        </div>

        <!-- Submit -->
        <button type="submit" :disabled="isLoading"
          class="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed">
          {{ isLoading ? "登录中..." : "登录" }}
        </button>
      </form>

      <!-- Footer links -->
      <div class="mt-6 text-center space-y-2 text-sm text-gray-600">
        <p>
          没有账户?
          <NuxtLink to="/auth/register" class="text-blue-600 hover:underline">
            注册
          </NuxtLink>
        </p>
        <NuxtLink to="/" class="inline-block text-blue-600 hover:underline">
          返回首页
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
