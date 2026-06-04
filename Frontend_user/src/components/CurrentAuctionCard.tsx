import React, { useEffect, useMemo, useState } from 'react'
import { Button, Loading, Toast } from 'antd-mobile'
import { AxiosError } from 'axios'
import { AuctionItem } from '../types'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './CurrentAuctionCard.scss'

interface Props {
  item: AuctionItem
  onClose: () => void
  onOpenDetail: () => void
}

const CurrentAuctionCard: React.FC<Props> = ({ item, onClose, onOpenDetail }) => {
  const { myBidStatus, submitBid } = useLiveRoomStore()
  const [countdown, setCountdown] = useState(0)
  const [biddingPrice, setBiddingPrice] = useState(0)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const increment = item.minIncrement || 1
  const basePrice = Math.max(item.currentPrice, item.startPrice)
  const minAllowedPrice = basePrice + increment

  const countdownTargetMs = useMemo(() => {
    if (!item.endTime) {
      return null
    }
    const ts = new Date(item.endTime).getTime()
    return Number.isFinite(ts) ? ts : null
  }, [item.endTime])

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

  useEffect(() => {
    setBiddingPrice(minAllowedPrice)
  }, [minAllowedPrice])

  const handleMinus = () => {
    setBiddingPrice((prev) => Math.max(minAllowedPrice, prev - increment))
  }

  const handlePlus = () => {
    if (item.maxPrice && biddingPrice + increment > item.maxPrice) {
      Toast.show('已达到封顶价')
      return
    }
    setBiddingPrice((prev) => prev + increment)
  }

  const handleQuickBid = async () => {
    if (item.status !== '竞拍中') {
      Toast.show('当前拍品未开拍，无法出价')
      return
    }
    if (biddingPrice < minAllowedPrice) {
      Toast.show(`最低可出价 ¥${minAllowedPrice}`)
      return
    }

    setIsSubmitting(true)
    try {
      await submitBid(biddingPrice)
      Toast.show('出价成功')
    } catch (error) {
      const err = error as AxiosError<{ message?: string }>
      const message = err.response?.data?.message ?? ''
      if (message.includes('session is not bidding')) {
        Toast.show('当前场次未开拍')
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

  return (
    <div className="current-auction-card">
      <div className="card-top-row">
        <div className="item-thumb">
          <img src={item.images[0]} alt={item.title} />
        </div>
        <div className="item-info-text">
          <span className="item-tag">当前拍品</span>
          <h4 className="item-title">{item.title}</h4>
          <div className="price-countdown-row">
            <span className="current-price">¥{item.currentPrice}</span>
            <span className="countdown-text">
              {item.status === '竞拍中' ? `⏱ ${countdown}s` : '⏱ 未开拍'}
            </span>
          </div>
        </div>
        <span className="close-btn" onClick={onClose}>×</span>
      </div>
      <div className="my-status-inline">
        <span>我的名次: {myBidStatus.myRank || '未上榜'}</span>
        <span>我的出价: ¥{myBidStatus.myHighestPrice}</span>
      </div>

      <div className="card-bottom-btns">
        <Button size="small" className="detail-btn" onClick={onOpenDetail}>详情</Button>
        <div className="inline-bid-controls">
          <button className="inline-adjust-btn" onClick={handleMinus}>-</button>
          <span className="inline-bid-price">¥{biddingPrice}</span>
          <button className="inline-adjust-btn" onClick={handlePlus}>+</button>
        </div>
      </div>
      <Button className="bid-btn" color="primary" onClick={handleQuickBid} disabled={isSubmitting}>
        {isSubmitting ? <Loading color="white" /> : '立即出价'}
      </Button>
      <p className="inline-bid-hint">
        当前最低可出 ¥{minAllowedPrice}（步长 ¥{increment}）
      </p>
    </div>
  )
}

export default CurrentAuctionCard
