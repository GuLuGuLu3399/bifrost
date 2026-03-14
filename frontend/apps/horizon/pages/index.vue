<script setup lang="ts">
import { usePostList } from "~/composables/usePost";

// 获取文章列表
const { data, error, pending } = usePostList(20);
</script>

<template>
  <main class="min-h-screen bg-white text-gray-900">
    <div class="max-w-4xl mx-auto px-6 py-12">
      <!-- Header -->
      <section class="mb-12">
        <p
          class="text-xs font-semibold uppercase tracking-[0.28em] text-gray-400 mb-2"
        >
          Bifrost Horizon
        </p>
        <h1 class="text-4xl font-bold mb-4">文章列表</h1>
        <p class="text-gray-600">高性能、SEO 友好的内容展现层</p>
      </section>

      <!-- Loading state -->
      <div v-if="pending" class="text-center py-8">
        <p class="text-gray-500">加载中...</p>
      </div>

      <!-- Error state -->
      <div
        v-else-if="error"
        class="bg-red-50 border border-red-200 rounded-lg p-6 text-red-700"
      >
        <p class="font-semibold mb-2">加载失败</p>
        <p class="text-sm">{{ error.message }}</p>
        <p class="text-sm mt-2">请检查后端服务是否正常运行</p>
      </div>

      <!-- Empty state -->
      <div
        v-else-if="!data?.posts || data.posts.length === 0"
        class="text-center py-12"
      >
        <p class="text-gray-500 mb-4">暂无文章</p>
        <p class="text-sm text-gray-400">请稍后再来查看</p>
      </div>

      <!-- Posts list -->
      <div v-else class="space-y-6">
        <BusinessPostCard
          v-for="post in data.posts"
          :key="post.id"
          :post="post"
        />
      </div>

      <!-- Pagination info -->
      <div
        v-if="data?.page?.total_count"
        class="mt-8 text-center text-gray-600"
      >
        <p class="text-sm">共 {{ data.page.total_count }} 篇文章</p>
      </div>
    </div>
  </main>
</template>
