import React from 'react'
import { Button } from 'antd-mobile'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './RuleModal.scss'

const RuleModal: React.FC = () => {
  const { closeAllModals } = useLiveRoomStore()

  return (
    <div className="custom-mask" onClick={closeAllModals}>
      <div className="rule-modal-inner" onClick={(e) => e.stopPropagation()}>
        <h2 className="rule-title">竞拍规则</h2>
        <div className="rule-item-row">
          <span className="rule-label">0元起拍</span>
        </div>
        <div className="rule-item-row">
          <span className="rule-label">每次至少加价X元</span>
        </div>
        <div className="rule-item-row">
          <span className="rule-label">封顶价Y或无封顶</span>
        </div>
        <div className="rule-item-row">
          <span className="rule-label">最后30s内出价自动延长30s</span>
        </div>
        <div className="rule-item-row">
          <span className="rule-label">达到封顶价直接成交</span>
        </div>
        <Button className="rule-close-btn" color="primary" block onClick={closeAllModals}>
          我知道了
        </Button>
      </div>
    </div>
  )
}

export default RuleModal
