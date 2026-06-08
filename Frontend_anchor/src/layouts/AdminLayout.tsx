import { Button, Layout, Menu, Select, Space, Tag } from 'antd'
import type { MenuProps } from 'antd'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAdminStore } from '@/stores/useAdminStore'
import { useAuthStore } from '@/stores/useAuthStore'

const menuItems: MenuProps['items'] = [
  { key: '/dashboard', label: '总览' },
  { key: '/publish', label: '竞拍发布' },
  { key: '/goods', label: '商品管理' },
  { key: '/orders', label: '订单管理' },
]

interface AdminLayoutProps {
  activePath: string
  title: string
  actions?: React.ReactNode
  children: React.ReactNode
}

function AdminLayout({ activePath, title, actions, children }: AdminLayoutProps) {
  const navigate = useNavigate()
  const { rooms, currentRoomId, setCurrentRoomId } = useAdminStore()
  const { user, logout } = useAuthStore()
  const currentRoom = useMemo(() => rooms.find((room) => room.id === currentRoomId), [rooms, currentRoomId])

  const roomOptions = useMemo(
    () =>
      rooms.map((room) => ({
        label: `${room.title} (${room.status === 'live' ? '直播中' : '未开播'})`,
        value: room.id,
      })),
    [rooms],
  )

  return (
    <Layout className="pc-admin-shell">
      <Layout.Sider width={220} className="pc-admin-sider">
        <div className="brand">珠宝竞拍后台</div>
        <Menu
          className="pc-admin-menu"
          mode="inline"
          selectedKeys={[activePath]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Layout.Sider>
      <Layout className="pc-admin-main">
        <Layout.Header className="pc-admin-header">
          <h1>{title}</h1>
          <Space size={12}>
            {roomOptions.length > 0 ? (
              <Select
                value={currentRoomId || undefined}
                style={{ width: 260 }}
                placeholder="请选择直播间"
                options={roomOptions}
                onChange={setCurrentRoomId}
              />
            ) : null}
            {currentRoom ? (
              <Tag color={currentRoom.status === 'live' ? 'green' : 'default'}>
                {currentRoom.status === 'live' ? '直播中' : '未开播'}
              </Tag>
            ) : null}
            {user ? <Tag color="blue">{user.nickname}</Tag> : null}
            {actions}
            <Button
              onClick={() => {
                logout()
                navigate('/login')
              }}
            >
              退出登录
            </Button>
          </Space>
        </Layout.Header>
        <Layout.Content className="pc-admin-content">{children}</Layout.Content>
      </Layout>
    </Layout>
  )
}

export default AdminLayout
