import React, { useEffect } from 'react'
import { Routes, Route, useNavigate, useLocation, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd-mobile'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import LiveRoomsPage from './pages/LiveRoomsPage'
import LiveRoomPage from './pages/LiveRoomPage'
import AuctionDetailPage from './pages/AuctionDetailPage'
import ProfilePage from './pages/ProfilePage'
import OrderPaymentPage from './pages/OrderPaymentPage'
import { useUserStore } from './stores/useUserStore'

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, isHydrated } = useUserStore()
  const navigate = useNavigate()
  const location = useLocation()
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register'

  useEffect(() => {
    if (isHydrated && !user?.isLoggedIn && !isAuthPage) {
      navigate('/login')
    }
  }, [isAuthPage, isHydrated, user, navigate])

  if (!isHydrated && !isAuthPage) {
    return null
  }

  return <>{children}</>
}

function App() {
  const hydrateUser = useUserStore((state) => state.hydrateUser)

  useEffect(() => {
    void hydrateUser()
  }, [hydrateUser])

  return (
    <ConfigProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/rooms"
          element={
            <ProtectedRoute>
              <LiveRoomsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/live/:roomId"
          element={
            <ProtectedRoute>
              <LiveRoomPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/auction/:itemId"
          element={
            <ProtectedRoute>
              <AuctionDetailPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <ProfilePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/orders/:orderId/pay"
          element={
            <ProtectedRoute>
              <OrderPaymentPage />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </ConfigProvider>
  )
}

export default App
