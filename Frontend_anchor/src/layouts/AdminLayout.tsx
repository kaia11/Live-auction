import { Layout, Menu } from 'antd'
import type { MenuProps } from 'antd'
import { useNavigate } from 'react-router-dom'

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
          <div>{actions}</div>
        </Layout.Header>
        <Layout.Content className="pc-admin-content">{children}</Layout.Content>
      </Layout>
    </Layout>
  )
}

export default AdminLayout
