import React, { useEffect, useMemo, useState } from 'react'
import { Button, Loading, Toast } from 'antd-mobile'
import { AxiosError } from 'axios'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useUserStore } from '../stores/useUserStore'
import DepositPaymentModal from './DepositPaymentModal'
import SmartBidModal from './SmartBidModal'
import { markDepositPaid, needsDepositPayment } from '../utils/deposit'
import './BidActionPanel.scss'

const BidActionPanel: React.FC = () => {
  const {
    items,
    currentItemId,
    closeAllModals,
    submitBid,
    bidPanelMode,
    openBidPanel,
    configureAutoProxy,
    myBidStatus,
    autoProxyFallbackMode,
    autoProxyLocalEnabled,
    autoProxyLocalMaxPrice,
  } = useLiveRoomStore()
  const currentUserId = useUserStore((state) => state.user?.id)
  const item = items.find(i => i.id === currentItemId)
  const [biddingPrice, setBiddingPrice] = useState(0)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [pendingBidPrice, setPendingBidPrice] = useState<number | null>(null)
  const [pendingDepositAction, setPendingDepositAction] = useState<'bid' | 'smartBid' | null>(null)
  const [showDepositModal, setShowDepositModal] = useState(false)
  const [showSmartBidModal, setShowSmartBidModal] = useState(false)
  const [smartBidPrice, setSmartBidPrice] = useState(0)

  const increment = item?.minIncrement || 1000
  const canBid = item?.status === '竞拍中'
  const smartBidEnabled = autoProxyFallbackMode ? autoProxyLocalEnabled : !!myBidStatus.autoProxyEnabled
  const smartBidMaxPrice = autoProxyFallbackMode
    ? autoProxyLocalMaxPrice
    : (myBidStatus.autoProxyMaxPrice ?? 0)
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
    if (!item) {
      return
    }
    const next = Math.max(minAllowedPrice, smartBidMaxPrice)
    setSmartBidPrice(next)
  }, [item, minAllowedPrice, smartBidMaxPrice])

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
    if (!canBid) return
    setBiddingPrice(prev => Math.max(minAllowedPrice, prev - increment))
  }

  const handlePlus = () => {
    if (!canBid) return
    const newPrice = biddingPrice + increment
    if (item?.maxPrice && newPrice > item.maxPrice) {
      Toast.show('已达到封顶价')
      return
    }
    setBiddingPrice(newPrice)
  }

  const submitBidRequest = async (targetBidPrice: number) => {
    setIsSubmitting(true)
    try {
      const outcome = await submitBid(targetBidPrice)
      if (outcome === 'leading') {
        Toast.show('出价成功！')
      }
    } catch (error) {
      const err = error as AxiosError<{ message?: string; code?: number }>
      const message = err.response?.data?.message ?? ''
      if (message.includes('session is not bidding')) {
        Toast.show('当前场次未在竞拍中')
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

  const handleSubmit = async () => {
    if (!item) return
    if (!canBid) {
      Toast.show(`当前拍品${item.status}，无法出价`)
      return
    }
    if (biddingPrice < minAllowedPrice) {
      Toast.show(`当前最低可出价 ¥${minAllowedPrice}`)
      return
    }

    if (needsDepositPayment(currentUserId, item.id, item.depositAmount)) {
      setPendingDepositAction('bid')
      setPendingBidPrice(biddingPrice)
      setShowDepositModal(true)
      return
    }

    await submitBidRequest(biddingPrice)
  }

  const openSmartBidModal = () => {
    if (!item) {
      return
    }
    if (needsDepositPayment(currentUserId, item.id, item.depositAmount)) {
      setPendingDepositAction('smartBid')
      setShowDepositModal(true)
      return
    }
    setShowSmartBidModal(true)
  }

  const handleDepositConfirm = () => {
    if (!currentUserId || !item) {
      return
    }
    markDepositPaid(currentUserId, item.id)
    const action = pendingDepositAction
    setShowDepositModal(false)
    setPendingDepositAction(null)

    if (action === 'smartBid') {
      setShowSmartBidModal(true)
      return
    }

    const nextBid = pendingBidPrice ?? biddingPrice
    setPendingBidPrice(null)
    void submitBidRequest(nextBid)
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
            <span className="countdown-text">
              {canBid ? `⏱ 当前剩余: ${countdown}s` : `⏱ 当前状态: ${item.status}`}
            </span>
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
                  <span className="k">最小加价额度</span>
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
                  <button className="adjust-btn" onClick={handleMinus} disabled={!canBid}>-</button>
                  <span className="my-big-price">¥{biddingPrice}</span>
                  <button className="adjust-btn" onClick={handlePlus} disabled={!canBid}>+</button>
                </div>
              </div>

              <p className="increment-tip">
                最小加价额度: {increment}元，当前最低可出: ¥{minAllowedPrice}
              </p>

              <Button
                className={`submit-bid-btn ${!canBid ? 'is-disabled' : ''}`}
                color="primary"
                block
                onClick={handleSubmit}
                disabled={isSubmitting || !canBid}
              >
                {isSubmitting ? <Loading color="white" /> : (canBid ? '立即出价' : '暂不可出价')}
              </Button>
              <Button
                className={`smart-bid-entry-btn ${!canBid ? 'is-disabled' : ''}`}
                block
                disabled={!canBid}
                onClick={openSmartBidModal}
              >
                {canBid ? '智能出价' : '暂不可智能出价'}
              </Button>
            </>
          )}
          {showDepositModal && (
            <DepositPaymentModal
              amount={item.depositAmount}
              onCancel={() => {
                setShowDepositModal(false)
                setPendingBidPrice(null)
                setPendingDepositAction(null)
              }}
              onConfirm={handleDepositConfirm}
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
      </div>
    </div>
  )
}

export default BidActionPanel
