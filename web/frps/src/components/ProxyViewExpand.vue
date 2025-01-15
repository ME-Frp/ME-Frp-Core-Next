<template>
  <div class="expand-container">
    <n-card size="small" :bordered="false" class="expand-card">
      <n-grid :cols="2" :x-gap="24">
        <n-grid-item>
          <div class="section-title">基本信息</div>
          <n-descriptions label-placement="left" :column="1" class="description-list">
            <n-descriptions-item label="隧道名称">
              {{ row.name }}
            </n-descriptions-item>
            <n-descriptions-item label="隧道类型">
              {{ row.type }}
            </n-descriptions-item>
            <n-descriptions-item label="数据加密">
              {{ row.encryption }}
            </n-descriptions-item>
            <n-descriptions-item label="数据压缩">
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
          <div class="section-title">连接信息</div>
          <n-descriptions label-placement="left" :column="1" class="description-list">
            <template v-if="proxyType === 'http' || proxyType === 'https'">
              <n-descriptions-item label="Run ID">
                {{ row.runId }}
              </n-descriptions-item>
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
              <n-descriptions-item label="Run ID">
                {{ row.runId }}
              </n-descriptions-item>
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
              <n-descriptions-item label="Run ID">
                {{ row.runId }}
              </n-descriptions-item>
              <n-descriptions-item label="地址">
                {{ row.addr }}
              </n-descriptions-item>
            </template>
          </n-descriptions>
        </n-grid-item>
      </n-grid>

      <template v-if="row.annotations && row.annotations.size > 0">
        <n-divider style="margin: 16px 0" />
        <div class="section-title">注解信息</div>
        <n-list bordered class="annotation-list">
          <n-list-item v-for="item in annotationsArray()" :key="item.key">
            <n-space justify="space-between" style="width: 100%">
              <n-tag type="info" size="small" round>{{ item.key }}</n-tag>
              <n-text>{{ item.value }}</n-text>
            </n-space>
          </n-list-item>
        </n-list>
      </template>
    </n-card>
  </div>
</template>

<style lang="scss" scoped>
.expand-container {
  padding: 8px 0;
}

.expand-card {
  background-color: var(--n-color-modal);
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--n-text-color-2);
  margin-bottom: 12px;
}

.description-list {
  :deep(.n-descriptions-item__label) {
    width: 100px;
    color: var(--n-text-color-3);
  }

  :deep(.n-descriptions-item__content) {
    color: var(--n-text-color);
  }
}

.annotation-list {
  margin-top: 8px;
  border-radius: 3px;
  
  :deep(.n-list-item) {
    padding: 8px 12px;
  }
}
</style>

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
