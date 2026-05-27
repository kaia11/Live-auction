import { Button, Card, Image, Input, Select, Space, Table, Tag } from 'antd'
import { useMemo, useState } from 'react'
import AdminLayout from '@/layouts/AdminLayout'
import { mockGoods } from '@/mock/data'
import type { AuctionGoods } from '@/types'
import { GoodsEditDrawer, RuleConfigDrawer } from '@/components/goods/GoodsDrawers'
import RuleModal from '@/components/modals/RuleModal'
import { AuctionResultModal, ExceptionCancelModal } from '@/components/modals/StatusModals'

function GoodsPage() {
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<string>('全部')
  const [selectedGoods, setSelectedGoods] = useState<AuctionGoods | undefined>()

  const [editOpen, setEditOpen] = useState(false)
  const [rulePanelOpen, setRulePanelOpen] = useState(false)
  const [ruleModalOpen, setRuleModalOpen] = useState(false)
  const [cancelOpen, setCancelOpen] = useState(false)
  const [resultOpen, setResultOpen] = useState(false)
  const [resultStatus, setResultStatus] = useState<'成交' | '流拍' | '异常取消'>('成交')

  const goodsList = useMemo(() => {
    return mockGoods.filter((item) => {
      const hitKeyword = !keyword || item.name.includes(keyword) || item.id.includes(keyword)
      const hitStatus = status === '全部' || item.status === status
      return hitKeyword && hitStatus
    })
  }, [keyword, status])

  return (
    <AdminLayout
      activePath="/goods"
      title="商品管理"
      actions={
        <Button type="primary" onClick={() => setRulePanelOpen(true)}>
          全局规则配置
        </Button>
      }
    >
      <Card className="jewel-card">
        <div className="toolbar">
          <Input.Search
            allowClear
            placeholder="请输入商品名称或ID"
            style={{ width: 320 }}
            onSearch={(v) => setKeyword(v)}
          />
          <Select
            style={{ width: 180 }}
            value={status}
            onChange={setStatus}
            options={['全部', '待上架', '即将开始', '竞拍中', '已成交', '已流拍', '已取消'].map(
              (item) => ({ label: item, value: item }),
            )}
          />
          <Button onClick={() => setRuleModalOpen(true)}>规则说明</Button>
        </div>

        <Table
          rowKey="id"
          dataSource={goodsList}
          pagination={false}
          columns={[
            {
              title: '商品',
              dataIndex: 'name',
              width: 340,
              render: (_, row: AuctionGoods) => (
                <Space>
                  <Image src={row.cover} width={56} height={56} style={{ borderRadius: 8 }} />
                  <div>
                    <div className="goods-name">{row.name}</div>
                    <div className="goods-sub">ID: {row.id}</div>
                  </div>
                </Space>
              ),
            },
            { title: '起拍价', dataIndex: 'startPrice', render: (v: number) => `¥${v}` },
            { title: '固定加价', dataIndex: 'increment', render: (v: number) => `¥${v}` },
            {
              title: '封顶价',
              dataIndex: 'ceilingPrice',
              render: (v?: number) => (typeof v === 'number' ? `¥${v}` : '无封顶'),
            },
            { title: '当前出价', dataIndex: 'currentPrice', render: (v: number) => `¥${v}` },
            { title: '出价次数', dataIndex: 'bidCount' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (v: AuctionGoods['status']) => {
                const colorMap: Record<AuctionGoods['status'], string> = {
                  待上架: 'default',
                  即将开始: 'gold',
                  竞拍中: 'processing',
                  已成交: 'success',
                  已流拍: 'warning',
                  已取消: 'error',
                }
                return <Tag color={colorMap[v]}>{v}</Tag>
              },
            },
            {
              title: '操作',
              width: 320,
              render: (_, row: AuctionGoods) => (
                <Space wrap>
                  <Button
                    size="small"
                    disabled={!['待上架', '即将开始'].includes(row.status)}
                    onClick={() => {
                      setSelectedGoods(row)
                      setEditOpen(true)
                    }}
                  >
                    编辑规则
                  </Button>
                  <Button
                    size="small"
                    danger
                    disabled={!['待上架', '即将开始'].includes(row.status)}
                    onClick={() => {
                      setSelectedGoods(row)
                      setCancelOpen(true)
                    }}
                  >
                    取消竞拍
                  </Button>
                  <Button
                    size="small"
                    type="primary"
                    ghost
                    onClick={() => {
                      setSelectedGoods(row)
                      setResultStatus(row.status === '已成交' ? '成交' : '流拍')
                      setResultOpen(true)
                    }}
                  >
                    查看结果
                  </Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <GoodsEditDrawer
        open={editOpen}
        goods={selectedGoods}
        onClose={() => setEditOpen(false)}
        onSave={() => setEditOpen(false)}
      />
      <RuleConfigDrawer open={rulePanelOpen} onClose={() => setRulePanelOpen(false)} />
      <RuleModal open={ruleModalOpen} onClose={() => setRuleModalOpen(false)} />
      <ExceptionCancelModal
        open={cancelOpen}
        goodsName={selectedGoods?.name ?? ''}
        onClose={() => setCancelOpen(false)}
        onConfirm={() => {
          setCancelOpen(false)
          setResultStatus('异常取消')
          setResultOpen(true)
        }}
      />
      <AuctionResultModal
        open={resultOpen}
        status={resultStatus}
        amount={selectedGoods?.currentPrice ?? 0}
        onClose={() => setResultOpen(false)}
      />
    </AdminLayout>
  )
}

export default GoodsPage
