<script setup lang="ts">
import { cn } from "../../lib/utils";

interface Props {
  class?: string;
}

withDefaults(defineProps<Props>(), {});

defineSlots<{
  default(props: {}): any;
}>();
</script>

<template>
  <div :class="cn('px-6 py-4', $props.class)">
    <slot />
  </div>
</template>
