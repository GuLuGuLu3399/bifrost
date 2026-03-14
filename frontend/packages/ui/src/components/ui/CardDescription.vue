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
  <div :class="cn('text-sm text-muted-foreground', $props.class)">
    <slot />
  </div>
</template>
