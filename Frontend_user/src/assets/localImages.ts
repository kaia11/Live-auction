import unifiedPreviewImage from '../../images/Image.png'

export const getLocalGalleryImage = (_seed?: string) => {
  return unifiedPreviewImage
}

export const getRoomCoverImage = (roomId: string) => getLocalGalleryImage(`room:${roomId}`)
export const getRoomThumbnailImage = (roomId: string) => getLocalGalleryImage(`thumb:${roomId}`)
export const getItemCoverImage = (itemId: string) => getLocalGalleryImage(`item:${itemId}`)
export const getHistoryImage = (itemId: string) => getLocalGalleryImage(`history:${itemId}`)
