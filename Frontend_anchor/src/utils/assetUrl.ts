import { API_BASE_URL } from '@/api/config'
import { localImagePathMap } from '@/assets/localImages'

export const isPersistedCoverImage = (url?: string) => {
  if (!url) {
    return false
  }
  if (url.startsWith('blob:')) {
    return false
  }
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return true
  }
  if (url.startsWith('/uploads/')) {
    return true
  }
  if (localImagePathMap[url]) {
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
  if (localImagePathMap[url]) {
    return localImagePathMap[url]
  }
  if (url.startsWith('/uploads/')) {
    return `${API_BASE_URL}${url}`
  }
  return url
}

export const resolveCoverImage = (coverImage: string | undefined, fallback: string) => {
  if (!isPersistedCoverImage(coverImage)) {
    return fallback
  }
  return resolveAssetUrl(coverImage) ?? fallback
}

export const normalizeCoverImageForSave = (url: string) => {
  if (url.startsWith(`${API_BASE_URL}/`)) {
    return url.slice(API_BASE_URL.length)
  }
  return url
}
