<script setup lang="ts">
import type { CommentItem } from "@bifrost/shared";
import { getImageUrl, formatDateDetailed } from "~/utils";

defineProps<{
  comments: CommentItem[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  loadMore: [];
  replyTo: [commentId: string];
}>();
</script>

<template>
  <section class="space-y-4">
    <h3 class="text-lg font-semibold text-gray-900">评论</h3>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-4">
      <p class="text-gray-500 text-sm">加载中...</p>
    </div>

    <!-- Empty state -->
    <div
      v-else-if="!comments || comments.length === 0"
      class="text-center py-8"
    >
      <p class="text-gray-500 text-sm">暂无评论</p>
    </div>

    <!-- Comments list -->
    <div v-else class="space-y-4">
      <article
        v-for="comment in comments"
        :key="comment.id"
        class="border border-gray-200 rounded-lg p-4"
      >
        <!-- Comment header -->
        <div class="flex items-center gap-3 mb-3">
          <img
            v-if="comment.author?.avatar_key"
            :src="getImageUrl(comment.author.avatar_key)"
            :alt="comment.author.nickname"
            class="w-8 h-8 rounded-full"
          />
          <div class="flex-grow">
            <p class="font-semibold text-gray-900 text-sm">
              {{ comment.author?.nickname || "匿名" }}
            </p>
            <p v-if="comment.created_at" class="text-xs text-gray-500">
              {{ formatDateDetailed(comment.created_at) }}
            </p>
          </div>
          <button
            @click="emit('replyTo', comment.id)"
            class="text-xs text-blue-600 hover:underline"
          >
            回复
          </button>
        </div>

        <!-- Comment content -->
        <p class="text-gray-800 text-sm leading-relaxed">
          {{ comment.content }}
        </p>

        <!-- Reply count -->
        <div v-if="comment.reply_count && comment.reply_count > 0" class="mt-3">
          <button
            @click="emit('loadMore')"
            class="text-xs text-blue-600 hover:underline"
          >
            查看 {{ comment.reply_count }} 条回复
          </button>
        </div>
      </article>
    </div>
  </section>
</template>
