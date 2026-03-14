<script setup lang="ts">
import { navigateTo } from "#app";

const handleSearch = (q: string) => {
  if (q.trim()) {
    navigateTo(`/search?q=${encodeURIComponent(q)}`);
  }
};
</script>

<template>
  <header class="sticky top-0 bg-white border-b border-gray-200 z-40">
    <nav class="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
      <!-- Logo -->
      <NuxtLink to="/" class="flex items-center gap-2 hover:no-underline">
        <div
          class="w-8 h-8 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center text-white font-bold"
        >
          H
        </div>
        <span class="font-bold text-lg text-gray-900">Horizon</span>
      </NuxtLink>

      <!-- Search bar -->
      <div class="flex-1 mx-8">
        <BusinessSearchBar @search="handleSearch" />
      </div>

      <!-- Navigation links -->
      <div class="flex items-center gap-6">
        <NuxtLink
          to="/"
          class="text-gray-600 hover:text-gray-900 transition-colors"
        >
          首页
        </NuxtLink>
        <NuxtLink
          to="/search"
          class="text-gray-600 hover:text-gray-900 transition-colors"
        >
          搜索
        </NuxtLink>
        <NuxtLink
          to="/auth/login"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          登录
        </NuxtLink>
      </div>
    </nav>
  </header>
</template>
