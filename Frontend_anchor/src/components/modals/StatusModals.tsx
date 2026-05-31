import { Button, Descriptions, Input, Modal, Result, Tag } from 'antd'
import { useState } from 'react'

interface PublishSuccessModalProps {
  open: boolean
  goodsName: string
  onClose: () => void
}

export function PublishSuccessModal({ open, goodsName, onClose }: PublishSuccessModalProps) {
  return (
    <Modal open={open} footer={null} onCancel={onClose} centered className="jewel-modal">
      <Result
        status="success"
        title="发布成功"
        subTitle={`${goodsName} 已进入待上架队列`}
        extra={[
          <Button key="ok" type="primary" onClick={onClose}>
            返回商品管理
          </Button>,
        ]}
      />
    </Modal>
  )
}

interface ExceptionCancelModalProps {
  open: boolean
  goodsName: string
  onConfirm: (reason: string) => void
  onClose: () => void
}

export function ExceptionCancelModal({
  open,
  goodsName,
  onConfirm,
  onClose,
}: ExceptionCancelModalProps) {
  const [reason, setReason] = useState('')
  return (
    <Modal
      open={open}
      title="异常竞拍确认"
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          返回
        </Button>,
        <Button
          key="confirm"
          danger
          type="primary"
          onClick={() => {
            onConfirm(reason || '主播手动取消')
            setReason('')
          }}
        >
          确认取消
        </Button>,
      ]}
      centered
      className="jewel-modal"
    >
      <p>你正在取消拍品：{goodsName}</p>
      <Input.TextArea
        rows={3}
        value={reason}
        onChange={(e) => setReason(e.target.value)}
        placeholder="填写取消原因（可选）"
      />
    </Modal>
  )
}

interface AuctionResultModalProps {
  open: boolean
  status: '成交' | '流拍' | '异常取消'
  amount: number
  onClose: () => void
}

export function AuctionResultModal({ open, status, amount, onClose }: AuctionResultModalProps) {
  const color = status === '成交' ? 'success' : status === '流拍' ? 'warning' : 'error'
  return (
    <Modal open={open} footer={null} onCancel={onClose} centered className="jewel-modal">
      <Descriptions title="竞拍结束结果" bordered column={1}>
        <Descriptions.Item label="结果状态">
          <Tag color={color}>{status}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="成交金额">¥ {amount.toLocaleString()}</Descriptions.Item>
      </Descriptions>
      <div className="modal-footer-center">
        <Button type="primary" onClick={onClose}>
          关闭
        </Button>
      </div>
    </Modal>
  )
}

interface OrderProgressModalProps {
  open: boolean
  orderId: string
  status: string
  logisticsNo?: string
  actions?: React.ReactNode
  onClose: () => void
}

export function OrderProgressModal({
  open,
  orderId,
  status,
  logisticsNo,
  actions,
  onClose,
}: OrderProgressModalProps) {
  return (
    <Modal
      open={open}
      title="订单进度更新"
      onCancel={onClose}
      onOk={onClose}
      okText="我知道了"
      centered
      className="jewel-modal"
    >
      <Descriptions bordered column={1}>
        <Descriptions.Item label="订单号">{orderId}</Descriptions.Item>
        <Descriptions.Item label="当前状态">{status}</Descriptions.Item>
        <Descriptions.Item label="物流单号">{logisticsNo ?? '待生成'}</Descriptions.Item>
      </Descriptions>
      {actions ? <div className="modal-footer-center">{actions}</div> : null}
    </Modal>
  )
}
