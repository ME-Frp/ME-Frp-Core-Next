<template>
  <div class="overview-container">
    <!-- 左侧信息栏 -->
    <div class="info-section">
      <n-descriptions :column="1" label-placement="left">
        <n-descriptions-item label="版本">
          {{ data.version }}
        </n-descriptions-item>
        <n-descriptions-item label="绑定端口">
          {{ data.bindPort }}
        </n-descriptions-item>
        <n-descriptions-item v-if="data.kcpBindPort != 0" label="KCP绑定端口">
          {{ data.kcpBindPort }}
        </n-descriptions-item>
        <n-descriptions-item label="最大连接池">
          {{ data.maxPoolCount }}
        </n-descriptions-item>
        <n-descriptions-item label="每客户端最大端口数">
          {{ data.maxPortsPerClient }}
        </n-descriptions-item>
        <n-descriptions-item v-if="data.allowPortsStr != ''" label="允许端口">
          <LongSpan :content="data.allowPortsStr" :length="30" />
        </n-descriptions-item>
        <n-descriptions-item label="心跳超时">
          {{ data.heartbeatTimeout }}
        </n-descriptions-item>
        <n-descriptions-item label="客户端数量">
          {{ data.clientCounts }}
        </n-descriptions-item>
        <n-descriptions-item label="当前连接数">
          {{ data.curConns }}
        </n-descriptions-item>
        <n-descriptions-item label="隧道数量">
          {{ data.proxyCounts }}
        </n-descriptions-item>
      </n-descriptions>
    </div>

    <!-- 右侧图表区域 -->
    <div class="charts-section">
      <div class="chart-container">
        <div class="chart-title">流量统计</div>
        <div class="chart-subtitle">今日</div>
        <div ref="trafficRef" class="chart" />
      </div>
      <div class="chart-container">
        <div class="chart-title">隧道统计</div>
        <div class="chart-subtitle">当前</div>
        <div ref="proxiesRef" class="chart" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useMessage } from 'naive-ui'
import * as echarts from 'echarts'
import { DrawTrafficChart, DrawProxyChart } from '../utils/chart'
import LongSpan from './LongSpan.vue'

const message = useMessage()
const trafficRef = ref<HTMLElement>()
const proxiesRef = ref<HTMLElement>()
let trafficChart: echarts.ECharts | null = null
let proxiesChart: echarts.ECharts | null = null

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

onMounted(() => {
  initCharts() // 初始化图表
  fetchData()   // 获取数据
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trafficChart?.dispose()
  proxiesChart?.dispose()
  trafficChart = null
  proxiesChart = null
})
</script>

<style lang="scss" scoped>
@use '../assets/styles/server-overview.scss';
</style>
