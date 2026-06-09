import { Image } from 'antd'
import type { ImageProps } from 'antd'
import { useEffect, useState } from 'react'
import { getMerchantItemImage } from '@/assets/localImages'
import { resolveCoverImage } from '@/utils/assetUrl'

interface GoodsCoverImageProps extends ImageProps {
  coverImage?: string
  itemId: string
}

export function GoodsCoverImage({ coverImage, itemId, ...props }: GoodsCoverImageProps) {
  const fallback = getMerchantItemImage(itemId)
  const [src, setSrc] = useState(() => resolveCoverImage(coverImage, fallback))

  useEffect(() => {
    setSrc(resolveCoverImage(coverImage, fallback))
  }, [coverImage, fallback])

  return (
    <Image
      {...props}
      src={src}
      onError={() => {
        if (src !== fallback) {
          setSrc(fallback)
        }
      }}
    />
  )
}
