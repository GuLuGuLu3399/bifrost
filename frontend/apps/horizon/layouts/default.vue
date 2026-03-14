<template>
  <div class="min-h-screen flex flex-col bg-white">
    <!-- Header -->
    <LayoutHeader />

    <!-- Main content -->
    <main class="flex-grow">
      <slot />
    </main>

    <!-- Footer -->
    <LayoutFooter />
  </div>
</template>
