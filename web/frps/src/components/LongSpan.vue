<template>
  <div class="long-span">
    <span class="content" :style="{ maxWidth: length + 'ch' }">{{ content }}</span>
    <n-tooltip trigger="hover">
      <template #trigger>
        <n-button text style="padding: 0">
          <n-icon>
            <OptionsOutline />
          </n-icon>
        </n-button>
      </template>
      {{ content }}
    </n-tooltip>
  </div>
</template>

<script setup lang="ts">
import { OptionsOutline } from '@vicons/ionicons5'

defineProps<{
  content: string
  length: number
}>()
</script>

<style lang="scss" scoped>
@use '../assets/styles/long-span.scss';
</style>
