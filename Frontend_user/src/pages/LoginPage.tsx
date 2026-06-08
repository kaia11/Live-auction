import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loading, Checkbox, Toast } from 'antd-mobile'
import { useUserStore } from '../stores/useUserStore'
import './LoginPage.scss'

const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login } = useUserStore()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [agreed, setAgreed] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleLogin = async () => {
    if (!username || !password) return
    if (!agreed) return
    setLoading(true)
    const success = await login(username, password)
    setLoading(false)
    if (success) {
      navigate('/rooms')
    } else {
      Toast.show('登录失败，请检查用户名或密码')
    }
  }

  return (
    <div className="login-page page-container">
      <div className="status-bar">
        <span className="time">15:25</span>
        <div className="status-icons"></div>
      </div>

      <div className="top-nav">
        <button className="back-btn" type="button" onClick={() => navigate(-1)}>‹</button>
        <span className="help-text">帮助</span>
      </div>

      <div className="spacer-48"></div>

      <h1 className="login-title">用户名密码登录</h1>
      <p className="auth-subtitle">用户端账号与主播端隔离，注册后将直接进入直播列表。</p>

      <div className="input-wrapper">
        <div className="auth-input">
          <span className="input-icon">@</span>
          <div className="divider"></div>
          <input
            className="input-field"
            placeholder="请输入用户名"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
          />
        </div>
      </div>

      <div className="input-wrapper">
        <div className="auth-input">
          <span className="eye-icon">◠</span>
          <div className="divider"></div>
          <input
            className="input-field password-field"
            placeholder="请输入密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            type={showPassword ? 'text' : 'password'}
            autoComplete="current-password"
          />
          <button
            className="password-toggle"
            type="button"
            onClick={() => setShowPassword((value) => !value)}
            aria-label={showPassword ? '隐藏密码' : '显示密码'}
          >
            {showPassword ? '◉' : '◎'}
          </button>
        </div>
      </div>

      <div className="links-block">
        <button className="link-btn" onClick={() => navigate('/register')} type="button">去注册</button>
        <span className="link-text">示例账号：viewer_demo / 123456</span>
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
