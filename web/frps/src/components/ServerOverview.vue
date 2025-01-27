<template>
  <div class="overview-container">
    <!-- 左侧信息栏 -->
    <div class="info-section">
      <div class="card-actions">
        <NButton secondary @click="fetchData" style="margin-right: 12px">
          <template #icon>
            <NIcon><refresh-outline /></NIcon>
          </template>
          刷新节点信息
        </NButton>
        <NPopconfirm
          @positive-click="handleRestart"
          positive-text="确认"
          negative-text="取消"
        >
          <template #trigger>
            <NButton type="warning">
              <template #icon>
                <NIcon><power-outline /></NIcon>
              </template>
              重启当前节点服务
            </NButton>
          </template>
          确定要重启当前节点吗？这可能会导致所有隧道短暂中断。
        </NPopconfirm>
      </div>
      <NCard title="服务端信息">
        <NDescriptions :column="1" label-placement="left">
          <NDescriptionsItem label="版本">
            {{ data.version }}
          </NDescriptionsItem>
          <NDescriptionsItem label="服务端口">
            {{ data.bindPort }}
          </NDescriptionsItem>
          <NDescriptionsItem v-if="data.kcpBindPort != 0" label="KCP 绑定端口">
            {{ data.kcpBindPort }}
          </NDescriptionsItem>
          <NDescriptionsItem label="最大连接池">
            {{ data.maxPoolCount }}
          </NDescriptionsItem>
          <NDescriptionsItem label="每客户端最大端口数">
            {{ data.maxPortsPerClient }}
          </NDescriptionsItem>
          <NDescriptionsItem label="端口范围限制">
            {{ data.allowPortsStr }}
          </NDescriptionsItem>
          <NDescriptionsItem label="心跳超时">
            {{ data.heartbeatTimeout }}
          </NDescriptionsItem>
          <NDescriptionsItem label="客户端数量">
            {{ data.clientCounts }}
          </NDescriptionsItem>
          <NDescriptionsItem label="当前连接数">
            {{ data.curConns }}
          </NDescriptionsItem>
          <NDescriptionsItem label="隧道数量">
            {{ data.proxyCounts }}
          </NDescriptionsItem>
        </NDescriptions>
      </NCard>
    </div>

    <!-- 右侧图表区域 -->
    <div class="charts-section">
      <NCard>
        <div class="chart-container">
          <div class="chart-header">
            <div>
              <div class="chart-title">流量统计</div>
              <div class="chart-subtitle">今日</div>
            </div>
          </div>
          <div ref="trafficRef" class="chart" />
        </div>
      </NCard>
      <NCard style="margin-top: 16px">
        <div class="chart-container">
          <div class="chart-header">
            <div>
              <div class="chart-title">隧道统计</div>
              <div class="chart-subtitle">当前</div>
            </div>
          </div>
          <div ref="proxiesRef" class="chart" />
        </div>
      </NCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useMessage } from 'naive-ui'
import * as echarts from 'echarts'
import { DrawTrafficChart, DrawProxyChart } from '../utils/chart'
import { RefreshOutline, PowerOutline } from '@vicons/ionicons5'

const message = useMessage()
const trafficRef = ref<HTMLElement>()
const proxiesRef = ref<HTMLElement>()
let trafficChart: echarts.ECharts | null = null
let proxiesChart: echarts.ECharts | null = null
let refreshTimer: number | null = null

let data = ref({
  version: '',
  bindPort: 0,
  kcpBindPort: 0,
  quicBindPort: 0,
  vhostHTTPPort: 0,
  vhostHTTPSPort: 0,
  tcpmuxHTTPConnectPort: 0,
  subdomainHost: '',
  maxPoolCount: 0,
  maxPortsPerClient: '',
  allowPortsStr: '',
  tlsForce: false,
  heartbeatTimeout: 0,
  clientCounts: 0,
  curConns: 0,
  proxyCounts: 0,
})

