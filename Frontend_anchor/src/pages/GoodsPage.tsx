import { App, Button, Card, Input, Select, Space, Table, Tag } from 'antd'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { AxiosError } from 'axios'
import AdminLayout from '@/layouts/AdminLayout'
import {
  activateNextItem,
  cancelSession,
  getRoomItems,
  getRoomSessions,
  settleSession,
  startRoomLive,
  startSession,
  stopRoomLive,
  updateItem,
} from '@/api/admin'
import { GoodsCoverImage } from '@/components/GoodsCoverImage'
import type { GoodsRow } from '@/types'
import { GoodsEditDrawer, RuleConfigDrawer } from '@/components/goods/GoodsDrawers'
import RuleModal from '@/components/modals/RuleModal'
import { AuctionResultModal, ExceptionCancelModal } from '@/components/modals/StatusModals'
import { useAdminStore } from '@/stores/useAdminStore'

const queueStatusLabelMap: Record<string, string> = {
  queued: '待上架',
  upcoming: '即将开始',
  active: '竞拍中',
  finished: '已结束',
  cancelled: '已取消',
}

const sessionStatusLabelMap: Record<string, string> = {
  pending: '待上架',
  bidding: '竞拍中',
  ended_sold: '已成交',
  ended_passed: '已流拍',
  cancelled: '已取消',
}

const buildDisplayStatus = (row: GoodsRow) => {
  if (row.sessionStatus === 'bidding' && row.endTime) {
    const endTimeMs = new Date(row.endTime).getTime()
    if (Number.isFinite(endTimeMs) && endTimeMs <= Date.now()) {
      return '已结束'
    }
  }
  if (row.sessionStatus === 'ended_sold') {
    return '已成交'
  }
  if (row.sessionStatus === 'ended_passed') {
    return '已流拍'
  }
  if (row.sessionStatus === 'cancelled' || row.queueStatus === 'cancelled') {
    return '已取消'
  }
  if (row.sessionStatus === 'bidding' || row.queueStatus === 'active') {
    return '竞拍中'
  }
  if (row.queueStatus === 'upcoming') {
    return '即将开始'
  }
  return queueStatusLabelMap[row.queueStatus] ?? sessionStatusLabelMap[row.sessionStatus] ?? '待上架'
}

