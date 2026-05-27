import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Toast } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useAppStore } from '../stores/useAppStore'
import CurrentAuctionCard from '../components/CurrentAuctionCard'
import AuctionItemDrawer from '../components/AuctionItemDrawer'
import BidActionPanel from '../components/BidActionPanel'
import RuleModal from '../components/RuleModal'
import BidSuccessModal from '../components/BidSuccessModal'
import OvertakenModal from '../components/OvertakenModal'
import AuctionEndPanel from '../components/AuctionEndPanel'
import DelayBanner from '../components/DelayBanner'
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
    isCurrentAuctionCardClosed,
    showAuctionItemDrawer,
    showBidPanel,
    showRuleModal,
    showBidSuccessModal,
    showOvertakenModal,
    showAuctionEndPanel,
    showDelayBanner,
    loadRooms,
    loadRoomRuntime,
    loadBidHistories,
    loadRoomComments,
    pollRoomEvents,
    submitComment,
    toggleAuctionItemDrawer,
    setCurrentAuctionCardClosed,
  } = useLiveRoomStore()
  const { setLastVisitedRoomId, setCurrentTab, currentTab } = useAppStore()
  const [commentDraft, setCommentDraft] = useState('')

  useEffect(() => {
    if (!roomId) {
      return
    }

    const hydrateRoom = async () => {
      await loadRooms()
      await loadRoomRuntime(roomId)
      await loadBidHistories()
      await loadRoomComments(roomId)
      setLastVisitedRoomId(roomId)
    }

    void hydrateRoom()
  }, [roomId, loadRooms, loadRoomRuntime, loadBidHistories, loadRoomComments, setLastVisitedRoomId])

  useEffect(() => {
    if (!roomId) {
      return
    }

    const timer = window.setInterval(() => {
      void pollRoomEvents(roomId)
    }, 3000)

    return () => {
      window.clearInterval(timer)
    }
  }, [roomId, pollRoomEvents])

  const currentItem = items.find(i => i.id === currentItemId)
  const currentRoom = rooms.find((room) => room.id === roomId)

  const goToRooms = () => {
    navigate('/rooms')
  }

  const goToProfile = () => {
    setCurrentTab('profile')
    navigate('/profile')
  }

  const handleCloseBidCard = () => {
    setCurrentAuctionCardClosed(true)
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
    }
  }

  return (
    <div className="live-room-page">
      <div className="live-bg">
        <img
          src={currentRoom?.coverImage ?? 'https://picsum.photos/livebg/1080/1920'}
          alt="live background"
          className="bg-image"
        />
        <div className="bg-overlay" />
      </div>

      <div className="top-area">
        <div className="anchor-info-bar">
          <div className="anchor-avatar">
            {(currentRoom?.anchorName ?? '主').slice(0, 1)}
          </div>
          <div className="anchor-text">
            <span className="anchor-name">{currentRoom?.anchorName ?? '直播间主播'}</span>
            <span className="online-label">{onlineCount}人在线</span>
          </div>
          <Button size="small" color="primary" className="follow-btn">关注</Button>
        </div>
        <div className="top-right-actions">
          <span className="search-action" onClick={goToRooms}>🔍</span>
          <span className="gift-action">🎁</span>
        </div>
      </div>

      {!isCurrentAuctionCardClosed && currentItem && (
        <CurrentAuctionCard
          item={currentItem}
          onClose={handleCloseBidCard}
        />
      )}

      <div className="right-side-actions">
        <div className="action-btn"><span>❤️</span><small>喜欢</small></div>
        <div className="action-btn"><span>💬</span><small>讨论</small></div>
        <div className="action-btn"><span>↗️</span><small>分享</small></div>
        <div className="action-btn" onClick={goToProfile}>👤</div>
      </div>

      <div className="comment-stream">
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
          <span className="func-btn" onClick={() => {
            const store = useLiveRoomStore.getState()
            store.toggleRuleModal()
          }}>规则</span>
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
        {showRuleModal && <RuleModal />}
        {showBidSuccessModal && <BidSuccessModal />}
        {showOvertakenModal && <OvertakenModal />}
        {showAuctionEndPanel && <AuctionEndPanel />}
      </div>
    </div>
  )
}

export default LiveRoomPage
