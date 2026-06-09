import { API_BASE_URL } from '@/api/config'

export const isPersistedCoverImage = (url?: string) => {
  if (!url) {
    return false
  }
  if (url.startsWith('blob:')) {
    return false
  }
  if (url.startsWith('/images/')) {
    return false
  }
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return true
  }
  if (url.startsWith('/uploads/')) {
    return true
  }
  return false
}

export const resolveAssetUrl = (url?: string) => {
  if (!url) {
    return url
  }
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  if (url.startsWith('/uploads/')) {
    return `${API_BASE_URL}${url}`
  }
  return url
}

export const resolveItemCoverImage = (coverImage: string | undefined, fallback: string) => {
  if (!isPersistedCoverImage(coverImage)) {
    return fallback
  }
  return resolveAssetUrl(coverImage) ?? fallback
}
