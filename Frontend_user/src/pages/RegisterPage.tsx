import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loading, Checkbox, Toast } from 'antd-mobile'
import { useUserStore } from '../stores/useUserStore'
import './LoginPage.scss'

const RegisterPage: React.FC = () => {
  const navigate = useNavigate()
  const { register } = useUserStore()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [agreed, setAgreed] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleRegister = async () => {
    if (!username || !password || !confirmPassword) return
    if (password !== confirmPassword) {
      Toast.show('两次输入的密码不一致')
      return
    }
    if (!agreed) return

    setLoading(true)
    const success = await register(username, password)
    setLoading(false)

    if (success) {
      navigate('/rooms')
      return
    }

    Toast.show('注册失败，用户名可能已存在')
  }

  return (
    <div className="login-page page-container">
      <div className="status-bar">
        <span className="time">15:25</span>
        <div className="status-icons"></div>
      </div>

      <div className="top-nav">
        <button className="nav-btn" onClick={() => navigate('/login')} type="button">‹</button>
        <span className="help-text">返回登录</span>
      </div>

      <div className="spacer-48"></div>

      <h1 className="login-title">创建用户端账号</h1>
      <p className="auth-subtitle">注册的新账号仅可登录用户端，默认直接拥有观众身份。</p>

      <div className="input-wrapper">
        <div className="auth-input">
          <span className="input-icon">@</span>
          <div className="divider"></div>
          <input
            className="input-field"
            placeholder="设置用户名"
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
            className="input-field"
            placeholder="设置密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            type="password"
            autoComplete="new-password"
          />
        </div>
      </div>

      <div className="input-wrapper">
        <div className="auth-input">
          <span className="eye-icon">✓</span>
          <div className="divider"></div>
          <input
            className="input-field"
            placeholder="确认密码"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            type="password"
            autoComplete="new-password"
          />
        </div>
      </div>

      <div className="links-row">
        <button className="link-btn" onClick={() => navigate('/login')} type="button">已有账号，去登录</button>
        <span className="link-text">用户名建议使用字母和数字组合</span>
      </div>

      <button className="login-btn" onClick={handleRegister} disabled={loading}>
        {loading ? <Loading color="white" /> : '注册并进入'}
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
          已阅读并同意 用户协议 和 隐私政策，并知晓本账号仅可在用户端使用
        </p>
      </div>
    </div>
  )
}

export default RegisterPage
