import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Toast } from 'antd-mobile'
import { USE_MOCK } from '../api/client'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useLiveRoomUIStore } from '../stores/useLiveRoomUIStore'
import { useLiveRuntimeStore } from '../stores/useLiveRuntimeStore'
import { useAppStore } from '../stores/useAppStore'
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
  const [isCardCollapsed, setIsCardCollapsed] = useState(false)
  const [collapsedItem, setCollapsedItem] = useState<AuctionItem | null>(null)
  const [collapsedPos, setCollapsedPos] = useState<{ left: number; top: number }>({ left: 0, top: 0 })
  const [collapsedReady, setCollapsedReady] = useState(false)
  const pageRef = React.useRef<HTMLDivElement | null>(null)
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

  const currentItem = items.find((item) => item.id === currentItemId)
  const currentRoom = rooms.find((room) => room.id === roomId)
  // 仅在“当前拍品拍卖卡片”可见时上移弹幕
  const shouldShiftComments = !isCardCollapsed && !!currentItem
  const collapsedPreviewItem = collapsedItem ?? currentItem

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
    if (window.history.length > 1) {
      navigate(-1)
      return
    }
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
        <video
          className="bg-video"
          src="/videos/live-bg.mp4"
          autoPlay
          muted
          loop
          playsInline
          preload="auto"
        />
        <div className="bg-overlay" />
      </div>

      <div className="top-area">
        <button className="exit-live-btn" onClick={goBackLiveList}>
          退出
        </button>
        <div className="anchor-info-bar">
          <div className="anchor-avatar">
            {(currentRoom?.anchorName ?? '主').slice(0, 1)}
          </div>
          <div className="anchor-text">
            <span className="anchor-name">{currentRoom?.anchorName ?? '直播间主播'}</span>
            <span className="online-label">{onlineCount}人在线</span>
          </div>
          <span className={`realtime-pill state-${connectionState}`}>
            {USE_MOCK ? 'Mock' : connectionState}
          </span>
          <Button
            size="small"
            color="primary"
            className={`follow-btn ${isFollowed ? 'followed' : 'unfollowed'}`}
            onClick={() => setIsFollowed((prev) => !prev)}
          >
            {isFollowed ? '已关注' : '关注'}
          </Button>
        </div>
        <div className="top-right-actions">
          <span className="search-action" onClick={goToRooms}>🔍</span>
          <span className="gift-action">🎁</span>
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
        <div className="action-btn"><span>❤️</span><small>喜欢</small></div>
        <div className="action-btn"><span>💬</span><small>讨论</small></div>
        <div className="action-btn"><span>↗️</span><small>分享</small></div>
        <div className="action-btn" onClick={goToProfile}>👤</div>
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
          <span className="func-btn" onClick={toggleAuctionItemDrawer}>拍品列表</span>
          <span className="func-btn">智能出价</span>
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
