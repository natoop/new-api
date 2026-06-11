/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useMemo } from 'react'
import { VChart } from '@visactor/react-vchart'
import { Skeleton } from '@/components/ui/skeleton'
import { useChartTheme } from '@/lib/use-chart-theme'
import { VCHART_OPTION } from '@/lib/vchart'

// Brand series colours sampled from the GS logomark (azure / fresh-green),
// resolved to concrete hex because the canvas renderer cannot read CSS vars.
const SERIES = {
  light: { blue: '#1786e0', green: '#46b23a' },
  dark: { blue: '#4aa8f0', green: '#6cc95a' },
}

export type OpsSeriesTone = 'blue' | 'green'

export interface OpsChartDatum {
  label: string
  value: number
}

interface OpsTimeSeriesChartProps {
  data: OpsChartDatum[]
  kind: 'area' | 'bar'
  tone?: OpsSeriesTone
  loading?: boolean
  height?: number
  valueFormatter?: (value: number) => string
}

export function OpsTimeSeriesChart(props: OpsTimeSeriesChartProps) {
  const { resolvedTheme, themeReady } = useChartTheme()
  const tone = props.tone ?? 'blue'
  const height = props.height ?? 220
  const color = SERIES[resolvedTheme === 'dark' ? 'dark' : 'light'][tone]

  const spec = useMemo(() => {
    const values = props.data.map((d) => ({ label: d.label, value: d.value }))
    const base = {
      type: props.kind,
      data: [{ id: 'ops', values }],
      xField: 'label',
      yField: 'value',
      background: 'transparent',
      padding: { top: 12, right: 12, bottom: 8, left: 8 },
      axes: [
        {
          orient: 'left',
          grid: { visible: true },
          label: {
            formatMethod: (v: unknown) =>
              props.valueFormatter
                ? props.valueFormatter(Number(v))
                : String(v),
          },
        },
        {
          orient: 'bottom',
          label: { autoRotate: false, autoHide: true },
          sampling: true,
        },
      ],
      tooltip: {
        mark: {
          content: [
            {
              key: (d: OpsChartDatum) => d.label,
              value: (d: OpsChartDatum) =>
                props.valueFormatter
                  ? props.valueFormatter(d.value)
                  : String(d.value),
            },
          ],
        },
      },
    }
    if (props.kind === 'area') {
      return {
        ...base,
        area: { style: { fill: color, fillOpacity: 0.16 } },
        line: { style: { stroke: color, lineWidth: 2, lineCap: 'round' } },
        point: { visible: false },
      }
    }
    return {
      ...base,
      bar: {
        style: { fill: color, cornerRadius: [4, 4, 0, 0], fillOpacity: 0.9 },
      },
    }
  }, [props.data, props.kind, props.valueFormatter, color])

  if (props.loading || !themeReady) {
    return <Skeleton style={{ height }} className='w-full rounded-lg' />
  }

  const chartKey = [
    props.kind,
    tone,
    resolvedTheme,
    props.data.length,
    props.data[props.data.length - 1]?.label ?? '',
  ].join('-')

  return (
    <div style={{ height }} className='w-full'>
      <VChart key={chartKey} spec={spec} option={VCHART_OPTION} />
    </div>
  )
}
