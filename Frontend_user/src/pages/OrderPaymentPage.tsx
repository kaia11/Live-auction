import React, { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { Button, Image, Toast } from 'antd-mobile'
import { getMyOrders, payOrder } from '../api/orders'
import { getMyBidHistories } from '../api/bids'
import { mapBidHistories } from '../adapters/auction'
import type { BidHistory, UserOrder } from '../types'
import './OrderPaymentPage.scss'

type LocationState = {
  itemTitle?: string
  itemImage?: string
}

const orderStatusLabelMap: Record<UserOrder['status'], string> = {
  pending_payment: '待支付',
  paid: '已支付待发货',
  shipped: '已发货',
  completed: '已完成',
  cancelled: '已取消',
}

const OrderPaymentPage: React.FC = () => {
  const navigate = useNavigate()
  const { orderId = '' } = useParams()
  const location = useLocation()
  const state = (location.state ?? {}) as LocationState

  const [order, setOrder] = useState<UserOrder | null>(null)
  const [histories, setHistories] = useState<BidHistory[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    const load = async () => {
      try {
        const [orders, bidHistories] = await Promise.all([
          getMyOrders(),
          getMyBidHistories(),
        ])
        setOrder(orders.find((entry) => entry.id === orderId) ?? null)
        setHistories(mapBidHistories(bidHistories))
      } catch (error) {
        const message = error instanceof Error ? error.message : '订单信息加载失败'
        Toast.show({ icon: 'fail', content: message })
      } finally {
        setLoading(false)
      }
    }

    void load()
  }, [orderId])

  const bidHistory = useMemo(
    () => histories.find((entry) => entry.itemId === order?.itemId),
    [histories, order?.itemId],
  )

  const itemTitle = bidHistory?.itemTitle ?? state.itemTitle ?? '成交拍品'
  const itemImage = bidHistory?.itemImage ?? state.itemImage ?? 'https://picsum.photos/400/400'

  const handlePay = async () => {
    if (!order || order.status !== 'pending_payment') {
      return
    }

    setSubmitting(true)
    try {
      const paidOrder = await payOrder(order.id)
      setOrder(paidOrder)
      Toast.show({ icon: 'success', content: '支付成功' })
    } catch (error) {
      const message = error instanceof Error ? error.message : '支付失败，请稍后重试'
      Toast.show({ icon: 'fail', content: message })
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return <div className="order-payment-page">加载中...</div>
  }

  if (!order) {
    return (
      <div className="order-payment-page">
        <div className="order-payment-card">
          <h2 className="payment-title">订单不存在</h2>
          <Button color="primary" onClick={() => navigate('/profile')}>
            返回个人主页
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="order-payment-page">
      <div className="order-payment-card">
        <div className="payment-scroll-content">
          <div className="payment-top-nav">
            <span className="payment-back" onClick={() => navigate('/profile')}>‹</span>
            <span className="payment-nav-title">订单支付</span>
            <span className="payment-nav-placeholder" />
          </div>

          <Image className="payment-item-image" src={itemImage} fit="cover" />

          <div className="payment-item-title">{itemTitle}</div>
          <div className="payment-status">{orderStatusLabelMap[order.status]}</div>

          <div className="payment-detail-list">
            <div className="payment-detail-row">
              <span>订单号</span>
              <span>{order.id}</span>
            </div>
            <div className="payment-detail-row">
              <span>拍品 ID</span>
              <span>{order.itemId}</span>
            </div>
            <div className="payment-detail-row">
              <span>成交金额</span>
              <span className="payment-amount">¥{order.amount}</span>
            </div>
            <div className="payment-detail-row">
              <span>创建时间</span>
              <span>{order.createTime}</span>
            </div>
          </div>
        </div>

        <div className="payment-action-bar">
          {order.status === 'pending_payment' ? (
            <Button
              block
              color="primary"
              loading={submitting}
              className="payment-submit-btn"
              onClick={() => void handlePay()}
            >
              确认支付
            </Button>
          ) : (
            <Button block className="payment-return-btn" onClick={() => navigate('/profile')}>
              返回个人主页
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

export default OrderPaymentPage
