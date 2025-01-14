<template>
  <n-card>
    <n-message-provider>
      <ProxyView :proxies="proxies" proxyType="tcp" @refresh="fetchData"/>
    </n-message-provider>
  </n-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { TCPProxy } from '../utils/proxy.js'
import ProxyView from './ProxyView.vue'

let proxies = ref<TCPProxy[]>([])

const fetchData = () => {
  fetch('../api/proxy/tcp', { credentials: 'include' })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      proxies.value = []
      for (let proxyStats of json.proxies) {
        proxies.value.push(new TCPProxy(proxyStats))
      }
    })
}
fetchData()
</script>

<style></style>
