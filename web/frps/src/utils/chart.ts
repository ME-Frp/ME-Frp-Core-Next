import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import * as Humanize from 'humanize-plus'

export function DrawTrafficChart(chart: echarts.ECharts, trafficIn: number, trafficOut: number) {
  const option: EChartsOption = {
    title: {
      text: '流量统计'
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      data: ['入站流量', '出站流量']
    },
    series: [
      {
        type: 'pie',
        radius: '50%',
        data: [
          { value: trafficIn, name: '入站流量' },
          { value: trafficOut, name: '出站流量' }
        ],
        label: {
          formatter: (params: any) => {
            return `${params.name}: ${Humanize.fileSize(params.value)}`
          }
        }
      }
    ]
  }
  chart.setOption(option)
}

export function DrawProxyChart(chart: echarts.ECharts, serverInfo: any) {
  const proxyTypeCount = serverInfo.proxyTypeCount || {}
  const data = [
    { value: proxyTypeCount.tcp || 0, name: 'TCP' },
    { value: proxyTypeCount.udp || 0, name: 'UDP' },
    { value: proxyTypeCount.http || 0, name: 'HTTP' },
    { value: proxyTypeCount.https || 0, name: 'HTTPS' },
    { value: proxyTypeCount.stcp || 0, name: 'STCP' },
    { value: proxyTypeCount.sudp || 0, name: 'SUDP' },
    { value: proxyTypeCount.xtcp || 0, name: 'XTCP' }
  ].filter(item => item.value > 0)

  const option: EChartsOption = {
    title: {
      text: '代理类型统计'
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      data: data.map(item => item.name)
    },
    series: [
      {
        type: 'pie',
        radius: '50%',
        data: data
      }
    ]
  }
  chart.setOption(option)
}
