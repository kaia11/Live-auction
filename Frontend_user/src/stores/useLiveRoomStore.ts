import { create } from 'zustand'
import { LiveRoom, AuctionItem, AuctionItemStatus, RankingItem, MyBidStatus, BidHistory } from '@/types'

interface LiveRoomState {
  rooms: LiveRoom[]
  currentRoomId: string | null
  items: AuctionItem[]
  currentItemId: string | null
  top3Ranking: RankingItem[]
  myBidStatus: MyBidStatus
  bidHistories: BidHistory[]
  onlineCount: number
  currentCountdown: number
  isCurrentAuctionCardClosed: boolean
  showAuctionItemDrawer: boolean
  showBidPanel: boolean
  showRuleModal: boolean
  showBidSuccessModal: boolean
  showOvertakenModal: boolean
  showAuctionEndPanel: boolean
  showDelayBanner: boolean
  setCurrentRoomId: (roomId: string) => void
  setCurrentItemId: (itemId: string) => void
  loadMockData: () => void
  toggleAuctionItemDrawer: () => void
  toggleBidPanel: () => void
  toggleRuleModal: () => void
  closeAllModals: () => void
  setCurrentAuctionCardClosed: (closed: boolean) => void
  submitBid: (price: number) => Promise<boolean>
}

export const useLiveRoomStore = create<LiveRoomState>((set, get) => ({
  rooms: [],
  currentRoomId: null,
  items: [],
  currentItemId: null,
  top3Ranking: [],
  myBidStatus: { myHighestPrice: 0, myRank: 0, isLeading: false },
  bidHistories: [],
  onlineCount: 1234,
  currentCountdown: 300,
  isCurrentAuctionCardClosed: false,
  showAuctionItemDrawer: false,
  showBidPanel: false,
  showRuleModal: false,
  showBidSuccessModal: false,
  showOvertakenModal: false,
  showAuctionEndPanel: false,
  showDelayBanner: false,

  setCurrentRoomId: (roomId) => set({ currentRoomId: roomId }),

  setCurrentItemId: (itemId) => set({ currentItemId: itemId }),

  loadMockData: () => {
    const mockRooms: LiveRoom[] = [
      {
        id: 'room-001',
        title: '翡翠专场直播',
        anchorName: '珠宝大师阿静',
        coverImage: 'https://picsum.photos/seed/jade1/800/500',
        status: 'living',
        onlineCount: 1234,
        thumbnail: 'https://picsum.photos/seed/jade2/200/200',
      },
      {
        id: 'room-002',
        title: '和田玉精品直播',
        anchorName: '玉石老王',
        coverImage: 'https://picsum.photos/seed/jade3/800/500',
        status: 'living',
        onlineCount: 856,
        thumbnail: 'https://picsum.photos/seed/jade4/200/200',
      },
    ]

    const mockItems: AuctionItem[] = [
      {
        id: 'item-001',
        title: '冰种翡翠手镯 58圈口',
        description: '天然A货冰种翡翠，水润通透，58mm标准圈口',
        images: ['https://picsum.photos/seed/gem1/1000/1000'],
        startPrice: 0,
        currentPrice: 28000,
        minIncrement: 1000,
        maxPrice: 100000,
        status: AuctionItemStatus.BIDDING,
        duration: 300,
        extendedSeconds: 0,
      },
      {
        id: 'item-002',
        title: '羊脂白玉挂件 观音',
        description: '新疆和田玉羊脂白玉，大师工观音挂件',
        images: ['https://picsum.photos/seed/gem2/1000/1000'],
        startPrice: 5000,
        currentPrice: 12000,
        minIncrement: 500,
        maxPrice: null,
        status: AuctionItemStatus.COMING_SOON,
        duration: 300,
        extendedSeconds: 0,
      },
      {
        id: 'item-003',
        title: '18K金钻石项链',
        description: '经典四爪镶嵌，主钻0.5克拉',
        images: ['https://picsum.photos/seed/gem3/1000/1000'],
        startPrice: 8000,
        currentPrice: 18500,
        minIncrement: 500,
        maxPrice: 30000,
        status: AuctionItemStatus.ENDED,
        duration: 300,
        extendedSeconds: 0,
      },
    ]

    const mockTop3: RankingItem[] = [
      { userId: 'u1', nickname: '翡翠之王', avatar: '', price: 30000 },
      { userId: 'u2', nickname: '玉石收藏家', avatar: '', price: 29000 },
      { userId: 'u3', nickname: '珠宝爱好者', avatar: '', price: 28500 },
    ]

    const mockHistories: BidHistory[] = [
      { id: 'h1', itemId: 'item-001', itemTitle: '冰种翡翠手镯', itemImage: '', bidPrice: 28000, result: 'win', bidTime: '2024-01-15 20:30' },
      { id: 'h2', itemId: 'item-002', itemTitle: '和田玉把件', itemImage: '', bidPrice: 8500, result: 'lose', bidTime: '2024-01-14 19:00' },
    ]

    set({
      rooms: mockRooms,
      items: mockItems,
      currentItemId: 'item-001',
      top3Ranking: mockTop3,
      bidHistories: mockHistories,
    })
  },

  toggleAuctionItemDrawer: () => set(state => ({ showAuctionItemDrawer: !state.showAuctionItemDrawer })),

  toggleBidPanel: () => set(state => ({ showBidPanel: !state.showBidPanel })),

  toggleRuleModal: () => set(state => ({ showRuleModal: !state.showRuleModal })),

  closeAllModals: () => set({
    showAuctionItemDrawer: false,
    showBidPanel: false,
    showRuleModal: false,
    showBidSuccessModal: false,
    showOvertakenModal: false,
    showAuctionEndPanel: false,
    showDelayBanner: false,
  }),

  setCurrentAuctionCardClosed: (closed) => set({ isCurrentAuctionCardClosed: closed }),

  submitBid: async (price) => {
    await new Promise(resolve => setTimeout(resolve, 800))
    set(state => ({
      myBidStatus: { ...state.myBidStatus, myHighestPrice: price, isLeading: true },
      showBidSuccessModal: true,
    }))
    return true
  },
}))
