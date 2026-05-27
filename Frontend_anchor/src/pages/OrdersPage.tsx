import { Button, Card, Select, Space, Table, Tag } from 'antd'
import { useMemo, useState } from 'react'
import AdminLayout from '@/layouts/AdminLayout'
import { mockOrders } from '@/mock/data'
import { OrderProgressModal } from '@/components/modals/StatusModals'
import type { OrderItem } from '@/types'

function OrdersPage() {
  const [status, setStatus] = useState<string>('全部')
  const [selectedOrder, setSelectedOrder] = useState<OrderItem | null>(null)

  const rows = useMemo(
    () => mockOrders.filter((item) => status === '全部' || item.status === status),
    [status],
  )

  return (
    <AdminLayout activePath="/orders" title="订单管理">
      <Card className="jewel-card">
        <div className="toolbar">
          <Select
            value={status}
            onChange={setStatus}
            style={{ width: 180 }}
            options={['全部', '待发货', '已发货', '已完成'].map((item) => ({
              label: item,
              value: item,
            }))}
          />
        </div>
        <Table
          rowKey="id"
          dataSource={rows}
          pagination={false}
          columns={[
            { title: '订单号', dataIndex: 'id' },
            { title: '拍品', dataIndex: 'goodsName' },
            { title: '买家', dataIndex: 'buyerName' },
            { title: '成交金额', dataIndex: 'amount', render: (v: number) => `¥${v}` },
            {
              title: '订单状态',
              dataIndex: 'status',
              render: (v: OrderItem['status']) => (
                <Tag color={v === '已完成' ? 'success' : v === '已发货' ? 'processing' : 'gold'}>
                  {v}
                </Tag>
              ),
            },
            { title: '物流单号', dataIndex: 'logisticsNo', render: (v?: string) => v ?? '--' },
            { title: '下单时间', dataIndex: 'createdAt' },
            {
              title: '操作',
              render: (_, row: OrderItem) => (
                <Space>
                  <Button
                    size="small"
                    onClick={() => {
                      setSelectedOrder(row)
                    }}
                  >
                    查看进度
                  </Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <OrderProgressModal
        open={Boolean(selectedOrder)}
        orderId={selectedOrder?.id ?? ''}
        status={selectedOrder?.status ?? ''}
        logisticsNo={selectedOrder?.logisticsNo}
        onClose={() => setSelectedOrder(null)}
      />
    </AdminLayout>
  )
}

export default OrdersPage
