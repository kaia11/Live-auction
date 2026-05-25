import React, { useEffect } from 'react'
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
    items,
    currentItemId,
    onlineCount,
    isCurrentAuctionCardClosed,
    showAuctionItemDrawer,
    showBidPanel,
    showRuleModal,
    showBidSuccessModal,
    showOvertakenModal,
    showAuctionEndPanel,
    showDelayBanner,
    loadMockData,
    toggleAuctionItemDrawer,
    closeAllModals,
    setCurrentAuctionCardClosed,
  } = useLiveRoomStore()
  const { setLastVisitedRoomId, setCurrentTab, currentTab } = useAppStore()

  useEffect(() => {
    loadMockData()
    if (roomId) {
      setLastVisitedRoomId(roomId)
    }
  }, [roomId, loadMockData, setLastVisitedRoomId])

  const currentItem = items.find(i => i.id === currentItemId)

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

  return (
    <div className="live-room-page">
      <div className="live-bg">
        <img
          src="https://picsum.photos/livebg/1080/1920"
          alt="live background"
          className="bg-image"
        />
        <div className="bg-overlay" />
      </div>

      <div className="top-area">
        <div className="anchor-info-bar">
          <div className="anchor-avatar" />
          <div className="anchor-text">
            <span className="anchor-name">珠宝大师阿静</span>
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
        <div className="action-btn">❤️</div>
        <div className="action-btn">💬</div>
        <div className="action-btn">↗️</div>
        <div className="action-btn" onClick={goToProfile}>👤</div>
      </div>

      <div className="bottom-function-area">
        <div className="comment-input-placeholder">说点什么...</div>
        <div className="func-buttons-row">
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
