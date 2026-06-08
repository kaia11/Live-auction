import { App, Button, Card, Col, Divider, Row, Space, Table, Tag } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { AxiosError } from 'axios'
import { useNavigate } from 'react-router-dom'
import AdminLayout from '@/layouts/AdminLayout'
import StatCard from '@/components/common/StatCard'
import {
  getOverview,
  getRoomItems,
  getTimeline,
  startRoomLive,
  stopRoomLive,
} from '@/api/admin'
import { getMerchantRoomImage } from '@/assets/localImages'
import type { AuctionItem, DashboardOverview, DashboardTimelineEvent } from '@/types'
import { useAdminStore } from '@/stores/useAdminStore'

const getApiErrorMessage = (error: unknown, fallback: string) => {
  const err = error as AxiosError<{ message?: string }>
  const apiMessage = err.response?.data?.message
  if (apiMessage) {
    return apiMessage
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  return fallback
}

function DashboardPage() {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [activeMetric, setActiveMetric] = useState<'traffic' | 'price'>('traffic')
  const [overview, setOverview] = useState<DashboardOverview | null>(null)
  const [items, setItems] = useState<AuctionItem[]>([])
  const [timeline, setTimeline] = useState<DashboardTimelineEvent[]>([])
  const currentRoomId = useAdminStore((state) => state.currentRoomId)
  const rooms = useAdminStore((state) => state.rooms)
  const updateRoomStatus = useAdminStore((state) => state.updateRoomStatus)
  const currentRoom = useMemo(
    () => rooms.find((room) => room.id === currentRoomId),
    [rooms, currentRoomId],
  )

  useEffect(() => {
    const loadDashboard = async () => {
      try {
        const [overviewResult, timelineResult] = await Promise.all([getOverview(), getTimeline()])
        setOverview(overviewResult)
        setTimeline(timelineResult)
        if (currentRoomId) {
          const itemResult = await getRoomItems(currentRoomId)
          setItems(itemResult)
        } else {
          setItems([])
        }
      } catch (error) {
        const nextMessage =
          error instanceof Error ? error.message : '总览数据加载失败，请稍后重试'
        message.error(nextMessage)
      }
    }

    void loadDashboard()
  }, [currentRoomId, message])

  const chartRows = useMemo(
    () =>
      timeline.slice(0, 6).map((item) => ({
        key: `${item.sessionId}-${item.time}`,
        name: item.itemId,
        data:
          activeMetric === 'traffic'
            ? `${item.event} / ${item.userId}`
            : `¥ ${item.price} / ${item.time}`,
      })),
    [activeMetric, timeline],
  )

  const itemRows = useMemo(
    () =>
      items.map((item) => ({
        key: item.id,
        id: item.id,
        name: item.title,
        status: item.queueStatus,
        currentPrice: item.startPrice,
        incrementStep: item.incrementStep,
      })),
    [items],
  )

  return (
    <AdminLayout
      activePath="/dashboard"
      title="运营总览"
      actions={
        <Space>
          <Button
            type="primary"
            disabled={currentRoom?.status === 'live'}
            onClick={async () => {
              if (!currentRoomId) {
                message.warning('请先选择直播间')
                return
              }

              try {
                await startRoomLive(currentRoomId)
                updateRoomStatus(currentRoomId, 'live')
                message.success('直播已开始')
              } catch (error) {
                message.error(getApiErrorMessage(error, '开始直播失败，请稍后重试'))
              }
            }}
          >
            {currentRoom?.status === 'live' ? '直播中' : '开始直播'}
          </Button>
          <Button
            danger
            disabled={currentRoom?.status !== 'live'}
            onClick={async () => {
              if (!currentRoomId) {
                message.warning('请先选择直播间')
                return
              }

              try {
                await stopRoomLive(currentRoomId)
                updateRoomStatus(currentRoomId, 'offline')
                message.success('已下播')
              } catch (error) {
                message.error(getApiErrorMessage(error, '下播失败，请稍后重试'))
              }
            }}
          >
            下播
          </Button>
          <Button type="primary" onClick={() => navigate('/publish')}>
            + 添加商品
          </Button>
        </Space>
      }
    >
      <Row gutter={16}>
        <Col span={6}>
          <StatCard label="直播间数" value={`${overview?.totalRooms ?? 0}`} tone="normal" />
        </Col>
        <Col span={6}>
          <StatCard label="总场次数" value={`${overview?.totalSessions ?? 0}`} tone="normal" />
        </Col>
        <Col span={6}>
          <StatCard label="已成交场次" value={`${overview?.soldSessions ?? 0}`} tone="success" />
        </Col>
        <Col span={6}>
          <StatCard label="已取消场次" value={`${overview?.cancelledSessions ?? 0}`} tone="warning" />
        </Col>
      </Row>

      <Row gutter={16} className="section-gap">
        <Col span={16}>
          <Card title="商品状态分布看板" className="jewel-card">
            <p className="subtle-line">
              当前直播间：{currentRoomId || '未选择'} | 价格字段基于房间内拍品快照
            </p>
            <div className="dashboard-mock-preview">
              <img src={getMerchantRoomImage(currentRoomId)} alt="mock live preview" />
              <span>当前直播画面使用商家端本地 mock 图片占位</span>
            </div>
            <div className="goods-status-legend">
              <Tag color="processing">active</Tag>
              <Tag color="success">finished</Tag>
              <Tag color="default">待上架</Tag>
              <Tag color="error">cancelled</Tag>
            </div>
            <Table
              size="small"
              pagination={false}
              dataSource={itemRows}
              columns={[
                { title: '拍品ID', dataIndex: 'id' },
                { title: '拍品', dataIndex: 'name' },
                { title: '状态', dataIndex: 'status' },
                { title: '当前价', dataIndex: 'currentPrice', render: (v) => `¥ ${v}` },
                { title: '加价幅度', dataIndex: 'incrementStep', render: (v) => `¥ ${v}` },
              ]}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card
            title="数据趋势入口"
            className="jewel-card"
            extra={
              <Button.Group>
                <Button
                  type={activeMetric === 'traffic' ? 'primary' : 'default'}
                  onClick={() => setActiveMetric('traffic')}
                >
                  流量
                </Button>
                <Button
                  type={activeMetric === 'price' ? 'primary' : 'default'}
                  onClick={() => setActiveMetric('price')}
                >
                  价格轨迹
                </Button>
              </Button.Group>
            }
          >
            {chartRows.map((row) => (
              <div key={row.key} className="trend-line-item">
                <strong>{row.name}</strong>
                <Divider />
                <span>{row.data}</span>
              </div>
            ))}
          </Card>
        </Col>
      </Row>
    </AdminLayout>
  )
}

export default DashboardPage
