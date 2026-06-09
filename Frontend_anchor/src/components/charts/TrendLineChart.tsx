import { Empty } from 'antd'
import { useMemo } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import type { AdminSession, AuctionItem, DashboardTimelineEvent } from '@/types'
import { buildPriceLineModel, buildTrafficBarData, formatRelativeSeconds } from '@/utils/trendChart'
import './TrendLineChart.scss'

interface Props {
  mode: 'traffic' | 'price'
  events: DashboardTimelineEvent[]
  items: AuctionItem[]
  sessions: AdminSession[]
}

function TrendLineChart({ mode, events, items, sessions }: Props) {
  const trafficData = useMemo(
    () => buildTrafficBarData(items, sessions),
    [items, sessions],
  )
  const priceModel = useMemo(
    () => buildPriceLineModel(events, items, sessions),
    [events, items, sessions],
  )

  if (items.length === 0) {
    return (
      <div className="trend-line-chart-empty">
        <Empty description="请先选择直播间并添加拍品" />
      </div>
    )
  }

  if (mode === 'traffic') {
    return (
      <div className="trend-line-chart">
        <p className="trend-line-chart-caption">
          展示当前直播间最近 6 个拍品在各自竞拍时段内的观看人数
        </p>
        <ResponsiveContainer width="100%" height={320}>
          <BarChart data={trafficData} margin={{ top: 8, right: 12, left: 0, bottom: 36 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 10, fill: '#6b7280' }}
              interval={0}
              angle={-18}
              textAnchor="end"
              height={56}
            />
            <YAxis
              tick={{ fontSize: 11, fill: '#6b7280' }}
              allowDecimals={false}
              label={{
                value: '观看人数',
                angle: -90,
                position: 'insideLeft',
                fill: '#9ca3af',
                fontSize: 11,
              }}
            />
            <Tooltip formatter={(value) => [`${value} 人`, '观看人数']} />
            <Bar dataKey="viewerCount" name="观看人数" radius={[8, 8, 0, 0]}>
              {trafficData.map((entry, index) => (
                <Cell
                  key={entry.itemId}
                  fill={['#2563eb', '#db2777', '#ea580c', '#059669', '#7c3aed', '#0891b2'][index % 6]}
                />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    )
  }

  if (priceModel.data.length === 0 || priceModel.series.length === 0) {
    return (
      <div className="trend-line-chart-empty">
        <Empty description="暂无出价数据，开拍后将自动生成价格轨迹" />
      </div>
    )
  }

  return (
    <div className="trend-line-chart">
      <p className="trend-line-chart-caption">
        展示最近 6 个拍品在各自竞拍时段内的相对时间与出价攀升轨迹
      </p>
      <ResponsiveContainer width="100%" height={320}>
        <LineChart data={priceModel.data} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis
            dataKey="relativeSeconds"
            type="number"
            domain={['dataMin', 'dataMax']}
            tick={{ fontSize: 11, fill: '#6b7280' }}
            tickFormatter={(value) => formatRelativeSeconds(Number(value))}
            label={{
              value: '相对开拍时间',
              position: 'insideBottom',
              offset: -2,
              fill: '#9ca3af',
              fontSize: 11,
            }}
          />
          <YAxis
            tick={{ fontSize: 11, fill: '#6b7280' }}
            allowDecimals={false}
            label={{
              value: '当前价 (¥)',
              angle: -90,
              position: 'insideLeft',
              fill: '#9ca3af',
              fontSize: 11,
            }}
          />
          <Tooltip
            formatter={(value, name) => {
              const numericValue = typeof value === 'number' ? value : Number(value ?? 0)
              const seriesKey = String(name ?? '')
              const label =
                priceModel.series.find((entry) => entry.key === seriesKey)?.label ?? seriesKey
              return [`¥ ${numericValue}`, label]
            }}
            labelFormatter={(label) => `开拍后 ${formatRelativeSeconds(Number(label))}`}
          />
          <Legend
            formatter={(value) =>
              priceModel.series.find((entry) => entry.key === value)?.label ?? value
            }
          />
          {priceModel.series.map((entry) => (
            <Line
              key={entry.key}
              type="monotone"
              dataKey={entry.key}
              name={entry.key}
              stroke={entry.color}
              strokeWidth={2}
              dot={{ r: 3 }}
              activeDot={{ r: 5 }}
              connectNulls
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

export default TrendLineChart
