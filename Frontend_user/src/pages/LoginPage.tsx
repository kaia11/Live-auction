import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loading, Checkbox, Toast } from 'antd-mobile'
import { useUserStore } from '../stores/useUserStore'
import './LoginPage.scss'

const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login } = useUserStore()
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [agreed, setAgreed] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleLogin = async () => {
    if (!phone || !password) return
    if (!agreed) return
    setLoading(true)
    const success = await login(phone, password)
    setLoading(false)
    if (success) {
      navigate('/rooms')
    } else {
      Toast.show('登录失败，请稍后重试')
    }
  }

  return (
    <div className="login-page page-container">
      <div className="status-bar">
        <span className="time">15:25</span>
        <div className="status-icons"></div>
      </div>

      <div className="top-nav">
        <span className="back-btn">‹</span>
        <span className="help-text">帮助</span>
      </div>

      <div className="spacer-48"></div>

      <h1 className="login-title">手机号密码登录</h1>

      <div className="input-wrapper">
        <div className="phone-input">
          <span className="phone-prefix">+86 ▾</span>
          <div className="divider"></div>
          <input
            className="input-field"
            placeholder="请输入手机号"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            type="tel"
          />
        </div>
      </div>

      <div className="input-wrapper">
        <div className="password-input">
          <span className="eye-icon">◠</span>
          <div className="divider"></div>
          <input
            className="input-field"
            placeholder="请输入密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            type="password"
          />
        </div>
      </div>

      <div className="links-row">
        <span className="link-text">⇄ 验证码登录</span>
        <span className="link-text">忘记密码</span>
      </div>

      <button className="login-btn" onClick={handleLogin} disabled={loading}>
        {loading ? <Loading color="white" /> : '登录'}
      </button>

      <div className="agreement-row">
        <Checkbox
          checked={agreed}
          onChange={(val) => setAgreed(val)}
          icon={(checked) => (
            <div className={`custom-checkbox ${checked ? 'checked' : ''}`}>
            </div>
          )}
        />
        <p className="agreement-text">
          已阅读并同意 用户协议 和 隐私政策，同时登录并使用抖音及其关联产品相关服务
        </p>
      </div>
    </div>
  )
}

export default LoginPage
