import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import * as Humanize from 'humanize-plus'

// 获取当前主题的文本颜色
const getTextColor = () => {
  const el = document.querySelector('body')
  if (!el) return '#333'
  const style = window.getComputedStyle(el)
  return style.getPropertyValue('--n-text-color').trim()
}

export function DrawTrafficChart(chart: echarts.ECharts, trafficIn: number, trafficOut: number) {
  const textColor = getTextColor()
  const option: EChartsOption = {
    title: {
      text: '流量统计',
      left: 'center',
      top: 0,
      textStyle: {
        color: textColor
      }
    },
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        return `${params.name}: ${Humanize.fileSize(params.value)}<br/>占比: ${params.percent}%`
      }
    },
    legend: {
      orient: 'horizontal',
      bottom: '5%',
      left: 'center',
      textStyle: {
        color: textColor
      }
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 10,
          borderColor: 'var(--n-color)',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: (params: any) => {
            return `${params.name}\n${Humanize.fileSize(params.value)}`
          },
          color: textColor
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold',
            color: textColor
          }
        },
        data: [
          { 
            value: trafficIn, 
            name: '入站流量',
            itemStyle: {
              color: '#91cc75'
            }
          },
          { 
            value: trafficOut, 
            name: '出站流量',
            itemStyle: {
              color: '#5470c6'
            }
          }
        ]
      }
    ]
  }
  chart.setOption(option)
}

export function DrawProxyChart(chart: echarts.ECharts, serverInfo: any) {
  const textColor = getTextColor()
  const proxyTypeCount = serverInfo.proxyTypeCount || {}
  const data = [
    { value: proxyTypeCount.tcp || 0, name: 'TCP', itemStyle: { color: '#5470c6' } },
    { value: proxyTypeCount.udp || 0, name: 'UDP', itemStyle: { color: '#91cc75' } },
    { value: proxyTypeCount.http || 0, name: 'HTTP', itemStyle: { color: '#fac858' } },
    { value: proxyTypeCount.https || 0, name: 'HTTPS', itemStyle: { color: '#ee6666' } },
    { value: proxyTypeCount.stcp || 0, name: 'STCP', itemStyle: { color: '#73c0de' } },
    { value: proxyTypeCount.sudp || 0, name: 'SUDP', itemStyle: { color: '#3ba272' } },
    { value: proxyTypeCount.xtcp || 0, name: 'XTCP', itemStyle: { color: '#fc8452' } }
  ].filter(item => item.value > 0)

  const option: EChartsOption = {
    title: {
      text: '隧道类型统计',
      left: 'center',
      top: 0,
      textStyle: {
        color: textColor
      }
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'horizontal',
      bottom: '5%',
      left: 'center',
      textStyle: {
        color: textColor
      }
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 10,
          borderColor: 'var(--n-color)',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}: {c}',
          color: textColor
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold',
            color: textColor
          }
        },
        data: data
      }
    ]
  }
  chart.setOption(option)
}
