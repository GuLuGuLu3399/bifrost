<script setup lang="ts"></script>

<template>
  <footer class="bg-gray-100 border-t border-gray-200 mt-16">
    <div class="max-w-6xl mx-auto px-6 py-12">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-8 mb-8">
        <!-- About -->
        <div>
          <h3 class="font-bold text-gray-900 mb-4">Bifrost Horizon</h3>
          <p class="text-gray-600 text-sm">
            高性能、SEO 友好的内容展现层，为你呈现精品内容。
          </p>
        </div>

        <!-- Links -->
        <div>
          <h3 class="font-bold text-gray-900 mb-4">导航</h3>
          <ul class="space-y-2 text-sm text-gray-600">
            <li>
              <NuxtLink to="/" class="hover:text-gray-900">首页</NuxtLink>
            </li>
            <li>
              <NuxtLink to="/search" class="hover:text-gray-900">搜索</NuxtLink>
            </li>
          </ul>
        </div>

        <!-- Contact -->
        <div>
          <h3 class="font-bold text-gray-900 mb-4">联系我们</h3>
          <p class="text-gray-600 text-sm">邮箱: contact@example.com</p>
        </div>
      </div>

      <div
        class="border-t border-gray-300 pt-8 text-center text-sm text-gray-600"
      >
        <p>&copy; 2025 Bifrost Horizon. All rights reserved.</p>
      </div>
    </div>
  </footer>
</template>
