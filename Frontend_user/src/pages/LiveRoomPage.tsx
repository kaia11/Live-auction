import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Toast } from 'antd-mobile'
import { motion } from 'framer-motion'
import {
  ChevronLeft,
  Gift,
  Heart,
  MessageCircle,
  Search,
  Share2,
  UserRound,
} from 'lucide-react'
import { USE_MOCK } from '../api/client'
import { createBid } from '../api/bids'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useLiveRoomUIStore } from '../stores/useLiveRoomUIStore'
import { useLiveRuntimeStore } from '../stores/useLiveRuntimeStore'
import { useAppStore } from '../stores/useAppStore'
import { useUserStore } from '../stores/useUserStore'
import { useRoomsQuery } from '../hooks/queries/useRoomsQuery'
import { useRoomRuntimeQuery } from '../hooks/queries/useRoomRuntimeQuery'
import { useBidHistoriesQuery } from '../hooks/queries/useBidHistoriesQuery'
import { useRoomCommentsQuery } from '../hooks/queries/useRoomCommentsQuery'
import { useLiveRoomRealtime } from '../hooks/useLiveRoomRealtime'
import CurrentAuctionCard from '../components/CurrentAuctionCard'
import AuctionItemDrawer from '../components/AuctionItemDrawer'
import BidActionPanel from '../components/BidActionPanel'
import BidSuccessModal from '../components/BidSuccessModal'
import OvertakenModal from '../components/OvertakenModal'
import AuctionEndPanel from '../components/AuctionEndPanel'
import DelayBanner from '../components/DelayBanner'
import { AuctionItem } from '../types'
import './LiveRoomPage.scss'

const ASSET_BASE = import.meta.env.BASE_URL
const LIVE_BG_VIDEO_SRC = `${ASSET_BASE}videos/live-bg.mp4?v=20260609`
const FALLBACK_IMAGE_SRC = `${ASSET_BASE}images/Image.png`