const getApiErrorMessage = (error: unknown, fallback: string) => {
  const err = error as AxiosError<{ message?: string; code?: number }>
  const apiMessage = err.response?.data?.message
  if (apiMessage) {
    return apiMessage
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  return fallback
}

function GoodsPage() {
  const { message } = App.useApp()
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<string>('全部')
  const [goodsRows, setGoodsRows] = useState<GoodsRow[]>([])
  const [selectedGoods, setSelectedGoods] = useState<GoodsRow | undefined>()
  const [loading, setLoading] = useState(false)

  const [editOpen, setEditOpen] = useState(false)
  const [rulePanelOpen, setRulePanelOpen] = useState(false)
  const [ruleModalOpen, setRuleModalOpen] = useState(false)
  const [cancelOpen, setCancelOpen] = useState(false)
  const [resultOpen, setResultOpen] = useState(false)
  const [resultStatus, setResultStatus] = useState<'成交' | '流拍' | '异常取消'>('成交')
  const currentRoomId = useAdminStore((state) => state.currentRoomId)
  const rooms = useAdminStore((state) => state.rooms)
  const updateRoomStatus = useAdminStore((state) => state.updateRoomStatus)
  const currentRoom = useMemo(
    () => rooms.find((room) => room.id === currentRoomId),
    [rooms, currentRoomId],
  )

  const refreshGoods = useCallback(async () => {
    if (!currentRoomId) {
      return
    }

    const [items, sessions] = await Promise.all([getRoomItems(currentRoomId), getRoomSessions(currentRoomId)])
    const rows = items.map((item) => {
      const session = sessions.find((entry) => entry.itemId === item.id)
      const row: GoodsRow = {
        itemId: item.id,
        sessionId: session?.sessionId ?? '',
        roomId: item.roomId,
        title: item.title,
        coverImage: item.coverImage,
        description: item.description,
        startPrice: item.startPrice,
        incrementStep: item.incrementStep,
        ceilingPrice: item.ceilingPrice,
        durationSeconds: item.durationSeconds,
        extensionSeconds: item.extensionSeconds,
        extensionTriggerSeconds: item.extensionTriggerSeconds,
        currentPrice: session?.currentPrice ?? item.startPrice,
        queueStatus: session?.queueStatus ?? item.queueStatus,
        sessionStatus: session?.status ?? 'pending',
        endTime: session?.endTime ?? '',
        displayStatus: '',
      }

      row.displayStatus = buildDisplayStatus(row)
      return row
    })
    setGoodsRows(rows)
  }, [currentRoomId])

  useEffect(() => {
    if (!currentRoomId) {
      setGoodsRows([])
      return
    }

    const loadGoods = async () => {
      setLoading(true)
      try {
        await refreshGoods()
      } catch (error) {
        const nextMessage =
          error instanceof Error ? error.message : '商品数据加载失败，请稍后重试'
        message.error(nextMessage)
      } finally {
        setLoading(false)
      }
    }

    void loadGoods()
  }, [currentRoomId, message, refreshGoods])

  const goodsList = useMemo(() => {
    return goodsRows.filter((item) => {
      const hitKeyword = !keyword || item.title.includes(keyword) || item.itemId.includes(keyword)
      const hitStatus = status === '全部' || item.displayStatus === status
      return hitKeyword && hitStatus
    })
  }, [goodsRows, keyword, status])

  useEffect(() => {
    if (!currentRoomId) {
      return
    }

    const timer = window.setInterval(() => {
      void refreshGoods()
    }, 5000)

    return () => window.clearInterval(timer)
  }, [currentRoomId, refreshGoods])

  return (
    <AdminLayout
      activePath="/goods"
      title="商品管理"
      actions={
        <Space>
          <Button
            type="primary"
            disabled={currentRoom?.status === 'live'}
            onClick={async () => {
              if (!currentRoomId) {
                message.warning('请先选择直播间')
                return
              }

              try {
                await startRoomLive(currentRoomId)
                updateRoomStatus(currentRoomId, 'live')
                message.success('直播已开始')
                await refreshGoods()
              } catch (error) {
                const nextMessage = getApiErrorMessage(error, '开始直播失败，请稍后重试')
                message.error(nextMessage)
              }
            }}
          >
            {currentRoom?.status === 'live' ? '直播中' : '开始直播'}
          </Button>
          <Button
            danger
            disabled={currentRoom?.status !== 'live'}
            onClick={async () => {
              if (!currentRoomId) {
                message.warning('请先选择直播间')
                return
              }

              try {
                await stopRoomLive(currentRoomId)
                updateRoomStatus(currentRoomId, 'offline')
                message.success('已下播')
                await refreshGoods()
              } catch (error) {
                const nextMessage = getApiErrorMessage(error, '下播失败，请稍后重试')
                message.error(nextMessage)
              }
            }}
          >
            下播
          </Button>
          <Button onClick={() => setRulePanelOpen(true)}>全局规则配置</Button>
          <Button
            type="primary"
            onClick={async () => {
              if (!currentRoomId) {
                message.warning('请先选择直播间')
                return
              }

              try {
                await activateNextItem(currentRoomId)
                message.success('已切换到下一件待开始拍品')
                await refreshGoods()
              } catch (error) {
                const nextMessage =
                  error instanceof Error ? error.message : '切换下一件失败，请稍后重试'
                message.error(nextMessage)
              }
            }}
          >
            切换下一件
          </Button>
        </Space>
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
          rowKey="itemId"
          dataSource={goodsList}
          loading={loading}
          pagination={false}
          columns={[
            {
              title: '商品',
              dataIndex: 'title',
              width: 340,
              render: (_, row: GoodsRow) => (
                <Space>
                  <GoodsCoverImage coverImage={row.coverImage} itemId={row.itemId} width={56} height={56} style={{ borderRadius: 8 }} />
                  <div>
                    <div className="goods-name">{row.title}</div>
                    <div className="goods-sub">ID: {row.itemId}</div>
                  </div>
                </Space>
              ),
            },
            { title: '起拍价', dataIndex: 'startPrice', render: (v: number) => `¥${v}` },
            { title: '固定加价', dataIndex: 'incrementStep', render: (v: number) => `¥${v}` },
            {
              title: '封顶价',
              dataIndex: 'ceilingPrice',
              render: (v?: number) => (typeof v === 'number' ? `¥${v}` : '无封顶'),
            },
            { title: '当前出价', dataIndex: 'currentPrice', render: (v: number) => `¥${v}` },
            { title: '场次', dataIndex: 'sessionId', render: (v: string) => v || '--' },
            {
              title: '状态',
              dataIndex: 'displayStatus',
              render: (v: GoodsRow['displayStatus']) => {
                const colorMap: Record<string, string> = {
                  待上架: 'default',
                  即将开始: 'gold',
                  竞拍中: 'processing',
                  已结束: 'default',
                  已成交: 'success',
                  已流拍: 'warning',
                  已取消: 'error',
                }
                return <Tag color={colorMap[v]}>{v}</Tag>
              },
            },
            {
              title: '操作',
              width: 420,
              render: (_, row: GoodsRow) => (
                <Space wrap>
                  <Button
                    size="small"
                    disabled={!['待上架', '即将开始'].includes(row.displayStatus)}
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
                    disabled={!['待上架', '即将开始', '竞拍中'].includes(row.displayStatus)}
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
                  disabled={row.sessionStatus !== 'pending' || currentRoom?.status !== 'live'}
                  onClick={async () => {
                    if (currentRoom?.status !== 'live') {
                      message.warning('请先开始直播')
                      return
                    }
                    try {
                      if (row.queueStatus === 'queued') {
                        await activateNextItem(currentRoomId)
                      }
                      await startSession(row.sessionId)
                      message.success('竞拍已开始')
                      await refreshGoods()
                    } catch (error) {
                        const nextMessage =
                          error instanceof Error ? error.message : '开始竞拍失败，请稍后重试'
                        message.error(nextMessage)
                      }
                    }}
                  >
                    开始竞拍
                  </Button>
                  <Button
                    size="small"
                    type="primary"
                    ghost
                    onClick={async () => {
                      setSelectedGoods(row)
                      if (row.sessionStatus === 'bidding') {
                        try {
                          const result = await settleSession(row.sessionId)
                          setResultStatus(result.status === 'ended_sold' ? '成交' : '流拍')
                          setResultOpen(true)
                          await refreshGoods()
                        } catch (error) {
                          const nextMessage =
                            error instanceof Error ? error.message : '结算失败，请稍后重试'
                          message.error(nextMessage)
                        }
                        return
                      }

                      setResultStatus(row.displayStatus === '已成交' ? '成交' : row.displayStatus === '已取消' ? '异常取消' : '流拍')
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
        onSave={async (values) => {
          if (!selectedGoods) {
            return
          }

          try {
            await updateItem(selectedGoods.itemId, {
              startPrice: values.startPrice,
              incrementStep: values.incrementStep,
              ceilingPrice: typeof values.ceilingPrice === 'number' ? values.ceilingPrice : null,
              durationSeconds: values.durationSeconds,
              extensionSeconds: values.extensionSeconds,
              extensionTriggerSeconds: 30,
            })
            message.success('规则已保存')
            setEditOpen(false)
            await refreshGoods()
          } catch (error) {
            const nextMessage =
              error instanceof Error ? error.message : '保存规则失败，请稍后重试'
            message.error(nextMessage)
          }
        }}
      />
      <RuleConfigDrawer open={rulePanelOpen} onClose={() => setRulePanelOpen(false)} />
      <RuleModal open={ruleModalOpen} onClose={() => setRuleModalOpen(false)} />
      <ExceptionCancelModal
        open={cancelOpen}
        goodsName={selectedGoods?.title ?? ''}
        onClose={() => setCancelOpen(false)}
        onConfirm={async () => {
          if (!selectedGoods?.sessionId) {
            return
          }

          try {
            await cancelSession(selectedGoods.sessionId)
            setCancelOpen(false)
            setResultStatus('异常取消')
            setResultOpen(true)
            await refreshGoods()
          } catch (error) {
            const nextMessage =
              error instanceof Error ? error.message : '取消竞拍失败，请稍后重试'
            message.error(nextMessage)
          }
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
