import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from 'antd-mobile'
import { AuctionItem } from '../types'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './CurrentAuctionCard.scss'

interface Props {
  item: AuctionItem
  onClose: () => void
}

const CurrentAuctionCard: React.FC<Props> = ({ item, onClose }) => {
  const navigate = useNavigate()
  const { toggleBidPanel, top3Ranking, myBidStatus } = useLiveRoomStore()

  const goToDetail = () => {
    navigate(`/auction/${item.id}`)
  }

  return (
    <div className="current-auction-card">
      <div className="card-top-row">
        <div className="item-thumb" />
        <div className="item-info-text">
          <h4 className="item-title">{item.title}</h4>
          <div className="price-countdown-row">
            <span className="current-price">¥{item.currentPrice}</span>
            <span className="countdown-text">⏱ {300}s</span>
          </div>
        </div>
        <span className="close-btn" onClick={onClose}>×</span>
      </div>

      <div className="ranking-area">
        <div className="top3-col">
          {top3Ranking.slice(0, 3).map((r, idx) => (
            <div key={idx} className="rank-item">
              <span className={`rank-tag rank-${idx + 1}`}>{idx + 1}</span>
              <span className="rank-name">{r.nickname}</span>
            </div>
          ))}
        </div>
        <div className="my-status-col">
          <div className="my-rank">我的名次: {myBidStatus.myRank || '未上榜'}</div>
          <div className="my-price">我的出价: ¥{myBidStatus.myHighestPrice}</div>
        </div>
      </div>

      <div className="card-bottom-btns">
        <Button size="small" className="detail-btn" onClick={goToDetail}>详情</Button>
        <Button className="bid-btn" color="primary" onClick={toggleBidPanel}>立即出价</Button>
      </div>
    </div>
  )
}

export default CurrentAuctionCard
