import { Button, Card, Col, Form, Input, InputNumber, Row, Upload } from 'antd'
import { useState } from 'react'
import AdminLayout from '@/layouts/AdminLayout'
import RuleModal from '@/components/modals/RuleModal'
import { PublishSuccessModal } from '@/components/modals/StatusModals'

function PublishPage() {
  const [ruleOpen, setRuleOpen] = useState(false)
  const [successOpen, setSuccessOpen] = useState(false)
  const [publishedName, setPublishedName] = useState('珠宝拍品')
  const [form] = Form.useForm()

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
          onFinish={(values) => {
            setPublishedName(values.name)
            setSuccessOpen(true)
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
                <Upload listType="picture-card" maxCount={3}>
                  上传图片
                </Upload>
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
            <Button type="primary" htmlType="submit">
              发布竞拍商品
            </Button>
          </div>
        </Form>
      </Card>

      <RuleModal open={ruleOpen} onClose={() => setRuleOpen(false)} />
      <PublishSuccessModal
        open={successOpen}
        goodsName={publishedName}
        onClose={() => setSuccessOpen(false)}
      />
    </AdminLayout>
  )
}

export default PublishPage
