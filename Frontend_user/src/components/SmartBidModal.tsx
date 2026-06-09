import React from 'react'
import { Button } from 'antd-mobile'
import './SmartBidModal.scss'

interface Props {
  maxPrice: number
  increment: number
  minAllowedPrice: number
  enabled: boolean
  onMinus: () => void
  onPlus: () => void
  onEnable: () => void
  onDisable: () => void
  onClose: () => void
}

const SmartBidModal: React.FC<Props> = ({
  maxPrice,
  increment,
  minAllowedPrice,
  enabled,
  onMinus,
  onPlus,
  onEnable,
  onDisable,
  onClose,
}) => {
  return (
    <div className="smart-bid-mask" onClick={onClose}>
      <div className="smart-bid-modal" onClick={(event) => event.stopPropagation()}>
        <div className="smart-bid-header">
          <h3>智能出价</h3>
          <button className="smart-bid-close" onClick={onClose}>×</button>
        </div>
        <p className="smart-bid-status">
          当前状态：{enabled ? `已开启（最高心理价 ¥${maxPrice}）` : '未开启'}
        </p>
        <div className="smart-bid-price-row">
          <button className="smart-adjust-btn" onClick={onMinus}>-</button>
          <span className="smart-price">¥{maxPrice}</span>
          <button className="smart-adjust-btn" onClick={onPlus}>+</button>
        </div>
        <p className="smart-tip">
          最低可设置 ¥{minAllowedPrice}，每次调节步长 ¥{increment}
        </p>

        <div className="smart-rules">
          <h4>规则说明</h4>
          <p>1. 你设置最高心理价后，系统会在不超过该价格时自动按最小加价幅度补价。</p>
          <p>2. 若多人同时开启智能出价，系统比较最高心理价，按更高心理价优先。</p>
          <p>3. 若最高心理价相同且都到达该价格，先出价/先开启的人保持领先。</p>
        </div>

        <div className="smart-actions">
          {enabled ? (
            <Button className="smart-disable-btn" onClick={onDisable}>
              关闭智能出价
            </Button>
          ) : null}
          <Button color="primary" className="smart-enable-btn" onClick={onEnable}>
            {enabled ? '更新最高心理价' : '开启智能出价'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default SmartBidModal
