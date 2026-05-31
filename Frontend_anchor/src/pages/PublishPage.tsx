import { App, Button, Card, Col, Form, Input, InputNumber, Row } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import AdminLayout from '@/layouts/AdminLayout'
import RuleModal from '@/components/modals/RuleModal'
import { PublishSuccessModal } from '@/components/modals/StatusModals'
import { createItem } from '@/api/admin'
import { getMerchantItemImage } from '@/assets/localImages'
import { useAdminStore } from '@/stores/useAdminStore'

function PublishPage() {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [ruleOpen, setRuleOpen] = useState(false)
  const [successOpen, setSuccessOpen] = useState(false)
  const [publishedName, setPublishedName] = useState('珠宝拍品')
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm()
  const currentRoomId = useAdminStore((state) => state.currentRoomId)

  return (
    <AdminLayout activePath="/publish" title="竞拍发布">
      <Card className="jewel-card">
        <Form
          layout="vertical"
          form={form}
          initialValues={{
            name: '天然冰种翡翠圆牌吊坠',
            intro: '18K金镶嵌，证书齐全，珠宝直播专场拍品',
            startPrice: 0,
            increment: 20,
            durationSec: 900,
            delaySec: 15,
          }}
          onFinish={async (values) => {
            if (!currentRoomId) {
              message.warning('请先在右上角选择直播间')
              return
            }

            setSubmitting(true)
            try {
              await createItem(currentRoomId, {
                title: values.name,
                coverImage: getMerchantItemImage(values.name),
                description: values.intro,
                startPrice: values.startPrice,
                incrementStep: values.increment,
                ceilingPrice: typeof values.ceilingPrice === 'number' ? values.ceilingPrice : null,
                durationSeconds: values.durationSec,
                extensionSeconds: values.delaySec,
                extensionTriggerSeconds: 30,
              })
              setPublishedName(values.name)
              setSuccessOpen(true)
              message.success('拍品发布成功')
            } catch (error) {
              const nextMessage =
                error instanceof Error ? error.message : '拍品发布失败，请稍后重试'
              message.error(nextMessage)
            } finally {
              setSubmitting(false)
            }
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="商品名称" name="name" rules={[{ required: true }]}>
                <Input size="large" placeholder="请输入拍卖品名称" />
              </Form.Item>
              <Form.Item label="商品简介" name="intro" rules={[{ required: true }]}>
                <Input.TextArea rows={4} placeholder="请输入拍卖品简介" />
              </Form.Item>
              <Form.Item label="商品图片（Mock）">
                <div className="mock-upload-preview">
                  <img src={getMerchantItemImage(form.getFieldValue('name'))} alt="mock cover" />
                  <div className="mock-upload-copy">当前使用商家端本地 mock 图作为直播占位画面</div>
                </div>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="起拍价" name="startPrice" rules={[{ required: true }]}>
                <InputNumber size="large" min={0} style={{ width: '100%' }} addonAfter="元" />
              </Form.Item>
              <Form.Item label="固定加价幅度" name="increment" rules={[{ required: true }]}>
                <InputNumber size="large" min={1} style={{ width: '100%' }} addonAfter="元" />
              </Form.Item>
              <Form.Item label="封顶价（可空）" name="ceilingPrice">
                <InputNumber size="large" min={0} style={{ width: '100%' }} addonAfter="元" />
              </Form.Item>
              <Form.Item label="竞拍总时长（秒）" name="durationSec" rules={[{ required: true }]}>
                <InputNumber size="large" min={60} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item label="延时机制（最后30秒出价）" name="delaySec" rules={[{ required: true }]}>
                <InputNumber size="large" min={5} style={{ width: '100%' }} addonAfter="秒" />
              </Form.Item>
            </Col>
          </Row>
          <div className="publish-actions">
            <Button onClick={() => setRuleOpen(true)}>规则说明</Button>
            <Button onClick={() => form.resetFields()}>清空重填</Button>
            <Button type="default">保存草稿</Button>
            <Button type="primary" htmlType="submit" loading={submitting}>
              发布竞拍商品
            </Button>
          </div>
        </Form>
      </Card>

      <RuleModal open={ruleOpen} onClose={() => setRuleOpen(false)} />
      <PublishSuccessModal
        open={successOpen}
        goodsName={publishedName}
        onClose={() => {
          setSuccessOpen(false)
          navigate('/goods')
        }}
      />
    </AdminLayout>
  )
}

export default PublishPage
