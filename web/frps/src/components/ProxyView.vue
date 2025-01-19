<template>
  <div>
    <div class="proxy-header">
      <n-text class="title">
        {{ proxyType.toUpperCase() }} 隧道
      </n-text>
      <n-space>
        <n-button @click="$emit('refresh')">
          <template #icon>
            <n-icon><refresh-outline /></n-icon>
          </template>
          刷新
        </n-button>
      </n-space>
    </div>

    <n-data-table :columns="columns" :data="proxies" :pagination="pagination" :row-key="(row: any) => row.name"
      :max-height="null" size="small" class="proxy-table" />

    <n-modal v-model:show="dialogVisible" :title="dialogVisibleName" preset="card" style="width: 700px">
      <Traffic :proxyName="dialogVisibleName" />
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, ref } from 'vue'
import { NTag, NButton, NIcon, useMessage, type DataTableColumns } from 'naive-ui'
import { 
  RefreshOutline, 
  TrashOutline,
  StatsChartOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline
} from '@vicons/ionicons5'
import * as Humanize from 'humanize-plus'
import type { BaseProxy } from '../utils/proxy.js'
import ProxyViewExpand from './ProxyViewExpand.vue'

const props = defineProps<{
  proxies: BaseProxy[]
  proxyType: string
}>()

const emit = defineEmits(['refresh'])
const message = useMessage()

const dialogVisible = ref(false)
const dialogVisibleName = ref("")

const pagination = {
  pageSize: 10
}

const columns: DataTableColumns<BaseProxy> = [
  {
    type: 'expand',
    expandable: () => true,
    renderExpand: (row) => {
      return h(ProxyViewExpand, {
        row,
        proxyType: props.proxyType
      })
    }
  },
  {
    title: '名称',
    key: 'name',
    sorter: true
  },
  {
    title: '端口',
    key: 'port',
    sorter: true
  },
  {
    title: '连接数',
    key: 'conns',
    sorter: true
  },
  {
    title: '入网流量',
    key: 'trafficIn',
    sorter: true,
    render(row) {
      return Humanize.fileSize(row.trafficIn)
    }
  },
  {
    title: '出网流量',
    key: 'trafficOut',
    sorter: true,
    render(row) {
      return Humanize.fileSize(row.trafficOut)
    }
  },
  {
    title: '客户端版本',
    key: 'clientVersion',
    sorter: true
  },
  {
    title: '状态',
    key: 'status',
    sorter: true,
    render(row) {
      return h(
        NTag,
        {
          type: row.status === 'online' ? 'success' : 'error',
        },
        {
          default: () => [
            h(NIcon, null, {
              default: () => h(row.status === 'online' ? CheckmarkCircleOutline : CloseCircleOutline)
            }),
            ' ',
            row.status === 'online' ? '在线' : '离线'
          ]
        }
      )
    }
  },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h('div', [
        h(
          NButton,
          {
            type: 'primary',
            style: 'margin-right: 5px',
            onClick: () => {
              dialogVisibleName.value = row.name
              dialogVisible.value = true
            }
          },
          {
            default: () => [
              h(NIcon, null, { default: () => h(StatsChartOutline) }),
              ' 流量'
            ]
          }
        ),
        h(
          NButton,
          {
            type: 'error',
            disabled: row.status !== 'online',
            onClick: () => {
              kickProxy(row)
            }
          },
          {
            default: () => [
              h(NIcon, null, { default: () => h(CloseCircleOutline) }),
              '强制下线'
            ]
          }
        )
      ])
    }
  }
]

const kickProxy = (proxy: BaseProxy) => {
  fetch(`../api/client/kick`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify({ runId: proxy.runId })
  })
    .then((res) => {
      if (res.ok) {
        message.success('成功下线隧道')
        emit('refresh')
      } else {
        message.warning('下线隧道失败: ' + res.status + ' ' + res.statusText)
      }
    })
    .catch((err) => {
      message.error('下线隧道失败: ' + err.message)
    })
}
</script>

<style lang="scss" scoped>
@use '../assets/styles/proxy-view.scss';
</style>