const initCharts = () => {
  if (trafficRef.value && proxiesRef.value) {
    // 销毁旧的图表实例
    trafficChart?.dispose()
    proxiesChart?.dispose()
    
    // 创建新的图表实例
    trafficChart = echarts.init(trafficRef.value)
    proxiesChart = echarts.init(proxiesRef.value)
  }
}

const handleResize = () => {
  trafficChart?.resize()
  proxiesChart?.resize()
}

const fetchData = () => {
  fetch('../api/serverinfo', { credentials: 'include' })
    .then((res) => res.json())
    .then((json) => {
      data.value.version = json.version
      data.value.bindPort = json.bindPort
      data.value.kcpBindPort = json.kcpBindPort
      data.value.quicBindPort = json.quicBindPort
      data.value.vhostHTTPPort = json.vhostHTTPPort
      data.value.vhostHTTPSPort = json.vhostHTTPSPort
      data.value.tcpmuxHTTPConnectPort = json.tcpmuxHTTPConnectPort
      data.value.subdomainHost = json.subdomainHost
      data.value.maxPoolCount = json.maxPoolCount
      data.value.maxPortsPerClient = json.maxPortsPerClient
      if (data.value.maxPortsPerClient == '0') {
        data.value.maxPortsPerClient = '无限制'
      }
      data.value.allowPortsStr = json.allowPortsStr
      if (data.value.allowPortsStr == '') {
        data.value.allowPortsStr = '未设置'
      }
      data.value.tlsForce = json.tlsForce
      data.value.heartbeatTimeout = json.heartbeatTimeout
      data.value.clientCounts = json.clientCounts
      data.value.curConns = json.curConns
      data.value.proxyCounts = 0
      if (json.proxyTypeCount != null) {
        if (json.proxyTypeCount.tcp != null) {
          data.value.proxyCounts += json.proxyTypeCount.tcp
        }
        if (json.proxyTypeCount.udp != null) {
          data.value.proxyCounts += json.proxyTypeCount.udp
        }
        if (json.proxyTypeCount.http != null) {
          data.value.proxyCounts += json.proxyTypeCount.http
        }
        if (json.proxyTypeCount.https != null) {
          data.value.proxyCounts += json.proxyTypeCount.https
        }
        if (json.proxyTypeCount.stcp != null) {
          data.value.proxyCounts += json.proxyTypeCount.stcp
        }
        if (json.proxyTypeCount.sudp != null) {
          data.value.proxyCounts += json.proxyTypeCount.sudp
        }
        if (json.proxyTypeCount.xtcp != null) {
          data.value.proxyCounts += json.proxyTypeCount.xtcp
        }
      }

      // 更新图表
      if (trafficRef.value && proxiesRef.value) {
        if (!trafficChart || !proxiesChart) {
          initCharts()
        }
        DrawTrafficChart(trafficChart!, json.totalTrafficIn, json.totalTrafficOut)
        DrawProxyChart(proxiesChart!, json)
      }
    })
    .catch(() => {
      message.warning('从 Frp 服务端获取服务器信息失败！')
    })
}

const handleRestart = () => {
  fetch('../api/serverinfo/restart', {
    method: 'POST',
    credentials: 'include'
  })
    .then((res) => res.json())
    .then((json) => {
      if (json.code === 200) {
        message.success('重启指令已发送')
      } else {
        message.error('重启失败：' + json.msg)
      }
    })
    .catch(() => {
      message.error('重启失败：网络错误')
    })
}

onMounted(() => {
  initCharts() // 初始化图表
  fetchData()   // 获取数据
  window.addEventListener('resize', handleResize)
  // 每30秒自动刷新一次
  refreshTimer = window.setInterval(fetchData, 30000)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  trafficChart?.dispose()
  proxiesChart?.dispose()
  trafficChart = null
  proxiesChart = null
})

</script>

<style lang="scss" scoped>
@use '../assets/styles/server-overview.scss';

.card-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.chart {
  height: 250px;
  width: 100%;
}
</style>
