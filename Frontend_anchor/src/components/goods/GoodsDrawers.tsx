import { Button, Descriptions, Drawer, Form, InputNumber, Space, Tag } from 'antd'
import type { AuctionGoods } from '@/types'

interface GoodsEditDrawerProps {
  open: boolean
  goods?: AuctionGoods
  onClose: () => void
  onSave: (values: {
    startPrice: number
    increment: number
    ceilingPrice?: number
    durationSec: number
  }) => void
}

export function GoodsEditDrawer({ open, goods, onClose, onSave }: GoodsEditDrawerProps) {
  return (
    <Drawer
      title="商品编辑面板"
      placement="right"
      width={460}
      onClose={onClose}
      open={open}
      className="jewel-drawer"
    >
      {goods ? (
        <>
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="拍品名">{goods.name}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag>{goods.status}</Tag>
            </Descriptions.Item>
          </Descriptions>
          <Form
            layout="vertical"
            className="drawer-form"
            initialValues={{
              startPrice: goods.startPrice,
              increment: goods.increment,
              ceilingPrice: goods.ceilingPrice,
              durationSec: goods.durationSec,
            }}
            onFinish={(values) => onSave(values)}
          >
            <Form.Item label="起拍价" name="startPrice" rules={[{ required: true }]}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="加价幅度" name="increment" rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="封顶价（可空）" name="ceilingPrice">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="竞拍总时长（秒）" name="durationSec" rules={[{ required: true }]}>
              <InputNumber min={60} style={{ width: '100%' }} />
            </Form.Item>
            <Space>
              <Button onClick={onClose}>取消</Button>
              <Button type="primary" htmlType="submit">
                保存规则
              </Button>
            </Space>
          </Form>
        </>
      ) : null}
    </Drawer>
  )
}

interface RuleConfigDrawerProps {
  open: boolean
  onClose: () => void
}

export function RuleConfigDrawer({ open, onClose }: RuleConfigDrawerProps) {
  return (
    <Drawer
      title="规则配置面板"
      placement="right"
      width={460}
      onClose={onClose}
      open={open}
      className="jewel-drawer"
    >
      <Form layout="vertical" className="drawer-form" initialValues={{ delaySec: 15 }}>
        <Form.Item label="默认延时规则（秒）" name="delaySec">
          <InputNumber min={5} style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item label="规则提示">
          <div className="rule-tip-box">
            最后 30 秒有人出价，系统在原时长基础上自动延时 15 秒。封顶价为空时即为无封顶。
          </div>
        </Form.Item>
        <Button type="primary" block onClick={onClose}>
          保存全局规则
        </Button>
      </Form>
    </Drawer>
  )
}
