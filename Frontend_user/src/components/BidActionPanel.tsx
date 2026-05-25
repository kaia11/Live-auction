import React, { useState } from 'react'
import { Button, Loading, Toast } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './BidActionPanel.scss'

const BidActionPanel: React.FC = () => {
  const { items, currentItemId, closeAllModals, submitBid, myBidStatus } = useLiveRoomStore()
  const item = items.find(i => i.id === currentItemId)
  const [biddingPrice, setBiddingPrice] = useState(item?.currentPrice || 0)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const increment = item?.minIncrement || 1000

  const handleMinus = () => {
    setBiddingPrice(prev => Math.max((item?.currentPrice || 0), prev - increment))
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
    setIsSubmitting(true)
    const success = await submitBid(biddingPrice)
    setIsSubmitting(false)
    if (success) {
      Toast.show('出价成功！')
    }
  }

  if (!item) return null

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="bid-panel" onClick={(e) => e.stopPropagation()}>
        <div className="bid-panel-inner">
          <div className="countdown-top">
            <span className="countdown-text">⏱ 当前剩余: 300s</span>
          </div>

          <div className="bid-item-card">
            <div className="bid-item-thumb" />
            <div className="bid-item-info">
              <h4 className="bid-item-title">{item.title}</h4>
              <div className="current-price-wrap">
                <span className="label">当前价</span>
                <span className="price">¥{item.currentPrice}</span>
              </div>
            </div>
          </div>

          <div className="my-bid-area">
            <span className="my-bid-label">我的出价</span>
            <div className="price-adjust-row">
              <button className="adjust-btn" onClick={handleMinus}>-</button>
              <span className="my-big-price">¥{biddingPrice}</span>
              <button className="adjust-btn" onClick={handlePlus}>+</button>
            </div>
          </div>

          <p className="increment-tip">最低加价幅度: {increment}元</p>

          <Button
            className="submit-bid-btn"
            color="primary"
            block
            onClick={handleSubmit}
            disabled={isSubmitting}
          >
            {isSubmitting ? <Loading color="white" /> : '立即出价'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default BidActionPanel
