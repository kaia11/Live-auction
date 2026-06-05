import { Button, Modal } from 'antd'

interface RuleModalProps {
  open: boolean
  onClose: () => void
}

function RuleModal({ open, onClose }: RuleModalProps) {
  return (
    <Modal
      title="竞拍规则说明"
      open={open}
      onCancel={onClose}
      footer={null}
      width={560}
      centered
      className="jewel-modal"
    >
      <ul className="rule-list">
        <li>起拍价可为 0 元</li>
        <li>每次加价需为加价幅度的整数倍</li>
        <li>封顶价可为空（无封顶）</li>
        <li>最后 30 秒出价，自动延时 N 秒（N 由商家填写）</li>
        <li>达到封顶价后立即成交并生成订单</li>
        <li>异常场次支持主播人工取消</li>
      </ul>
      <div className="modal-footer-center">
        <Button type="primary" onClick={onClose}>
          我知道了
        </Button>
      </div>
    </Modal>
  )
}

export default RuleModal
