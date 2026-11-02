<template>
  <div ref="root" class="dashboard-trend-chart" role="img" :aria-label="description" />
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { use, init, type ECharts, type EChartsCoreOption } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([CanvasRenderer, LineChart, GridComponent, LegendComponent, TooltipComponent])

const props = defineProps<{
  option: EChartsCoreOption
  description: string
}>()

const root = ref<HTMLDivElement | null>(null)
let chart: ECharts | undefined
let resizeObserver: ResizeObserver | undefined

function renderChart() {
  if (!root.value) return
  chart ||= init(root.value)
  const option = JSON.parse(JSON.stringify(props.option)) as EChartsCoreOption
  chart.setOption(option, { notMerge: true })
}

watch(() => props.option, () => {
  void nextTick(renderChart)
}, { deep: true })

onMounted(() => {
  renderChart()
  if (!root.value) return
  resizeObserver = new ResizeObserver(() => chart?.resize())
  resizeObserver.observe(root.value)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = undefined
  chart?.dispose()
  chart = undefined
})
</script>

<style scoped>
.dashboard-trend-chart {
  width: 100%;
  min-width: 0;
  height: 100%;
}
</style>

