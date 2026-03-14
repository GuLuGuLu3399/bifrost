<script setup lang="ts">
import { watch } from "vue";
import { useRoute, navigateTo } from "#app";
import { useSearch } from "~/composables/useSearch";
import { formatDate } from "~/utils";

const route = useRoute();
const {
  query,
  page,
  categoryId,
  tagId,
  searchResult,
  suggestions,
  showSuggestions,
  pending,
  selectSuggestion,
} = useSearch();

// 初始化搜索关键词
query.value = (route.query.q as string) || "";

// 监听路由查询参数变化
watch(
  () => route.query.q,
  (newQuery: any) => {
    if (newQuery) {
      query.value = newQuery as string;
    }
  }
);

// 提交搜索表单
const handleSearch = (e: Event) => {
  e.preventDefault();
  if (query.value.trim()) {
    navigateTo(`/search?q=${encodeURIComponent(query.value)}`);
  }
};
</script>

<template>
  <main class="min-h-screen bg-white text-gray-900">
    <div class="max-w-4xl mx-auto px-6 py-12">
      <!-- Search header -->
      <section class="mb-12">
        <h1 class="text-4xl font-bold mb-6">搜索文章</h1>

        <!-- Search form -->
        <form @submit="handleSearch" class="relative mb-8">
          <div class="flex gap-2">
            <div class="flex-grow relative">
              <input
                v-model="query"
                type="text"
                placeholder="输入关键词..."
                class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />

              <!-- Search suggestions -->
              <div
                v-if="showSuggestions"
                class="absolute top-full left-0 right-0 mt-2 bg-white border border-gray-300 rounded-lg shadow-lg z-10"
              >
                <button
                  v-for="suggestion in suggestions"
                  :key="suggestion"
                  type="button"
                  @click="selectSuggestion(suggestion)"
                  class="w-full text-left px-4 py-2 hover:bg-gray-100 transition-colors"
                >
                  {{ suggestion }}
                </button>
              </div>
            </div>

            <button
              type="submit"
              class="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold"
            >
              搜索
            </button>
          </div>
        </form>
      </section>

      <!-- Loading state -->
      <div v-if="pending" class="text-center py-8">
        <p class="text-gray-500">搜索中...</p>
      </div>

      <!-- No query -->
      <div v-else-if="!query.trim()" class="text-center py-12">
        <p class="text-gray-500">输入关键词开始搜索</p>
      </div>

      <!-- No results -->
      <div
        v-else-if="searchResult && searchResult.hits.length === 0"
        class="text-center py-12"
      >
        <p class="text-gray-500 mb-4">没有找到相关文章</p>
        <p class="text-sm text-gray-400">
          尝试其他关键词或
          <NuxtLink to="/" class="text-blue-600 hover:underline"
            >返回首页</NuxtLink
          >
        </p>
      </div>

      <!-- Results -->
      <div v-else-if="searchResult" class="space-y-6">
        <!-- Results count -->
        <p class="text-sm text-gray-600 mb-6">
          找到 {{ searchResult.total_hits }} 个结果（耗时
          {{ searchResult.took_ms }} ms）
        </p>

        <!-- Search hits -->
        <article
          v-for="hit in searchResult.hits"
          :key="hit.id"
          class="border border-gray-200 rounded-lg p-6 hover:shadow-lg transition-shadow"
        >
          <div class="mb-2 flex items-center gap-2">
            <span class="text-xs font-semibold text-gray-500">
              相关度: {{ (hit.score * 100).toFixed(0) }}%
            </span>
            <span v-if="hit.published_at" class="text-xs text-gray-400">
              · {{ formatDate(hit.published_at) }}
            </span>
          </div>

          <NuxtLink
            :to="`/posts/${hit.slug}`"
            class="block group mb-3 hover:no-underline"
          >
            <h2
              v-if="hit.highlight_title"
              class="text-xl font-semibold text-gray-900 group-hover:text-blue-600"
              v-html="hit.highlight_title"
            />
            <h2
              v-else
              class="text-xl font-semibold text-gray-900 group-hover:text-blue-600"
            >
              {{ hit.title }}
            </h2>
          </NuxtLink>

          <!-- Highlight content -->
          <div
            v-if="hit.highlight_content"
            class="text-gray-600 line-clamp-3 mb-3"
            v-html="hit.highlight_content"
          />
        </article>

        <!-- Facets (if available) -->
        <aside
          v-if="searchResult.facets"
          class="mt-12 p-6 bg-gray-50 rounded-lg border border-gray-200"
        >
          <h3 class="font-semibold text-gray-900 mb-4">筛选</h3>

          <!-- Categories -->
          <div v-if="searchResult.facets.categories" class="mb-6">
            <p class="text-sm font-medium text-gray-700 mb-2">分类</p>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="(count, categoryName) in searchResult.facets.categories"
                :key="categoryName"
                class="px-3 py-1 text-sm bg-white border border-gray-300 rounded hover:bg-blue-50 transition-colors"
              >
                {{ categoryName }} ({{ count }})
              </button>
            </div>
          </div>

          <!-- Tags -->
          <div v-if="searchResult.facets.tags">
            <p class="text-sm font-medium text-gray-700 mb-2">标签</p>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="(count, tagName) in searchResult.facets.tags"
                :key="tagName"
                class="px-3 py-1 text-sm bg-white border border-gray-300 rounded hover:bg-blue-50 transition-colors"
              >
                {{ tagName }} ({{ count }})
              </button>
            </div>
          </div>
        </aside>
      </div>
    </div>
  </main>
</template>
