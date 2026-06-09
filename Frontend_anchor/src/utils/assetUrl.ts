import { API_BASE_URL } from '@/api/config'

export const resolveAssetUrl = (url?: string) => {
  if (!url) {
    return url
  }
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('blob:')) {
    return url
  }
  if (url.startsWith('/')) {
    return `${API_BASE_URL}${url}`
  }
  return url
}

export const resolveCoverImage = (coverImage: string | undefined, fallback: string) => {
  if (coverImage && !coverImage.startsWith('blob:')) {
    return resolveAssetUrl(coverImage) ?? fallback
  }
  return fallback
}
