<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useHead, useSeoMeta, createError } from "#app";
import type { PostDetail } from "@bifrost/shared";
import { getImageUrl, formatDateDetailed, calculateReadingTime } from "~/utils";
import { usePostDetail } from "~/composables/usePost";

const route = useRoute();
const slug = route.params.slug as string;

// 获取文章详情
const { data: postResponse, error, pending } = await usePostDetail(slug);
const post = computed(() => postResponse.value?.post);

// 处理错误情况
if (error.value || !post.value) {
  throw createError({
    statusCode: 404,
    statusMessage: "Post Not Found",
  });
}

// SEO Meta 信息
useSeoMeta({
  title: post.value.title,
  description: post.value.summary,
  ogTitle: post.value.title,
  ogDescription: post.value.summary,
  ogImage: getImageUrl(post.value.cover_image_key),
  ogUrl: `https://example.com/posts/${slug}`, // 需要配置实际域名
  twitterCard: "summary_large_image",
});

// 设置页面标题
useHead({
  title: post.value.title,
});

// 计算阅读时间（粗略估算）
const readingTime = computed(() => {
  if (!post.value?.html_body) return 0;
  const words = post.value.html_body.replace(/<[^>]*>/g, "").length;
  return Math.max(1, Math.ceil(words / 300)); // 假设每分钟 300 字
});
</script>

<template>
  <main class="min-h-screen bg-white text-gray-900">
    <!-- Loading state -->
    <div v-if="pending" class="flex items-center justify-center h-screen">
      <p class="text-gray-500">加载中...</p>
    </div>

    <!-- Content -->
    <article v-else-if="post" class="max-w-3xl mx-auto px-6 py-12">
      <!-- Header -->
      <header class="mb-8 border-b border-gray-200 pb-8">
        <h1 class="text-4xl font-bold mb-4 text-gray-900">
          {{ post.title }}
        </h1>

        <p v-if="post.summary" class="text-xl text-gray-600 mb-6">
          {{ post.summary }}
        </p>

        <!-- Meta info -->
        <div class="flex flex-wrap items-center gap-4 text-sm text-gray-600">
          <!-- Author -->
          <div v-if="post.author" class="flex items-center gap-2">
            <img
              v-if="post.author.avatar_key"
              :src="getImageUrl(post.author.avatar_key)"
              :alt="post.author.nickname"
              class="w-8 h-8 rounded-full"
            />
            <span class="font-medium">{{ post.author.nickname }}</span>
          </div>

          <!-- Category -->
          <NuxtLink
            v-if="post.category"
            :to="`/?category=${post.category.id}`"
            class="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded transition-colors"
          >
            {{ post.category.name }}
          </NuxtLink>

          <!-- Date -->
          <span v-if="post.published_at">
            {{ formatDateDetailed(post.published_at) }}
          </span>

          <!-- Reading time -->
          <span>{{ readingTime }} 分钟阅读</span>

          <!-- Views -->
          <span v-if="post.view_count" class="ml-auto">
            👁️ {{ post.view_count }} 次浏览
          </span>
        </div>

        <!-- Tags -->
        <div
          v-if="post.tags && post.tags.length > 0"
          class="flex flex-wrap gap-2 mt-4"
        >
          <NuxtLink
            v-for="tag in post.tags"
            :key="tag.id"
            :to="`/?tag=${tag.id}`"
            class="inline-block px-2 py-1 text-xs bg-blue-50 text-blue-600 hover:bg-blue-100 rounded transition-colors"
          >
            #{{ tag.name }}
          </NuxtLink>
        </div>
      </header>

      <!-- Cover image -->
      <figure v-if="post.cover_image_key" class="mb-8">
        <img
          :src="getImageUrl(post.cover_image_key)"
          :alt="post.title"
          class="w-full rounded-lg"
        />
      </figure>

      <!-- Table of contents (if available) -->
      <aside
        v-if="post.toc_json"
        class="mb-8 p-4 bg-gray-50 rounded-lg border border-gray-200"
      >
        <h2 class="font-semibold text-gray-900 mb-2">目录</h2>
        <!-- TOC rendering logic would depend on toc_json structure -->
      </aside>

      <!-- Article content -->
      <div
        class="prose dark:prose-invert max-w-none mb-12"
        v-html="post.html_body"
      />

      <!-- Navigation -->
      <nav class="flex gap-4 pt-8 border-t border-gray-200">
        <!-- Previous post -->
        <NuxtLink
          v-if="post.prev_post_slug"
          :to="`/posts/${post.prev_post_slug}`"
          class="flex-1 px-4 py-3 bg-gray-100 hover:bg-gray-200 rounded transition-colors group"
        >
          <p class="text-xs text-gray-600 mb-1">← 上一篇</p>
          <p
            class="text-sm font-semibold text-gray-900 group-hover:text-blue-600 line-clamp-2"
          >
            {{ post.prev_post_title }}
          </p>
        </NuxtLink>

        <!-- Next post -->
        <NuxtLink
          v-if="post.next_post_slug"
          :to="`/posts/${post.next_post_slug}`"
          class="flex-1 px-4 py-3 bg-gray-100 hover:bg-gray-200 rounded transition-colors group text-right"
        >
          <p class="text-xs text-gray-600 mb-1">下一篇 →</p>
          <p
            class="text-sm font-semibold text-gray-900 group-hover:text-blue-600 line-clamp-2"
          >
            {{ post.next_post_title }}
          </p>
        </NuxtLink>
      </nav>
    </article>

    <!-- Error state -->
    <div v-else class="flex items-center justify-center h-screen">
      <div class="text-center">
        <p class="text-xl font-semibold text-gray-900 mb-2">文章未找到</p>
        <p class="text-gray-600 mb-6">抱歉，我们找不到你要找的文章。</p>
        <NuxtLink to="/" class="text-blue-600 hover:underline">
          返回首页
        </NuxtLink>
      </div>
    </div>
  </main>
</template>

<style scoped>
@import "tailwindcss" reference;

/* Prose styles can be customized here if needed */
:deep(.prose) {
  @apply text-gray-900;
}

:deep(.prose a) {
  @apply text-blue-600 hover:underline;
}

:deep(.prose code) {
  @apply px-2 py-1 bg-gray-100 rounded text-sm;
}

:deep(.prose pre) {
  @apply bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto;
}

:deep(.prose blockquote) {
  @apply border-l-4 border-blue-600 pl-4 italic text-gray-700;
}
</style>
