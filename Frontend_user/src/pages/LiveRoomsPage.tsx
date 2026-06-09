import React, { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { SearchBar, Tabs, Card, Image, Toast } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useAppStore } from '../stores/useAppStore'
import { useRoomsQuery } from '../hooks/queries/useRoomsQuery'
import './LiveRoomsPage.scss'

const ASSET_BASE = import.meta.env.BASE_URL
const LIVE_BG_VIDEO_SRC = `${ASSET_BASE}videos/live-bg.mp4?v=20260609`

const LiveRoomsPage: React.FC = () => {
  const navigate = useNavigate()
  const { rooms, syncRooms, setCurrentRoomId } = useLiveRoomStore()
  const { setCurrentTab } = useAppStore()
  const roomsQuery = useRoomsQuery()

  useEffect(() => {
    if (roomsQuery.data) {
      syncRooms(roomsQuery.data)
    }
  }, [roomsQuery.data, syncRooms])

  const enterRoom = (roomId: string) => {
    const room = rooms.find((entry) => entry.id === roomId)
    if (!room || room.status !== 'living') {
      Toast.show('直播间未开播，暂不可进入')
      return
    }
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
              className={`room-card-wrapper ${room.status !== 'living' ? 'disabled' : ''}`}
              onClick={() => enterRoom(room.id)}
            >
              <Card className="room-card">
                <div className="room-cover">
                  {room.status === 'living' ? (
                    <video
                      className="room-cover-video"
                      src={LIVE_BG_VIDEO_SRC}
                      autoPlay
                      muted
                      loop
                      playsInline
                      preload="auto"
                      onCanPlay={(event) => {
                        event.currentTarget.defaultMuted = true
                        void event.currentTarget.play().catch(() => {
                          // Keep poster fallback and retry on next render/user interaction.
                        })
                      }}
                    />
                  ) : (
                    <img className="room-cover-image" src={room.coverImage} alt={room.title} />
                  )}
                  <div className="live-badge">{room.status === 'living' ? '直播中' : '未开播'}</div>
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
