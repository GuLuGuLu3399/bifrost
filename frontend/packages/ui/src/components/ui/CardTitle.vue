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
  <h2
    :class="
      cn('text-2xl font-semibold leading-none tracking-tight', $props.class)
    "
  >
    <slot />
  </h2>
</template>
