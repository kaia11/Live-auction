import React, { useEffect, useMemo, useState } from 'react'
import { Button, Loading, Toast } from 'antd-mobile'
import { AxiosError } from 'axios'
import { AuctionItem } from '../types'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useUserStore } from '../stores/useUserStore'
import DepositPaymentModal from './DepositPaymentModal'
import SmartBidModal from './SmartBidModal'
import { hasPaidDeposit, markDepositPaid } from '../utils/deposit'
import './CurrentAuctionCard.scss'

interface Props {
  item: AuctionItem
  onClose: () => void
  onOpenDetail: () => void
}

const CurrentAuctionCard: React.FC<Props> = ({ item, onClose, onOpenDetail }) => {
  const {
    myBidStatus,
    submitBid,
    configureAutoProxy,
    autoProxyFallbackMode,
    autoProxyLocalEnabled,
    autoProxyLocalMaxPrice,
  } = useLiveRoomStore()
  const currentUserId = useUserStore((state) => state.user?.id)
  const [countdown, setCountdown] = useState(0)
  const [biddingPrice, setBiddingPrice] = useState(0)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [pendingBidPrice, setPendingBidPrice] = useState<number | null>(null)
  const [showDepositModal, setShowDepositModal] = useState(false)
  const [showSmartBidModal, setShowSmartBidModal] = useState(false)
  const [smartBidPrice, setSmartBidPrice] = useState(0)

  const increment = item.minIncrement || 1
  const basePrice = Math.max(item.currentPrice, item.startPrice)
  const minAllowedPrice = basePrice + increment
  const canBid = item.status === '竞拍中'
  const smartBidEnabled = autoProxyFallbackMode ? autoProxyLocalEnabled : !!myBidStatus.autoProxyEnabled
  const smartBidMaxPrice = autoProxyFallbackMode
    ? autoProxyLocalMaxPrice
    : (myBidStatus.autoProxyMaxPrice ?? 0)

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

  useEffect(() => {
    const next = Math.max(minAllowedPrice, smartBidMaxPrice)
    setSmartBidPrice(next)
  }, [minAllowedPrice, smartBidMaxPrice])

  const handleMinus = () => {
    if (!canBid) return
    setBiddingPrice((prev) => Math.max(minAllowedPrice, prev - increment))
  }

  const handlePlus = () => {
    if (!canBid) return
    if (item.maxPrice && biddingPrice + increment > item.maxPrice) {
      Toast.show('已达到封顶价')
      return
    }
    setBiddingPrice((prev) => prev + increment)
  }

  const submitBidRequest = async (targetBidPrice: number) => {
    setIsSubmitting(true)
    try {
      await submitBid(targetBidPrice)
      Toast.show('出价成功')
    } catch (error) {
      const err = error as AxiosError<{ message?: string }>
      const message = err.response?.data?.message ?? ''
      if (message.includes('session is not bidding')) {
        Toast.show('当前场次未开拍')
      } else if (message.includes('already leading bid')) {
        Toast.show('您已经是最高价')
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

  const handleQuickBid = async () => {
    if (!canBid) {
      Toast.show(`当前拍品${item.status}，无法出价`)
      return
    }
    if (biddingPrice < minAllowedPrice) {
      Toast.show(`最低可出价 ¥${minAllowedPrice}`)
      return
    }

    if (item.depositAmount > 0 && currentUserId && !hasPaidDeposit(currentUserId, item.id)) {
      setPendingBidPrice(biddingPrice)
      setShowDepositModal(true)
      return
    }

    await submitBidRequest(biddingPrice)
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
              {canBid ? `⏱ ${countdown}s` : `⏱ ${item.status}`}
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
          <button className="inline-adjust-btn" onClick={handleMinus} disabled={!canBid}>-</button>
          <span className="inline-bid-price">¥{biddingPrice}</span>
          <button className="inline-adjust-btn" onClick={handlePlus} disabled={!canBid}>+</button>
        </div>
      </div>
      <Button className={`bid-btn ${!canBid ? 'is-disabled' : ''}`} color="primary" onClick={handleQuickBid} disabled={isSubmitting || !canBid}>
        {isSubmitting ? <Loading color="white" /> : (canBid ? '立即出价' : '暂不可出价')}
      </Button>
      <Button
        className={`smart-bid-inline-btn ${!canBid ? 'is-disabled' : ''}`}
        onClick={() => setShowSmartBidModal(true)}
        disabled={!canBid}
      >
        {canBid ? '智能出价' : '暂不可智能出价'}
      </Button>
      <p className="inline-bid-hint">
        当前最低可出 ¥{minAllowedPrice}（最小加价额度 ¥{increment}）
      </p>
      {showDepositModal && (
        <DepositPaymentModal
          amount={item.depositAmount}
          onCancel={() => {
            setShowDepositModal(false)
            setPendingBidPrice(null)
          }}
          onConfirm={() => {
            if (currentUserId) {
              markDepositPaid(currentUserId, item.id)
            }
            const nextBid = pendingBidPrice ?? biddingPrice
            setShowDepositModal(false)
            setPendingBidPrice(null)
            void submitBidRequest(nextBid)
          }}
        />
      )}
      {showSmartBidModal && (
        <SmartBidModal
          maxPrice={smartBidPrice}
          increment={increment}
          minAllowedPrice={minAllowedPrice}
          enabled={smartBidEnabled}
          onMinus={() => setSmartBidPrice((prev) => Math.max(minAllowedPrice, prev - increment))}
          onPlus={() => setSmartBidPrice((prev) => prev + increment)}
          onEnable={() => {
            void configureAutoProxy(smartBidPrice, true)
              .then((ok) => {
                if (ok) {
                  Toast.show('智能出价已开启')
                  setShowSmartBidModal(false)
                }
              })
              .catch((error) => {
                const err = error as AxiosError<{ message?: string }>
                Toast.show(err.response?.data?.message || err.message || '智能出价设置失败')
              })
          }}
          onDisable={() => {
            void configureAutoProxy(0, false)
              .then((ok) => {
                if (ok) {
                  Toast.show('智能出价已关闭')
                  setShowSmartBidModal(false)
                }
              })
              .catch((error) => {
                const err = error as AxiosError<{ message?: string }>
                Toast.show(err.response?.data?.message || err.message || '关闭智能出价失败')
              })
          }}
          onClose={() => setShowSmartBidModal(false)}
        />
      )}
    </div>
  )
}

export default CurrentAuctionCard
