<template>
  <NCard class="proxy-card">
    <ProxyView
      :proxies="proxies"
      proxy-type="http"
      @refresh="fetchData"
    />
  </NCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { BaseProxy } from '../utils/proxy'

const proxies = ref<BaseProxy[]>([])

const fetchData = () => {
  fetch('../api/proxies/http', { credentials: 'include' })
    .then((res) => res.json())
    .then((json) => {
      proxies.value = json.proxies
    })
}

fetchData()
</script>

<style lang="scss" scoped>
@use '../assets/styles/proxy-type.scss';
</style>
