import React from 'react'
import { Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './BidSuccessModal.scss'

const BidSuccessModal: React.FC = () => {
  const { closeAllModals, myBidStatus, items, currentItemId } = useLiveRoomStore()
  const item = items.find(i => i.id === currentItemId)

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="modal-inner" onClick={(e) => e.stopPropagation()}>
        <div className="success-icon">🎉</div>
        <h2 className="success-title">出价成功!</h2>
        <p className="success-price">当前价格: ¥{item?.currentPrice}</p>
        <p className="is-leading-text">{myBidStatus.isLeading ? '恭喜，目前您领先!' : '正在竞争中...'}</p>
        <Button className="bid-success-back-btn" color="primary" block onClick={closeAllModals}>
          返回直播间
        </Button>
      </div>
    </div>
  )
}

export default BidSuccessModal
