import coverPrimary from '../../images/generated-1779436183124.png'
import coverSecondary from '../../images/generated-1779437166813.png'

const gallery = [coverPrimary, coverSecondary]

export const localImagePathMap: Record<string, string> = {
  '/images/generated-1779436183124.png': coverPrimary,
  '/images/generated-1779437166813.png': coverSecondary,
}

export const getLocalGalleryImage = (seed?: string) => {
  if (!seed) {
    return gallery[0]
  }

  let hash = 0
  for (let i = 0; i < seed.length; i += 1) {
    hash = (hash * 31 + seed.charCodeAt(i)) >>> 0
  }

  return gallery[hash % gallery.length]
}

export const getMerchantItemImage = (itemId?: string) => getLocalGalleryImage(`merchant:item:${itemId ?? 'default'}`)
export const getMerchantRoomImage = (roomId?: string) => getLocalGalleryImage(`merchant:room:${roomId ?? 'default'}`)

