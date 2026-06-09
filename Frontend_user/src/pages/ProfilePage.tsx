import React, { useEffect, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Image, Card } from 'antd-mobile'
import { useUserStore } from '../stores/useUserStore'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useAppStore } from '../stores/useAppStore'
import { computeProfileStats, getProfileBadge } from '../utils/profileStats'
import './ProfilePage.scss'

const ProfilePage: React.FC = () => {
  const navigate = useNavigate()
  const { user, logout, hydrateUser } = useUserStore()
  const { bidHistories, loadBidHistories } = useLiveRoomStore()
  const { currentTab, setCurrentTab, lastVisitedRoomId } = useAppStore()

  const profileStats = useMemo(() => computeProfileStats(bidHistories), [bidHistories])
  const profileBadge = useMemo(() => getProfileBadge(profileStats), [profileStats])

  const dedupedBidHistories = useMemo(() => {
    const byItem = new Map<string, typeof bidHistories[number]>()
    for (const history of bidHistories) {
      const existing = byItem.get(history.itemId)
      if (!existing) {
        byItem.set(history.itemId, history)
        continue
      }
      if (history.bidPrice > existing.bidPrice) {
        byItem.set(history.itemId, history)
        continue
      }
      if (history.bidPrice === existing.bidPrice) {
        const nextTime = new Date(history.bidTime).getTime()
        const oldTime = new Date(existing.bidTime).getTime()
        if (Number.isFinite(nextTime) && Number.isFinite(oldTime) && nextTime > oldTime) {
          byItem.set(history.itemId, history)
        }
      }
    }
    return Array.from(byItem.values()).sort((a, b) => {
      const ta = new Date(a.bidTime).getTime()
      const tb = new Date(b.bidTime).getTime()
      if (Number.isFinite(ta) && Number.isFinite(tb)) {
        return tb - ta
      }
      return b.bidPrice - a.bidPrice
    })
  }, [bidHistories])

  useEffect(() => {
    void hydrateUser()
    void loadBidHistories()
  }, [hydrateUser, loadBidHistories])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const goBack = () => {
    if (lastVisitedRoomId) {
      navigate(`/live/${lastVisitedRoomId}`)
    } else {
      navigate('/rooms')
    }
  }

  return (
    <div className="profile-page">
      <div className="scrollable-content">
        <div className="status-bar">
          <span className="time">15:25</span>
        </div>

        <div className="profile-top-nav">
          <span className="back-btn" onClick={goBack}>‹</span>
          <span className="page-title">个人主页</span>
          <span className="setting-text" onClick={handleLogout}>退出登录</span>
        </div>

        <div className="user-info-card">
          <Image className="user-avatar" src={user?.avatar || 'https://picsum.photos/avatar/150/150'} fit="cover" />
          <div className="user-text">
            <p className="user-badge">{profileBadge}</p>
            <h2 className="nickname">{user?.nickname || '未登录'}</h2>
            <p className="user-id-text">ID: {user?.id || '-'}</p>
          </div>
        </div>

        <div className="stats-row">
          <Card className="stat-card">
            <div className="stat-num">{profileStats.participatedSessions}</div>
            <div className="stat-label">参与场次</div>
          </Card>
          <Card className="stat-card">
            <div className="stat-num">{profileStats.successfulBids}</div>
            <div className="stat-label">成功竞拍</div>
          </Card>
          <Card className="stat-card">
            <div className="stat-num">{profileStats.totalBids}</div>
            <div className="stat-label">累计出价</div>
          </Card>
        </div>

        <div className="history-section-card">
          <div className="history-header-row">
            <h3 className="history-title">我的竞拍记录</h3>
          </div>
          {dedupedBidHistories.map((h) => (
            <div key={h.id} className="history-item-row">
              <Image className="history-thumb" src={h.itemImage} fit="cover" />
              <div className="history-info">
                <div className="history-item-title">{h.itemTitle}</div>
                <div className="history-time">{h.bidTime}</div>
              </div>
              <div className="history-result">
                <span className={`price-tag ${h.result}`}>¥{h.bidPrice}</span>
                <span className={`result-text ${h.result}`}>
                  {h.result === 'win' ? '已成交' : h.result === 'lose' ? '未中标' : '进行中'}
                </span>
              </div>
            </div>
          ))}
        </div>

        <button className="logout-btn" onClick={handleLogout}>退出登录</button>
        <div className="bottom-spacer" />
      </div>

      <div className="bottom-tab-bar">
        <div
          className={`tab-item ${currentTab === 'live' ? 'active' : ''}`}
          onClick={() => {
            setCurrentTab('live')
            if (lastVisitedRoomId) {
              navigate(`/live/${lastVisitedRoomId}`)
            } else {
              navigate('/rooms')
            }
          }}
        >
          <span className="tab-icon">📺</span>
          <span className="tab-text">直播</span>
        </div>
        <div className={`tab-item ${currentTab === 'profile' ? 'active' : ''}`}>
          <span className="tab-icon">👤</span>
          <span className="tab-text">我的</span>
        </div>
      </div>
    </div>
  )
}

export default ProfilePage
