<script setup lang="ts">
import AppTitleBar from "./components/layout/AppTitleBar.vue";
import ImageUploader from "./components/ImageUploader.vue";
import { RouterView, RouterLink } from "vue-router";
</script>

<template>
  <div class="flex flex-col h-screen w-screen bg-background">
    <AppTitleBar />

    <div class="flex flex-1 overflow-hidden">
      <aside class="sidebar border-r p-4 hidden md-block">
        <nav class="flex flex-col gap-2">
          <div
            class="text-xs font-mono text-muted-foreground"
            style="margin-bottom: 8px"
          >
            NAVIGATOR
          </div>
          <RouterLink to="/" class="titlebar-btn rounded-sm p-2 hover-bg-muted"
            >DASHBOARD</RouterLink
          >
          <RouterLink
            to="/auth/register"
            class="titlebar-btn rounded-sm p-2 hover-bg-muted"
            >REGISTER</RouterLink
          >
          <RouterLink
            to="/auth/login"
            class="titlebar-btn rounded-sm p-2 hover-bg-muted"
            >LOGIN</RouterLink
          >
        </nav>
      </aside>

      <main class="flex-1 overflow-y-auto p-6 relative">
        <RouterView />
        <section style="margin-top: 24px">
          <ImageUploader />
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* Page-specific minimal tweaks can go here if needed */
</style>
