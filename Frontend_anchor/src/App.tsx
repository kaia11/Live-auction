import { useEffect, type ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'
import PublishPage from '@/pages/PublishPage'
import GoodsPage from '@/pages/GoodsPage'
import OrdersPage from '@/pages/OrdersPage'
import { getRooms } from '@/api/rooms'
import { useAdminStore } from '@/stores/useAdminStore'
import { useAuthStore } from '@/stores/useAuthStore'

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { token, hydrated } = useAuthStore()

  if (!hydrated) {
    return null
  }

  if (!token) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  const restoreSession = useAuthStore((state) => state.restoreSession)
  const token = useAuthStore((state) => state.token)
  const setRooms = useAdminStore((state) => state.setRooms)

  useEffect(() => {
    void restoreSession()
  }, [restoreSession])

  useEffect(() => {
    if (!token) {
      setRooms([])
      return
    }

    const loadRooms = async () => {
      try {
        const rooms = await getRooms()
        setRooms(rooms)
      } catch {
        setRooms([])
      }
    }

    void loadRooms()
  }, [setRooms, token])

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
      <Route path="/publish" element={<ProtectedRoute><PublishPage /></ProtectedRoute>} />
      <Route path="/goods" element={<ProtectedRoute><GoodsPage /></ProtectedRoute>} />
      <Route path="/orders" element={<ProtectedRoute><OrdersPage /></ProtectedRoute>} />
      <Route path="*" element={<Navigate to={token ? '/dashboard' : '/login'} replace />} />
    </Routes>
  )
}

export default App
