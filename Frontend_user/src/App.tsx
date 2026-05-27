import React, { useEffect } from 'react'
import { Routes, Route, useNavigate, useLocation, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd-mobile'
import LoginPage from './pages/LoginPage'
import LiveRoomsPage from './pages/LiveRoomsPage'
import LiveRoomPage from './pages/LiveRoomPage'
import AuctionDetailPage from './pages/AuctionDetailPage'
import ProfilePage from './pages/ProfilePage'
import { useUserStore } from './stores/useUserStore'

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, isHydrated } = useUserStore()
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    if (isHydrated && !user?.isLoggedIn && location.pathname !== '/login') {
      navigate('/login')
    }
  }, [isHydrated, user, navigate, location.pathname])

  if (!isHydrated && location.pathname !== '/login') {
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
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </ConfigProvider>
  )
}

export default App
