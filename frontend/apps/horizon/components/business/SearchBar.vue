<script setup lang="ts">
import { Search } from "lucide-vue-next";
import { Input, Button } from "@bifrost/ui";
import { navigateTo } from "#app";
import { useSearch } from "~/composables/useSearch";

const { query, pending } = useSearch();

const handleSearch = () => {
  if (!query.value.trim()) return;
  navigateTo(`/search?q=${encodeURIComponent(query.value)}`);
};
</script>

<template>
  <div class="relative w-full max-w-sm">
    <div class="relative flex items-center gap-2">
      <!-- Search Icon -->
      <Search
        class="absolute left-3 h-4 w-4 text-muted-foreground pointer-events-none"
      />

      <!-- Input Field -->
      <Input
        v-model="query"
        placeholder="搜索文章、教程..."
        class="pl-10 pr-16"
        @keydown.enter="handleSearch"
      />

      <!-- Search Button -->
      <Button
        size="sm"
        class="absolute right-1 h-8"
        :disabled="pending || !query.trim()"
        @click="handleSearch"
      >
        {{ pending ? "..." : "搜索" }}
      </Button>
    </div>
  </div>
</template>
