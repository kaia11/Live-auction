import React from 'react'
import { createPortal } from 'react-dom'
import { Button } from 'antd-mobile'
import './DepositPaymentModal.scss'

interface Props {
  amount: number
  onCancel: () => void
  onConfirm: () => void
}

const DepositPaymentModal: React.FC<Props> = ({ amount, onCancel, onConfirm }) => {
  if (typeof document === 'undefined') {
    return null
  }

  return createPortal(
    <div className="deposit-mask" onClick={onCancel}>
      <div className="deposit-modal" onClick={(event) => event.stopPropagation()}>
        <h3 className="deposit-title">支付保证金后可继续出价</h3>
        <p className="deposit-amount">保证金金额：¥{amount}</p>
        <div className="deposit-desc">
          <p>1. 拍下并付款成功后，保证金将原路退回。</p>
          <p>2. 若拍下后超过 24 小时未付款，保证金归卖家所有。</p>
        </div>
        <div className="deposit-actions">
          <Button className="deposit-btn cancel" onClick={onCancel}>
            暂不支付
          </Button>
          <Button className="deposit-btn confirm" color="primary" onClick={onConfirm}>
            确认支付
          </Button>
        </div>
      </div>
    </div>,
    document.body,
  )
}

export default DepositPaymentModal
