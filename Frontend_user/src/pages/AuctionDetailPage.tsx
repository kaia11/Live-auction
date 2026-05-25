import React from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Image, Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './AuctionDetailPage.scss'

const AuctionDetailPage: React.FC = () => {
  const { itemId } = useParams<{ itemId: string }>()
  const navigate = useNavigate()
  const { items, toggleBidPanel } = useLiveRoomStore()
  const item = items.find(i => i.id === itemId)

  if (!item) {
    return <div className="auction-detail-page">拍品不存在</div>
  }

  return (
    <div className="auction-detail-page">
      <div className="scrollable-content">
        <div className="top-nav-bar">
          <span className="back-btn" onClick={() => navigate(-1)}>‹</span>
          <span className="anchor-name">{item.title.slice(0, 8)}</span>
          <div className="nav-actions">
            <span>关注</span>
            <span>分享</span>
          </div>
        </div>

        <div className="image-swiper-area">
          <Image src={item.images[0]} fit="cover" className="main-image" />
        </div>

        <div className="detail-content">
          <div className="price-info-row">
            <div className="price-col">
              <span className="label">起拍价</span>
              <span className="val">¥{item.startPrice}</span>
            </div>
            <div className="price-col">
              <span className="label">加价幅度</span>
              <span className="val">¥{item.minIncrement}</span>
            </div>
            <div className="price-col">
              <span className="label">封顶价</span>
              <span className="val">{item.maxPrice ? `¥${item.maxPrice}` : '无封顶'}</span>
            </div>
          </div>

          <div className="status-row">
            <span className="duration-info">竞拍时长: {item.duration}秒</span>
            <span className="status-tag">{item.status}</span>
          </div>

          <h2 className="item-title">{item.title}</h2>
          <p className="item-desc">{item.description}</p>

          <div className="rule-summary-card">
            <h4>竞拍规则摘要</h4>
            <ul>
              <li>0元起拍，每次至少加价{item.minIncrement}元</li>
              <li>最后30秒内出价自动延长30秒</li>
              <li>达到封顶价直接成交</li>
            </ul>
          </div>
        </div>
        <div className="bottom-spacer" />
      </div>

      <div className="bottom-op-bar">
        <div className="op-buttons-left">
          <button className="small-op-btn">直播</button>
          <button className="small-op-btn">客服</button>
        </div>
        <Button className="bid-main-btn" color="primary" block onClick={toggleBidPanel}>
          立即出价
        </Button>
      </div>
    </div>
  )
}

export default AuctionDetailPage
