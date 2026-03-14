<script setup lang="ts">
import type { PostSummary } from "@bifrost/shared";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from "@bifrost/ui";
import { getImageUrl, formatDate } from "~/utils";

defineProps<{
  post: PostSummary;
}>();
</script>

<template>
  <NuxtLink :to="`/posts/${post.slug}`" class="block group">
    <Card
      class="h-full overflow-hidden transition-all hover:shadow-lg hover:border-primary/50"
    >
      <!-- Thumbnail -->
      <div
        v-if="post.cover_image_key"
        class="aspect-video w-full overflow-hidden bg-muted"
      >
        <img
          :src="getImageUrl(post.cover_image_key)"
          :alt="post.title"
          class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
        />
      </div>

      <CardHeader>
        <!-- Meta info -->
        <div class="flex items-center gap-2 text-xs text-muted-foreground mb-2">
          <span v-if="post.category" class="font-medium text-primary">
            {{ post.category.name }}
          </span>
          <span v-if="post.published_at">
            {{ formatDate(post.published_at) }}
          </span>
        </div>

        <!-- Title -->
        <CardTitle
          class="line-clamp-2 group-hover:text-primary transition-colors"
        >
          {{ post.title }}
        </CardTitle>
      </CardHeader>

      <!-- Summary -->
      <CardContent>
        <p
          v-if="post.summary"
          class="text-sm text-muted-foreground line-clamp-3"
        >
          {{ post.summary }}
        </p>
      </CardContent>

      <!-- Footer -->
      <CardFooter
        v-if="post.author || post.view_count"
        class="border-t bg-muted/20 flex justify-between"
      >
        <span v-if="post.author" class="text-xs font-medium">
          {{ post.author.nickname }}
        </span>
        <span v-if="post.view_count" class="text-xs flex items-center gap-1">
          👁️ {{ post.view_count }}
        </span>
      </CardFooter>
    </Card>
  </NuxtLink>
</template>
