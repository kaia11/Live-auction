import { App, Button, Card, Select, Space, Table, Tag } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import AdminLayout from '@/layouts/AdminLayout'
import { OrderProgressModal } from '@/components/modals/StatusModals'
import { getAdminOrders, updateOrderStatus } from '@/api/admin'
import type { AdminOrder, OrderAction } from '@/types'

const orderStatusLabelMap: Record<string, string> = {
  pending_payment: '待支付',
  paid: '已支付待发货',
  shipped: '已发货',
  completed: '已完成',
  cancelled: '已取消',
}

function OrdersPage() {
  const { message } = App.useApp()
  const [status, setStatus] = useState<string>('全部')
  const [selectedOrder, setSelectedOrder] = useState<AdminOrder | null>(null)
  const [orders, setOrders] = useState<AdminOrder[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const loadOrders = async () => {
      setLoading(true)
      try {
        const result = await getAdminOrders()
        setOrders(result)
      } catch (error) {
        const nextMessage =
          error instanceof Error ? error.message : '订单数据加载失败，请稍后重试'
        message.error(nextMessage)
      } finally {
        setLoading(false)
      }
    }

    void loadOrders()
  }, [message])

  const rows = useMemo(() => {
    return orders.filter(
      (item) => status === '全部' || orderStatusLabelMap[item.status] === status,
    )
  }, [orders, status])

  const runOrderAction = async (orderId: string, action: OrderAction) => {
    try {
      await updateOrderStatus(orderId, action)
      message.success('订单状态已更新')
      const result = await getAdminOrders()
      setOrders(result)
      if (selectedOrder?.orderId === orderId) {
        const nextSelected = result.find((item) => item.orderId === orderId) ?? null
        setSelectedOrder(nextSelected)
      }
    } catch (error) {
      const nextMessage =
        error instanceof Error ? error.message : '订单状态更新失败，请稍后重试'
      message.error(nextMessage)
    }
  }

  return (
    <AdminLayout activePath="/orders" title="订单管理">
      <Card className="jewel-card">
        <div className="toolbar">
          <Select
            value={status}
            onChange={setStatus}
            style={{ width: 180 }}
            options={['全部', '待支付', '已支付待发货', '已发货', '已完成', '已取消'].map((item) => ({
              label: item,
              value: item,
            }))}
          />
        </div>
        <Table
          rowKey="orderId"
          dataSource={rows}
          loading={loading}
          pagination={false}
          columns={[
            { title: '订单号', dataIndex: 'orderId' },
            { title: '拍品ID', dataIndex: 'itemId' },
            { title: '买家ID', dataIndex: 'buyerUserId' },
            { title: '成交金额', dataIndex: 'amount', render: (v: number) => `¥${v}` },
            {
              title: '订单状态',
              dataIndex: 'status',
              render: (v: AdminOrder['status']) => (
                <Tag
                  color={
                    v === 'completed'
                      ? 'success'
                      : v === 'shipped'
                        ? 'processing'
                        : v === 'cancelled'
                          ? 'error'
                          : 'gold'
                  }
                >
                  {orderStatusLabelMap[v]}
                </Tag>
              ),
            },
            { title: '物流单号', render: () => '--' },
            { title: '下单时间', dataIndex: 'createTime' },
            {
              title: '操作',
              render: (_, row: AdminOrder) => (
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
        orderId={selectedOrder?.orderId ?? ''}
        status={selectedOrder ? orderStatusLabelMap[selectedOrder.status] : ''}
        logisticsNo={undefined}
        actions={
          selectedOrder ? (
            <Space wrap>
              {selectedOrder.status === 'paid' ? (
                <Button type="primary" onClick={() => void runOrderAction(selectedOrder.orderId, 'ship')}>
                  发货
                </Button>
              ) : null}
              {selectedOrder.status === 'shipped' ? (
                <Button onClick={() => void runOrderAction(selectedOrder.orderId, 'complete')}>
                  完成订单
                </Button>
              ) : null}
              {['pending_payment', 'paid'].includes(selectedOrder.status) ? (
                <Button danger onClick={() => void runOrderAction(selectedOrder.orderId, 'cancel')}>
                  取消订单
                </Button>
              ) : null}
            </Space>
          ) : null
        }
        onClose={() => setSelectedOrder(null)}
      />
    </AdminLayout>
  )
}

export default OrdersPage
