<template>
  <div class="expand-container">
    <NCard size="small" :bordered="false" class="expand-card">
      <NGrid :cols="2" :x-gap="24">
        <NGridItem>
          <div class="section-title">基本信息</div>
          <NDescriptions label-placement="left" :column="1" class="description-list">
            <NDescriptionsItem label="隧道名称">
              {{ row.name }}
            </NDescriptionsItem>
            <NDescriptionsItem label="隧道类型">
              {{ row.type }}
            </NDescriptionsItem>
            <NDescriptionsItem label="数据加密">
              {{ row.encryption }}
            </NDescriptionsItem>
            <NDescriptionsItem label="数据压缩">
              {{ row.compression }}
            </NDescriptionsItem>
            <NDescriptionsItem label="最后启动时间">
              {{ row.lastStartTime }}
            </NDescriptionsItem>
            <NDescriptionsItem label="最后关闭时间">
              {{ row.lastCloseTime }}
            </NDescriptionsItem>
          </NDescriptions>
        </NGridItem>

        <NGridItem>
          <div class="section-title">连接信息</div>
          <NDescriptions label-placement="left" :column="1" class="description-list">
            <template v-if="proxyType === 'http' || proxyType === 'https'">
              <NDescriptionsItem label="Run ID">
                {{ row.runId }}
              </NDescriptionsItem>
              <NDescriptionsItem label="域名">
                {{ row.customDomains }}
              </NDescriptionsItem>
              <NDescriptionsItem label="子域名">
                {{ row.subdomain }}
              </NDescriptionsItem>
              <NDescriptionsItem label="路径">
                {{ row.locations }}
              </NDescriptionsItem>
              <NDescriptionsItem label="主机重写">
                {{ row.hostHeaderRewrite }}
              </NDescriptionsItem>
            </template>

            <template v-else-if="proxyType === 'tcpmux'">
              <NDescriptionsItem label="Run ID">
                {{ row.runId }}
              </NDescriptionsItem>
              <NDescriptionsItem label="复用器">
                {{ row.multiplexer }}
              </NDescriptionsItem>
              <NDescriptionsItem label="HTTP用户路由">
                {{ row.routeByHTTPUser }}
              </NDescriptionsItem>
              <NDescriptionsItem label="域名">
                {{ row.customDomains }}
              </NDescriptionsItem>
              <NDescriptionsItem label="子域名">
                {{ row.subdomain }}
              </NDescriptionsItem>
            </template>
            <template v-else>
              <NDescriptionsItem label="Run ID">
                {{ row.runId }}
              </NDescriptionsItem>
              <NDescriptionsItem label="地址">
                {{ row.addr }}
              </NDescriptionsItem>
            </template>
          </NDescriptions>
        </NGridItem>
      </NGrid>

      <template v-if="row.annotations && row.annotations.size > 0">
        <NDivider style="margin: 16px 0" />
        <div class="section-title">注解信息</div>
        <NList bordered class="annotation-list">
          <NListItem v-for="item in annotationsArray()" :key="item.key">
            <NSpace justify="space-between" style="width: 100%">
              <NTag type="info" size="small" round>{{ item.key }}</NTag>
              <NText>{{ item.value }}</NText>
            </NSpace>
          </NListItem>
        </NList>
      </template>
    </NCard>
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
