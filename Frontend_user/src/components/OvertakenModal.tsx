import React from 'react'
import { Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './OvertakenModal.scss'

const OvertakenModal: React.FC = () => {
  const { closeAllModals, toggleBidPanel, myBidStatus, items, currentItemId } = useLiveRoomStore()
  const item = items.find(i => i.id === currentItemId)

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="modal-inner" onClick={(e) => e.stopPropagation()}>
        <h2 className="overtaken-title">⚠️ 您已被超越!</h2>
        <p className="new-price-text">当前最新价: ¥{item?.currentPrice}</p>
        <p className="my-max-text">我的最高出价: ¥{myBidStatus.myHighestPrice}</p>
        <div className="btn-row">
          <Button className="close-btn" onClick={closeAllModals}>稍后再说</Button>
          <Button className="bid-again-btn" color="primary" onClick={() => { closeAllModals(); toggleBidPanel(); }}>
            再次出价
          </Button>
        </div>
      </div>
    </div>
  )
}

export default OvertakenModal
