import React from 'react'
import { Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './AuctionEndPanel.scss'

const AuctionEndPanel: React.FC = () => {
  const { closeAllModals, items, currentItemId } = useLiveRoomStore()
  const item = items.find(i => i.id === currentItemId)
  const statusText = item?.status === '已成交' ? '恭喜成交!' : item?.status === '已流拍' ? '本场流拍' : '竞拍已结束'

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="panel-inner" onClick={(e) => e.stopPropagation()}>
        <div className="end-icon">🏁</div>
        <h2 className="end-title">{statusText}</h2>
        <p className="end-price">最终成交价: ¥{item?.currentPrice}</p>
        <Button className="back-btn" color="primary" block onClick={closeAllModals}>
          返回直播间
        </Button>
      </div>
    </div>
  )
}

export default AuctionEndPanel
