import { Alert, Button, Card, Checkbox, Form, Input, Typography } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/useAuthStore'

function RegisterPage() {
  const navigate = useNavigate()
  const [errorMessage, setErrorMessage] = useState('')
  const { registerWithPassword, loading } = useAuthStore()

  return (
    <div className="login-page">
      <Card className="login-card">
        <Typography.Title level={2}>创建主播端账号</Typography.Title>
        <Typography.Paragraph type="secondary">
          注册的新账号仅用于主播/商家后台，成功后会直接进入总览页。
        </Typography.Paragraph>
        {errorMessage ? <Alert type="error" showIcon message={errorMessage} style={{ marginBottom: 16 }} /> : null}
        <Form
          layout="vertical"
          initialValues={{ agree: true }}
          onFinish={async (values) => {
            try {
              setErrorMessage('')
              if (values.password !== values.confirmPassword) {
                setErrorMessage('两次输入的密码不一致')
                return
              }
              await registerWithPassword({
                username: values.username,
                password: values.password,
              })
              navigate('/dashboard')
            } catch (error) {
              const nextMessage =
                error instanceof Error ? error.message : '注册失败，请稍后重试'
              setErrorMessage(nextMessage)
            }
          }}
        >
          <Form.Item label="用户名" name="username" rules={[{ required: true }]}>
            <Input size="large" placeholder="设置主播端用户名" autoComplete="username" />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true }]}>
            <Input.Password size="large" placeholder="设置密码" autoComplete="new-password" />
          </Form.Item>
          <Form.Item label="确认密码" name="confirmPassword" rules={[{ required: true }]}>
            <Input.Password size="large" placeholder="再次输入密码" autoComplete="new-password" />
          </Form.Item>
          <Form.Item name="agree" valuePropName="checked">
            <Checkbox>我已阅读并同意《直播竞拍服务协议》</Checkbox>
          </Form.Item>
          <Button size="large" type="primary" htmlType="submit" block loading={loading}>
            注册并进入后台
          </Button>
          <div className="auth-switch-row">
            <Button type="link" onClick={() => navigate('/login')} style={{ paddingInline: 0 }}>
              已有账号？去登录
            </Button>
            <Typography.Text type="secondary">注册账号与用户端不互通</Typography.Text>
          </div>
        </Form>
      </Card>
    </div>
  )
}

export default RegisterPage
