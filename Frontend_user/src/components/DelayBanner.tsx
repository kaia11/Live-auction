import React from 'react'
import { useLiveRoomStore } from '../stores/useLiveRoomStore'
import './DelayBanner.scss'

const DelayBanner: React.FC = () => {
  const { items, currentItemId } = useLiveRoomStore()
  const item = items.find((entry) => entry.id === currentItemId)
  const extensionSeconds = item?.extendedSeconds ?? 0

  return (
    <div className="delay-banner">
      ⏰ 倒计时自动延长{extensionSeconds}秒，继续出价!
    </div>
  )
}

export default DelayBanner
