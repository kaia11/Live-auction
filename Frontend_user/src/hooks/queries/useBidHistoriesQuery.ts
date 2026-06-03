import { useQuery } from '@tanstack/react-query'
import { getMyBidHistories } from '@/api/bids'
import { mapBidHistories } from '@/adapters/auction'

export const bidHistoriesQueryKey = ['my-bid-histories']

export const useBidHistoriesQuery = () =>
  useQuery({
    queryKey: bidHistoriesQueryKey,
    queryFn: async () => {
      const histories = await getMyBidHistories()
      return mapBidHistories(histories)
    },
  })

