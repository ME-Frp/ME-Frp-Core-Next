<template>
  <div>
    <div ref="chartRef" style="width: 100%; height: 400px"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import * as Humanize from 'humanize-plus'

const props = defineProps<{
  proxyName: string
}>()

const chartRef = ref<HTMLElement>()
let chart: echarts.ECharts | null = null
let timer: number | null = null

const option: EChartsOption = {
  title: {
    text: '实时流量监控'
  },
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'cross',
      label: {
        backgroundColor: '#6a7985'
      }
    }
  },
  legend: {
    data: ['入站流量', '出站流量']
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: [
    {
      type: 'category',
      boundaryGap: false,
      data: []
    }
  ],
  yAxis: [
    {
      type: 'value',
      axisLabel: {
        formatter: (value: number) => {
          return Humanize.fileSize(value)
        }
      }
    }
  ],
  series: [
    {
      name: '入站流量',
      type: 'line',
      areaStyle: {},
      emphasis: {
        focus: 'series'
      },
      data: []
    },
    {
      name: '出站流量',
      type: 'line',
      areaStyle: {},
      emphasis: {
        focus: 'series'
      },
      data: []
    }
  ]
}

const fetchData = () => {
  fetch(`../api/traffic/${props.proxyName}`, { credentials: 'include' })
    .then((res) => res.json())
    .then((json) => {
      if (chart) {
        const xAxisData = json.names
        const inData = json.trafficIn
        const outData = json.trafficOut

        chart.setOption({
          xAxis: [
            {
              data: xAxisData
            }
          ],
          series: [
            {
              data: inData
            },
            {
              data: outData
            }
          ]
        })
      }
    })
}

onMounted(() => {
  if (chartRef.value) {
    chart = echarts.init(chartRef.value)
    chart.setOption(option)
    fetchData()
    timer = window.setInterval(fetchData, 3000)
  }
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  if (chart) {
    chart.dispose()
    chart = null
  }
})
</script>

<style lang="scss" scoped>
@use '../assets/styles/traffic.scss';
</style>
