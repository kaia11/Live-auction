import React from 'react'
import { Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import { useLiveRoomUIStore } from '../stores/useLiveRoomUIStore'
import { useUserStore } from '../stores/useUserStore'
import './AuctionEndPanel.scss'

const AuctionEndPanel: React.FC = () => {
  const currentUserId = useUserStore((state) => state.user?.id)
  const { closeAllModals, myBidStatus, items, currentItemId } = useLiveRoomStore()
  const endedAuctionItem = useLiveRoomUIStore((state) => state.endedAuctionItem)
  const item = endedAuctionItem ?? items.find(i => i.id === currentItemId)
  const isSold = item?.status === '已成交'
  const isPassed = item?.status === '已流拍'
  const isWinner =
    isSold &&
    (!!currentUserId &&
      (item?.currentLeader === currentUserId || myBidStatus.isLeading))

  let statusText = '竞拍已结束'
  if (isPassed) {
    statusText = '本场流拍'
  } else if (isSold && isWinner) {
    statusText = '恭喜成交!'
  } else if (isSold) {
    statusText = '未竞拍成功，您的保证金已退回'
  }

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="panel-inner" onClick={(e) => e.stopPropagation()}>
        <div className="end-icon">🏁</div>
        <h2 className={`end-title ${isSold && !isWinner ? 'is-lost' : ''}`}>{statusText}</h2>
        {isSold ? (
          <p className="end-price">最终成交价: ¥{item?.currentPrice}</p>
        ) : null}
        <Button className="auction-end-back-btn" color="primary" block onClick={closeAllModals}>
          返回直播间
        </Button>
      </div>
    </div>
  )
}

export default AuctionEndPanel
