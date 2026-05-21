import { Injectable } from '@nestjs/common';

@Injectable()
export class AuctionSessionsService {
  getCurrentSnapshot(roomId: string) {
    return {
      roomId,
      sessionId: 'session-1',
      itemId: 'item-1',
      status: 'bidding',
      currentPrice: 850
    };
  }
}
