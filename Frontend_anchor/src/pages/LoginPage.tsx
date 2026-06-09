import { Alert, Button, Card, Checkbox, Form, Input, Typography } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/useAuthStore'

function LoginPage() {
  const navigate = useNavigate()
  const [errorMessage, setErrorMessage] = useState('')
  const { loginWithPassword, loading } = useAuthStore()

  return (
    <div className="login-page">
      <Card className="login-card">
        <Typography.Title level={2}>珠宝直播拍卖后台</Typography.Title>
        <Typography.Paragraph type="secondary">
          主播/商家端（PC 管理后台） - 使用独立账号体系登录
        </Typography.Paragraph>
        {errorMessage ? <Alert type="error" showIcon message={errorMessage} style={{ marginBottom: 16 }} /> : null}
        <Form
          layout="vertical"
          initialValues={{ username: 'silence001', password: '111111', agree: true }}
          onFinish={async (values) => {
            try {
              setErrorMessage('')
              await loginWithPassword({
                username: values.username,
                password: values.password,
              })
              navigate('/dashboard')
            } catch (error) {
              const nextMessage =
                error instanceof Error ? error.message : '登录失败，请稍后重试'
              setErrorMessage(nextMessage)
            }
          }}
        >
          <Form.Item label="用户名" name="username" rules={[{ required: true }]}>
            <Input size="large" placeholder="请输入用户名" autoComplete="username" />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true }]}>
            <Input.Password size="large" placeholder="请输入密码" autoComplete="current-password" />
          </Form.Item>
          <Form.Item name="agree" valuePropName="checked">
            <Checkbox>我已阅读并同意《直播竞拍服务协议》</Checkbox>
          </Form.Item>
          <Button size="large" type="primary" htmlType="submit" block loading={loading}>
            登录后台
          </Button>
          <div className="auth-switch-row">
            <Button type="link" onClick={() => navigate('/register')} style={{ paddingInline: 0 }}>
              没有账号？去注册主播端账号
            </Button>
            <Typography.Text type="secondary">示例账号：silence001 / 111111</Typography.Text>
          </div>
        </Form>
      </Card>
    </div>
  )
}

export default LoginPage
