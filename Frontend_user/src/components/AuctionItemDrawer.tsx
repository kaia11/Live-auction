import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './AuctionItemDrawer.scss'

const AuctionItemDrawer: React.FC = () => {
  const navigate = useNavigate()
  const { items, closeAllModals, toggleBidPanel } = useLiveRoomStore()

  const goToDetail = (itemId: string) => {
    closeAllModals()
    navigate(`/auction/${itemId}`)
  }

  const getActionButtonText = (status: string) => {
    if (status === '竞拍中') return '立即出价'
    if (['已成交', '已流拍', '已取消', '已结束'].includes(status)) return '查看拍卖结果'
    return '查看倒计时'
  }

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="drawer-container" onClick={(e) => e.stopPropagation()}>
        <div className="drawer-handle-bar" />
        <h3 className="drawer-title">拍品列表</h3>
        <div className="drawer-divider" />
        <div className="items-scroll-area">
          {items.map((item) => (
            <div key={item.id} className="item-row-card" onClick={() => goToDetail(item.id)}>
              <div className="item-row-thumb">
                <img src={item.images[0]} alt={item.title} />
              </div>
              <div className="item-row-info">
                <div className="item-row-title">{item.title}</div>
                <div className="item-row-price">¥{item.currentPrice}</div>
              </div>
              <div className="item-row-action">
                <span className="status-label">{item.status}</span>
                <Button
                  size="small"
                  color="primary"
                  onClick={(e) => {
                    e.stopPropagation()
                    if (item.status === '竞拍中') {
                      toggleBidPanel()
                    } else {
                      goToDetail(item.id)
                    }
                  }}
                >
                  {getActionButtonText(item.status)}
                </Button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default AuctionItemDrawer
