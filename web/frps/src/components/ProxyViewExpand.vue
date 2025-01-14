<template>
  <div>
    <n-grid :cols="2" :x-gap="12">
      <n-grid-item>
        <n-descriptions label-placement="left" :column="1">
          <n-descriptions-item label="隧道名称">
            {{ row.name }}
          </n-descriptions-item>
          <n-descriptions-item label="隧道类型">
            {{ row.type }}
          </n-descriptions-item>
          <n-descriptions-item label="加密">
            {{ row.encryption }}
          </n-descriptions-item>
          <n-descriptions-item label="压缩">
            {{ row.compression }}
          </n-descriptions-item>
          <n-descriptions-item label="最后启动时间">
            {{ row.lastStartTime }}
          </n-descriptions-item>
          <n-descriptions-item label="最后关闭时间">
            {{ row.lastCloseTime }}
          </n-descriptions-item>
        </n-descriptions>
      </n-grid-item>

      <n-grid-item>
        <n-descriptions label-placement="left" :column="1">
          <template v-if="proxyType === 'http' || proxyType === 'https'">
            <n-descriptions-item label="域名">
              {{ row.customDomains }}
            </n-descriptions-item>
            <n-descriptions-item label="子域名">
              {{ row.subdomain }}
            </n-descriptions-item>
            <n-descriptions-item label="路径">
              {{ row.locations }}
            </n-descriptions-item>
            <n-descriptions-item label="主机重写">
              {{ row.hostHeaderRewrite }}
            </n-descriptions-item>
          </template>

          <template v-else-if="proxyType === 'tcpmux'">
            <n-descriptions-item label="复用器">
              {{ row.multiplexer }}
            </n-descriptions-item>
            <n-descriptions-item label="HTTP用户路由">
              {{ row.routeByHTTPUser }}
            </n-descriptions-item>
            <n-descriptions-item label="域名">
              {{ row.customDomains }}
            </n-descriptions-item>
            <n-descriptions-item label="子域名">
              {{ row.subdomain }}
            </n-descriptions-item>
          </template>

          <template v-else>
            <n-descriptions-item label="地址">
              {{ row.addr }}
            </n-descriptions-item>
          </template>
        </n-descriptions>
      </n-grid-item>
    </n-grid>

    <template v-if="row.annotations && row.annotations.size > 0">
      <n-divider />
      <n-text depth="3" style="font-size: 16px">注解</n-text>
      <n-list>
        <n-list-item v-for="item in annotationsArray()" :key="item.key">
          <n-space justify="space-between" style="width: 100%">
            <n-text class="annotation-key">{{ item.key }}</n-text>
            <n-text>{{ item.value }}</n-text>
          </n-space>
        </n-list-item>
      </n-list>
    </template>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  row: any
  proxyType: string
}>()

// annotationsArray returns an array of key-value pairs from the annotations map.
const annotationsArray = (): Array<{ key: string; value: string }> => {
  const array: Array<{ key: string; value: any }> = []
  if (props.row.annotations) {
    props.row.annotations.forEach((value: any, key: string) => {
      array.push({ key, value })
    })
  }
  return array
}
</script>

<style lang="scss" scoped>
@use '../assets/styles/proxy-view-expand.scss';
</style>