const LiveRoomPage: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>()
  const navigate = useNavigate()
  const {
    rooms,
    items,
    currentItemId,
    comments,
    onlineCount,
    top3Ranking,
    autoProxyFallbackMode,
    autoProxyLocalEnabled,
    autoProxyLocalMaxPrice,
    myBidStatus,
    syncRooms,
    syncRuntimeSnapshot,
    syncBidHistories,
    syncCommentsSnapshot,
    pollRoomEvents,
    submitComment,
  } = useLiveRoomStore()
  const {
    showAuctionItemDrawer,
    showBidPanel,
    showBidSuccessModal,
    showOvertakenModal,
    showAuctionEndPanel,
    showDelayBanner,
    toggleAuctionItemDrawer,
    openBidPanel,
    setCurrentAuctionCardClosed,
  } = useLiveRoomUIStore()
  const { setLastVisitedRoomId, setCurrentTab, currentTab } = useAppStore()
  const connectionState = useLiveRuntimeStore((state) => state.connectionState)
  const [commentDraft, setCommentDraft] = useState('')
  const [isFollowed, setIsFollowed] = useState(false)
  const [bgVideoError, setBgVideoError] = useState(false)
  const [isCardCollapsed, setIsCardCollapsed] = useState(false)
  const [collapsedItem, setCollapsedItem] = useState<AuctionItem | null>(null)
  const [collapsedPos, setCollapsedPos] = useState<{ left: number; top: number }>({ left: 0, top: 0 })
  const [collapsedReady, setCollapsedReady] = useState(false)
  const pageRef = React.useRef<HTMLDivElement | null>(null)
  const bgVideoRef = React.useRef<HTMLVideoElement | null>(null)
  const videoProgressRef = React.useRef<{ lastTime: number; stuckCount: number }>({ lastTime: 0, stuckCount: 0 })
  const autoBidRef = React.useRef<{ inFlight: boolean; lastAttemptAt: number }>({ inFlight: false, lastAttemptAt: 0 })
  const dragRef = React.useRef({
    dragging: false,
    moved: false,
    startX: 0,
    startY: 0,
    startLeft: 0,
    startTop: 0,
  })

  const getPageBounds = () => {
    const el = pageRef.current
    if (!el) {
      return { width: window.innerWidth, height: window.innerHeight }
    }
    return { width: el.clientWidth, height: el.clientHeight }
  }

  const clampTopWithinPage = (top: number) => {
    const { height } = getPageBounds()
    const maxTop = Math.max(72, height - 136)
    return Math.min(Math.max(72, top), maxTop)
  }

  const getSnappedLeft = (left: number) => {
    const { width } = getPageBounds()
    const leftEdge = 8
    const rightEdge = Math.max(8, width - 74)
    const toLeft = Math.abs(left - leftEdge)
    const toRight = Math.abs(rightEdge - left)
    return toLeft <= toRight ? leftEdge : rightEdge
  }

  const roomsQuery = useRoomsQuery()
  const roomRuntimeQuery = useRoomRuntimeQuery(roomId)
  const bidHistoriesQuery = useBidHistoriesQuery()
  const roomCommentsQuery = useRoomCommentsQuery(roomId)

  useLiveRoomRealtime(roomId)

  useEffect(() => {
    if (roomsQuery.data) {
      syncRooms(roomsQuery.data)
    }
  }, [roomsQuery.data, syncRooms])

  useEffect(() => {
    if (!roomId || !roomRuntimeQuery.data) {
      return
    }

    syncRuntimeSnapshot({
      currentRoomId: roomId,
      items: roomRuntimeQuery.data.items,
      currentItemId: roomRuntimeQuery.data.currentItemId,
      currentSessionId: roomRuntimeQuery.data.currentSessionId,
      top3Ranking: roomRuntimeQuery.data.top3Ranking,
      myBidStatus: roomRuntimeQuery.data.myBidStatus,
      currentCountdown: roomRuntimeQuery.data.currentCountdown,
      onlineCount: roomRuntimeQuery.data.onlineCount,
    })
  }, [roomId, roomRuntimeQuery.data, syncRuntimeSnapshot])

  useEffect(() => {
    if (bidHistoriesQuery.data) {
      syncBidHistories(bidHistoriesQuery.data)
    }
  }, [bidHistoriesQuery.data, syncBidHistories])

  useEffect(() => {
    if (!roomCommentsQuery.data) {
      return
    }

    syncCommentsSnapshot(
      roomCommentsQuery.data.comments,
      roomCommentsQuery.data.lastEventVersion,
    )
  }, [roomCommentsQuery.data, syncCommentsSnapshot])

  useEffect(() => {
    if (!roomId) {
      return
    }

    setLastVisitedRoomId(roomId)
  }, [roomId, setLastVisitedRoomId])

  useEffect(() => {
    if (!roomId) {
      return
    }

    if (!USE_MOCK && connectionState === 'connected') {
      return
    }

    const timer = window.setInterval(() => {
      void pollRoomEvents(roomId)
    }, USE_MOCK ? 3000 : 8000)

    return () => {
      window.clearInterval(timer)
    }
  }, [roomId, pollRoomEvents, connectionState])

  useEffect(() => {
    const video = bgVideoRef.current
    if (!video || bgVideoError) {
      return
    }

    const tryPlay = () => {
      video.defaultMuted = true
      void video.play().catch(() => {
        // Browsers may block autoplay on first paint; keep retrying on interaction/visibility.
      })
    }

    const onVisible = () => {
      if (document.visibilityState === 'visible') {
        tryPlay()
      }
    }

    tryPlay()
    const timer = window.setInterval(() => {
      const nowTime = video.currentTime
      if (!video.paused && Number.isFinite(nowTime)) {
        if (Math.abs(nowTime - videoProgressRef.current.lastTime) < 0.01) {
          videoProgressRef.current.stuckCount += 1
        } else {
          videoProgressRef.current.stuckCount = 0
        }
        videoProgressRef.current.lastTime = nowTime
      }
      if (video.paused && !video.ended) {
        tryPlay()
      }
      if (videoProgressRef.current.stuckCount >= 3) {
        videoProgressRef.current.stuckCount = 0
        video.load()
        tryPlay()
      }
    }, 1500)

    document.addEventListener('visibilitychange', onVisible)
    window.addEventListener('focus', tryPlay)
    return () => {
      window.clearInterval(timer)
      document.removeEventListener('visibilitychange', onVisible)
      window.removeEventListener('focus', tryPlay)
    }
  }, [bgVideoError])

  useEffect(() => {
    const smartEnabled = autoProxyFallbackMode
      ? autoProxyLocalEnabled
      : !!myBidStatus.autoProxyEnabled
    const smartMaxPrice = autoProxyFallbackMode
      ? autoProxyLocalMaxPrice
      : (myBidStatus.autoProxyMaxPrice ?? 0)

    if (!smartEnabled || smartMaxPrice <= 0) {
      return
    }

    const tick = () => {
      const store = useLiveRoomStore.getState()
      const userId = useUserStore.getState().user?.id
      if (!userId || !store.currentItemId) {
        return
      }
      const currentItemForAutoBid = store.items.find((item) => item.id === store.currentItemId)
      if (!currentItemForAutoBid || currentItemForAutoBid.status !== '竞拍中') {
        return
      }

      const enabled = store.autoProxyFallbackMode
        ? store.autoProxyLocalEnabled
        : !!store.myBidStatus.autoProxyEnabled
      const maxPrice = store.autoProxyFallbackMode
        ? store.autoProxyLocalMaxPrice
        : (store.myBidStatus.autoProxyMaxPrice ?? 0)
      if (!enabled || maxPrice <= 0) {
        return
      }

      const isLeadingByItem = currentItemForAutoBid.currentLeader === userId
      if (isLeadingByItem || store.myBidStatus.isLeading) {
        return
      }

      const nextBid = currentItemForAutoBid.currentPrice + (currentItemForAutoBid.minIncrement || 1)
      if (nextBid > maxPrice) {
        return
      }

      const now = Date.now()
      if (autoBidRef.current.inFlight || now - autoBidRef.current.lastAttemptAt < 1200) {
        return
      }

      autoBidRef.current.inFlight = true
      autoBidRef.current.lastAttemptAt = now
      void createBid({
        roomId: store.currentRoomId!,
        sessionId: store.currentSessionId!,
        itemId: store.currentItemId!,
        userId,
        bidPrice: nextBid,
        requestId: `proxy-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      })
        .then(() => store.loadRoomRuntime(store.currentRoomId!))
        .catch(() => {
          // Ignore transient conflicts; next tick will retry.
        })
        .finally(() => {
          autoBidRef.current.inFlight = false
        })
    }

    tick()
    const timer = window.setInterval(tick, 1200)
    return () => window.clearInterval(timer)
  }, [
    autoProxyFallbackMode,
    autoProxyLocalEnabled,
    autoProxyLocalMaxPrice,
    myBidStatus.autoProxyEnabled,
    myBidStatus.autoProxyMaxPrice,
  ])

  const currentItem = items.find((item) => item.id === currentItemId)
  const currentRoom = rooms.find((room) => room.id === roomId)
  const hasItems = items.length > 0
  // 仅在“当前拍品拍卖卡片”可见时上移弹幕
  const shouldShiftComments = !isCardCollapsed && !!currentItem
  const collapsedPreviewItem = collapsedItem ?? currentItem

  useEffect(() => {
    if (!currentRoom) {
      return
    }
    if (currentRoom.status !== 'living') {
      Toast.show('直播间未开播，已返回列表')
      navigate('/rooms')
    }
  }, [currentRoom, navigate])

  useEffect(() => {
    if (collapsedReady) {
      return
    }
    const { width, height } = getPageBounds()
    const left = Math.max(8, width - 74)
    const top = Math.max(72, height - 136)
    setCollapsedPos({ left, top })
    setCollapsedReady(true)
  }, [collapsedReady])

  useEffect(() => {
    const onResize = () => {
      const { width } = getPageBounds()
      setCollapsedPos((prev) => {
        const maxLeft = Math.max(8, width - 74)
        return {
          left: getSnappedLeft(Math.min(Math.max(8, prev.left), maxLeft)),
          top: clampTopWithinPage(prev.top),
        }
      })
    }
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const goToRooms = () => {
    navigate('/rooms')
  }

  const goToProfile = () => {
    setCurrentTab('profile')
    navigate('/profile')
  }

  const goBackLiveList = () => {
    navigate('/rooms')
  }

  const handleCloseBidCard = () => {
    if (currentItem) {
      setCollapsedItem(currentItem)
    }
    setIsCardCollapsed(true)
    setCurrentAuctionCardClosed(true)
  }

  const handleReopenBidCard = () => {
    if (dragRef.current.moved) {
      dragRef.current.moved = false
      return
    }
    setIsCardCollapsed(false)
    setCurrentAuctionCardClosed(false)
  }

  useEffect(() => {
    if (!isCardCollapsed && currentItem) {
      setCollapsedItem(currentItem)
    }
  }, [isCardCollapsed, currentItem])

  const handleCollapsedPointerDown = (event: React.PointerEvent<HTMLButtonElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId)
    dragRef.current = {
      dragging: true,
      moved: false,
      startX: event.clientX,
      startY: event.clientY,
      startLeft: collapsedPos.left,
      startTop: collapsedPos.top,
    }
  }

  const handleCollapsedPointerMove = (event: React.PointerEvent<HTMLButtonElement>) => {
    if (!dragRef.current.dragging) {
      return
    }
    const dx = event.clientX - dragRef.current.startX
    const dy = event.clientY - dragRef.current.startY
    if (Math.abs(dx) > 3 || Math.abs(dy) > 3) {
      dragRef.current.moved = true
    }
    const nextLeft = dragRef.current.startLeft + dx
    const nextTop = dragRef.current.startTop + dy
    const { width, height } = getPageBounds()
    const maxLeft = Math.max(8, width - 74)
    const maxTop = Math.max(72, height - 136)
    setCollapsedPos({
      left: Math.min(Math.max(8, nextLeft), maxLeft),
      top: Math.min(Math.max(72, nextTop), maxTop),
    })
  }

  const handleCollapsedPointerUp = () => {
    setCollapsedPos((prev) => ({
      left: getSnappedLeft(prev.left),
      top: clampTopWithinPage(prev.top),
    }))
    dragRef.current.dragging = false
  }

  const handleSubmitComment = async () => {
    const content = commentDraft.trim()
    if (!content) {
      return
    }

    const success = await submitComment(content)
    if (success) {
      setCommentDraft('')
      Toast.show('弹幕已发送')
      await roomCommentsQuery.refetch()
    }
  }

  return (
    <div className="live-room-page" ref={pageRef}>
      <div className="live-bg">
        {bgVideoError ? (
          <img
            className="bg-video bg-fallback-image"
            src={currentRoom?.coverImage || FALLBACK_IMAGE_SRC}
            alt="live background fallback"
          />
        ) : (
          <video
            ref={bgVideoRef}
            className="bg-video"
            src={LIVE_BG_VIDEO_SRC}
            autoPlay
            muted
            loop
            playsInline
            preload="auto"
            onCanPlay={() => {
              if (bgVideoRef.current) {
                void bgVideoRef.current.play().catch(() => {
                  // Ignore autoplay block and keep fallback behavior.
                })
              }
            }}
            onError={() => setBgVideoError(true)}
          />
        )}
        <div className="bg-overlay" />
      </div>

      <div className="top-area">
        <button className="exit-live-btn" onClick={goBackLiveList}>
          <ChevronLeft size={14} />
        </button>
        <div className="anchor-info-bar">
          <div className="anchor-avatar">
            {(currentRoom?.anchorName ?? '主').slice(0, 1)}
          </div>
          <div className="anchor-text">
            <span className="anchor-name">{currentRoom?.anchorName ?? '直播间主播'}</span>
            <span className="online-label">{onlineCount}人在线</span>
          </div>
          <Button
            size="small"
            className={`follow-btn ${isFollowed ? 'followed' : 'unfollowed'}`}
            onClick={() => setIsFollowed((prev) => !prev)}
          >
            {isFollowed ? '已关注' : '关注'}
          </Button>
        </div>
        <div className="top-right-actions">
          <button className="search-action icon-btn" onClick={goToRooms}>
            <Search size={16} />
          </button>
          <button className="gift-action icon-btn">
            <Gift size={16} />
          </button>
        </div>
      </div>

      <div className="top-ranking-panel">
        {top3Ranking.slice(0, 3).map((entry, index) => (
          <div key={`${entry.userId}-${index}`} className="top-ranking-item">
            <img src={entry.avatar} alt={entry.nickname} className="rank-avatar" />
            <span className="rank-nickname">{entry.nickname}</span>
            <span className="rank-price">¥{entry.price}</span>
          </div>
        ))}
      </div>

      {!isCardCollapsed && currentItem && (
        <CurrentAuctionCard
          item={currentItem}
          onClose={handleCloseBidCard}
          onOpenDetail={() => openBidPanel('detail')}
        />
      )}

      {isCardCollapsed && (
        <button
          className="collapsed-current-item"
          onClick={handleReopenBidCard}
          onPointerDown={handleCollapsedPointerDown}
          onPointerMove={handleCollapsedPointerMove}
          onPointerUp={handleCollapsedPointerUp}
          style={collapsedReady ? { left: `${collapsedPos.left}px`, top: `${collapsedPos.top}px` } : undefined}
        >
          {collapsedPreviewItem?.images?.[0] ? (
            <img src={collapsedPreviewItem.images[0]} alt={collapsedPreviewItem.title} />
          ) : (
            <div className="collapsed-placeholder">拍品</div>
          )}
          <span className="collapsed-dot">当前拍品</span>
        </button>
      )}

      <div className="right-side-actions">
        <div className="like-burst" aria-hidden>
          <span>❤</span>
          <span>❤</span>
          <span>❤</span>
        </div>
        <motion.div className="action-btn" whileTap={{ scale: 0.92 }}>
          <Heart size={18} />
          <small>喜欢</small>
        </motion.div>
        <motion.div className="action-btn" whileTap={{ scale: 0.92 }}>
          <MessageCircle size={18} />
          <small>讨论</small>
        </motion.div>
        <motion.div className="action-btn" whileTap={{ scale: 0.92 }}>
          <Share2 size={18} />
          <small>分享</small>
        </motion.div>
        <motion.div className="action-btn" onClick={goToProfile} whileTap={{ scale: 0.92 }}>
          <UserRound size={18} />
        </motion.div>
      </div>

      <div className={`comment-stream ${shouldShiftComments ? 'with-panel' : ''}`}>
        {comments.slice(-4).map((comment, index) => (
          <div key={`${comment.userId}-${index}-${comment.content}`} className="comment-bubble">
            <span className="comment-name">{comment.nickname}</span>
            <span className="comment-text">{comment.content}</span>
          </div>
        ))}
      </div>

      <div className="bottom-function-area">
        <div className="comment-input-shell">
          <input
            className="comment-input"
            placeholder="说点什么..."
            value={commentDraft}
            onChange={(e) => setCommentDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                void handleSubmitComment()
              }
            }}
          />
        </div>
        <div className="func-buttons-row">
          <button className="func-btn send-btn" onClick={() => void handleSubmitComment()}>发送</button>
          {hasItems ? (
            <span className="func-btn" onClick={toggleAuctionItemDrawer}>拍品列表</span>
          ) : null}
        </div>
      </div>

      <div className="bottom-tab-bar">
        <div
          className={`tab-item ${currentTab === 'live' ? 'active' : ''}`}
          onClick={() => { setCurrentTab('live'); navigate('/rooms') }}
        >
          <span className="tab-icon">📺</span>
          <span className="tab-text">直播</span>
        </div>
        <div
          className={`tab-item ${currentTab === 'profile' ? 'active' : ''}`}
          onClick={goToProfile}
        >
          <span className="tab-icon">👤</span>
          <span className="tab-text">我的</span>
        </div>
      </div>

      <div className="portals-container">
        {showDelayBanner && <DelayBanner />}
        {showAuctionItemDrawer && <AuctionItemDrawer />}
        {showBidPanel && <BidActionPanel />}
        {showBidSuccessModal && <BidSuccessModal />}
        {showOvertakenModal && <OvertakenModal />}
        {showAuctionEndPanel && <AuctionEndPanel />}
      </div>
    </div>
  )
}

export default LiveRoomPage
