const STORAGE_KEY = 'live-auction-paid-deposits-v2'

type DepositMap = Record<string, true>

const getKey = (userId: string, itemId: string) => `${userId}::${itemId}`

const readStore = (): DepositMap => {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return {}
    }
    const parsed = JSON.parse(raw) as DepositMap
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

const writeStore = (payload: DepositMap) => {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
  } catch {
    // Ignore storage write errors to avoid blocking bidding flow.
  }
}

export const hasPaidDeposit = (userId: string, itemId: string) => {
  const store = readStore()
  return !!store[getKey(userId, itemId)]
}

export const markDepositPaid = (userId: string, itemId: string) => {
  const store = readStore()
  store[getKey(userId, itemId)] = true
  writeStore(store)
}

export const clearDepositPaid = (userId: string, itemId: string) => {
  const store = readStore()
  delete store[getKey(userId, itemId)]
  writeStore(store)
}

export const needsDepositPayment = (
  userId: string | undefined,
  itemId: string | undefined,
  depositAmount: number,
) => depositAmount > 0 && !!userId && !!itemId && !hasPaidDeposit(userId, itemId)
