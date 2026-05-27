import React, { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { SearchBar, Tabs, Card, Image } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useAppStore } from '../stores/useAppStore'
import './LiveRoomsPage.scss'

const LiveRoomsPage: React.FC = () => {
  const navigate = useNavigate()
  const { rooms, loadRooms, setCurrentRoomId } = useLiveRoomStore()
  const { setCurrentTab } = useAppStore()

  useEffect(() => {
    void loadRooms()
  }, [loadRooms])

  const enterRoom = (roomId: string) => {
    setCurrentRoomId(roomId)
    navigate(`/live/${roomId}`)
  }

  return (
    <div className="live-rooms-page">
      <div className="scrollable-content">
        <div className="status-bar">
          <span className="time">15:25</span>
        </div>

        <div className="page-hero">
          <p className="hero-kicker">LIVE AUCTION</p>
          <h1 className="hero-title">今晚值得蹲守的珠宝专场</h1>
          <p className="hero-subtitle">从直播间直接出价，实时查看排行和拍卖进度。</p>
        </div>

        <div className="top-search-section">
          <span className="back-btn">‹</span>
          <div className="search-bar-wrapper">
            <SearchBar placeholder="搜索直播间" className="custom-search" />
          </div>
        </div>

        <Tabs className="custom-tabs">
          <Tabs.Tab title="全部" key="all" />
          <Tabs.Tab title="翡翠" key="jade" />
          <Tabs.Tab title="和田玉" key="hetian" />
          <Tabs.Tab title="钻石" key="diamond" />
          <Tabs.Tab title="彩宝" key="gem" />
          <Tabs.Tab title="珍珠" key="pearl" />
        </Tabs>

        <div className="rooms-list">
          {rooms.map((room) => (
            <div
              key={room.id}
              className="room-card-wrapper"
              onClick={() => enterRoom(room.id)}
            >
              <Card className="room-card">
                <div className="room-cover">
                  <Image src={room.coverImage} fit="cover" />
                  <div className="live-badge">直播中</div>
                  {room.thumbnail && (
                    <div className="current-item-thumb">
                      <Image src={room.thumbnail} fit="cover" />
                    </div>
                  )}
                </div>
                <div className="room-info">
                  <div className="room-chip-row">
                    <span className="room-chip">精选专场</span>
                    <span className="room-chip subtle">{room.status === 'living' ? '进行中' : '已下播'}</span>
                  </div>
                  <h3 className="room-title">{room.title}</h3>
                  <div className="room-meta">
                    <span className="anchor-name">{room.anchorName}</span>
                    <span className="online-count">{room.onlineCount}人在看</span>
                  </div>
                </div>
              </Card>
            </div>
          ))}
        </div>
        <div className="bottom-spacer" />
      </div>

      <div className="bottom-tab-bar">
        <div
          className="tab-item active"
          onClick={() => setCurrentTab('live')}
        >
          <span className="tab-icon">📺</span>
          <span className="tab-text">直播</span>
        </div>
        <div
          className="tab-item"
          onClick={() => {
            setCurrentTab('profile')
            navigate('/profile')
          }}
        >
          <span className="tab-icon">👤</span>
          <span className="tab-text">我的</span>
        </div>
      </div>
    </div>
  )
}

export default LiveRoomsPage
