import { Button, Card, Checkbox, Form, Input, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'

function LoginPage() {
  const navigate = useNavigate()

  return (
    <div className="login-page">
      <Card className="login-card">
        <Typography.Title level={2}>珠宝直播拍卖后台</Typography.Title>
        <Typography.Paragraph type="secondary">
          主播/商家端（PC 管理后台） - 本页为 Mock 登录
        </Typography.Paragraph>
        <Form
          layout="vertical"
          initialValues={{ account: 'anchor_admin', password: '123456', agree: true }}
          onFinish={() => navigate('/dashboard')}
        >
          <Form.Item label="账号 / 手机号" name="account" rules={[{ required: true }]}>
            <Input size="large" placeholder="请输入账号" />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true }]}>
            <Input.Password size="large" placeholder="请输入密码" />
          </Form.Item>
          <Form.Item name="agree" valuePropName="checked">
            <Checkbox>我已阅读并同意《直播竞拍服务协议》</Checkbox>
          </Form.Item>
          <Button size="large" type="primary" htmlType="submit" block>
            登录后台
          </Button>
        </Form>
      </Card>
    </div>
  )
}

export default LoginPage
