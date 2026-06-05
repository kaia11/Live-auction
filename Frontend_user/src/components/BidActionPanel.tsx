import React, { useEffect, useMemo, useState } from 'react'
import { Button, Loading, Toast } from 'antd-mobile'
import { AxiosError } from 'axios'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './BidActionPanel.scss'

const BidActionPanel: React.FC = () => {
  const { items, currentItemId, closeAllModals, submitBid, bidPanelMode, openBidPanel } = useLiveRoomStore()
  const item = items.find(i => i.id === currentItemId)
  const [biddingPrice, setBiddingPrice] = useState(0)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)

  const increment = item?.minIncrement || 1000
  const minAllowedPrice = item
    ? (item.status === '竞拍中' ? item.currentPrice + increment : Math.max(item.startPrice, item.currentPrice))
    : 0

  const countdownTargetMs = useMemo(() => {
    if (!item?.endTime) {
      return null
    }
    const ts = new Date(item.endTime).getTime()
    return Number.isFinite(ts) ? ts : null
  }, [item?.endTime])

  useEffect(() => {
    setBiddingPrice(minAllowedPrice)
  }, [minAllowedPrice])

  useEffect(() => {
    if (!countdownTargetMs) {
      setCountdown(0)
      return
    }

    const tick = () => {
      const next = Math.max(0, Math.ceil((countdownTargetMs - Date.now()) / 1000))
      setCountdown(next)
    }

    tick()
    const timer = window.setInterval(tick, 1000)
    return () => window.clearInterval(timer)
  }, [countdownTargetMs])

  const handleMinus = () => {
    setBiddingPrice(prev => Math.max(minAllowedPrice, prev - increment))
  }

  const handlePlus = () => {
    const newPrice = biddingPrice + increment
    if (item?.maxPrice && newPrice > item.maxPrice) {
      Toast.show('已达到封顶价')
      return
    }
    setBiddingPrice(newPrice)
  }

  const handleSubmit = async () => {
    if (!item) return
    if (biddingPrice < minAllowedPrice) {
      Toast.show(`当前最低可出价 ¥${minAllowedPrice}`)
      return
    }
    setIsSubmitting(true)
    try {
      const success = await submitBid(biddingPrice)
      if (success) {
        Toast.show('出价成功！')
      }
    } catch (error) {
      const err = error as AxiosError<{ message?: string; code?: number }>
      const message = err.response?.data?.message ?? ''
      if (message.includes('session is not bidding')) {
        Toast.show('当前场次未在竞拍中')
      } else if (message.includes('invalid bid price')) {
        Toast.show(`出价无效，最低可出 ¥${minAllowedPrice}`)
      } else if (message.includes('duplicate bid request')) {
        Toast.show('请求重复，请稍后再试')
      } else {
        Toast.show(message || '出价失败，请稍后重试')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!item) return null

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="bid-panel" onClick={(e) => e.stopPropagation()}>
        <div className="bid-panel-inner">
          <div className="panel-top-bar">
            <div className="panel-tabs">
              <button
                className={`panel-tab ${bidPanelMode === 'detail' ? 'active' : ''}`}
                onClick={() => openBidPanel('detail')}
              >
                商品详情
              </button>
              <button
                className={`panel-tab ${bidPanelMode === 'bid' ? 'active' : ''}`}
                onClick={() => openBidPanel('bid')}
              >
                立即出价
              </button>
            </div>
            <button className="panel-close-btn" onClick={closeAllModals}>×</button>
          </div>

          <div className="countdown-top">
            <span className="countdown-text">⏱ 当前剩余: {countdown}s</span>
          </div>

          <div className="bid-item-card">
            <div className="bid-item-thumb">
              <img src={item.images[0]} alt={item.title} />
            </div>
            <div className="bid-item-info">
              <span className="panel-badge">限时竞拍</span>
              <h4 className="bid-item-title">{item.title}</h4>
              <div className="current-price-wrap">
                <span className="label">当前价</span>
                <span className="price">¥{item.currentPrice}</span>
              </div>
            </div>
          </div>

          {bidPanelMode === 'detail' ? (
            <div className="detail-panel-content">
              <h5 className="detail-title">拍品介绍</h5>
              <p className="detail-desc">{item.description}</p>
              <div className="detail-grid">
                <div className="detail-cell">
                  <span className="k">起拍价</span>
                  <span className="v">¥{item.startPrice}</span>
                </div>
                <div className="detail-cell">
                  <span className="k">加价步长</span>
                  <span className="v">¥{increment}</span>
                </div>
                <div className="detail-cell">
                  <span className="k">封顶价</span>
                  <span className="v">{item.maxPrice ? `¥${item.maxPrice}` : '无封顶'}</span>
                </div>
                <div className="detail-cell">
                  <span className="k">状态</span>
                  <span className="v">{item.status}</span>
                </div>
                <div className="detail-cell">
                  <span className="k">延时规则</span>
                  <span className="v">
                    最后{item.extensionTriggerSeconds}s内出价延长{item.extendedSeconds}s
                  </span>
                </div>
              </div>
            </div>
          ) : (
            <>
              <div className="my-bid-area">
                <span className="my-bid-label">我的出价</span>
                <div className="price-adjust-row">
                  <button className="adjust-btn" onClick={handleMinus}>-</button>
                  <span className="my-big-price">¥{biddingPrice}</span>
                  <button className="adjust-btn" onClick={handlePlus}>+</button>
                </div>
              </div>

              <p className="increment-tip">
                最低加价步长: {increment}元，当前最低可出: ¥{minAllowedPrice}
              </p>

              <Button
                className="submit-bid-btn"
                color="primary"
                block
                onClick={handleSubmit}
                disabled={isSubmitting}
              >
                {isSubmitting ? <Loading color="white" /> : '立即出价'}
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export default BidActionPanel
