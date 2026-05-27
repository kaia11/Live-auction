import { Navigate, Route, Routes } from 'react-router-dom'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'
import PublishPage from '@/pages/PublishPage'
import GoodsPage from '@/pages/GoodsPage'
import OrdersPage from '@/pages/OrdersPage'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/dashboard" element={<DashboardPage />} />
      <Route path="/publish" element={<PublishPage />} />
      <Route path="/goods" element={<GoodsPage />} />
      <Route path="/orders" element={<OrdersPage />} />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  )
}

export default App
