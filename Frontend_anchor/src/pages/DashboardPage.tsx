import { Button, Card, Col, Divider, Row, Table, Tag } from 'antd'
import { useMemo, useState } from 'react'
import AdminLayout from '@/layouts/AdminLayout'
import StatCard from '@/components/common/StatCard'
import { dashboardStats, mockGoods } from '@/mock/data'

function DashboardPage() {
  const [activeMetric, setActiveMetric] = useState<'traffic' | 'price'>('traffic')

  const chartRows = useMemo(
    () =>
      mockGoods.map((item) => ({
        key: item.id,
        name: item.name,
        data: activeMetric === 'traffic' ? item.traffic.join(' / ') : item.priceTrack.join(' / '),
      })),
    [activeMetric],
  )

  return (
    <AdminLayout
      activePath="/dashboard"
      title="运营总览"
      actions={
        <Button type="primary" href="#/publish">
          + 添加商品
        </Button>
      }
    >
      <Row gutter={16}>
        <Col span={6}>
          <StatCard label="在拍中拍品" value={`${dashboardStats.inProgress}`} tone="normal" />
        </Col>
        <Col span={6}>
          <StatCard label="已成交拍品" value={`${dashboardStats.sold}`} tone="success" />
        </Col>
        <Col span={6}>
          <StatCard label="流拍拍品" value={`${dashboardStats.unsold}`} tone="warning" />
        </Col>
        <Col span={6}>
          <StatCard label="成交总额" value={`¥ ${dashboardStats.totalAmount.toLocaleString()}`} />
        </Col>
      </Row>

      <Row gutter={16} className="section-gap">
        <Col span={16}>
          <Card title="商品状态分布看板" className="jewel-card">
            <p className="subtle-line">
              在线人数：{dashboardStats.onlineCount} | 今日出价次数：{dashboardStats.bidsToday}
            </p>
            <div className="goods-status-legend">
              <Tag color="processing">竞拍中</Tag>
              <Tag color="success">已成交</Tag>
              <Tag color="default">待上架</Tag>
              <Tag color="warning">已流拍</Tag>
            </div>
            <Table
              size="small"
              pagination={false}
              dataSource={mockGoods.map((item) => ({
                key: item.id,
                id: item.id,
                name: item.name,
                status: item.status,
                currentPrice: item.currentPrice,
                bidCount: item.bidCount,
              }))}
              columns={[
                { title: '拍品ID', dataIndex: 'id' },
                { title: '拍品', dataIndex: 'name' },
                { title: '状态', dataIndex: 'status' },
                { title: '当前价', dataIndex: 'currentPrice', render: (v) => `¥ ${v}` },
                { title: '出价次数', dataIndex: 'bidCount' },
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
