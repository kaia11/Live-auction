import { Injectable } from '@nestjs/common';

@Injectable()
export class AuctionItemsService {
  findByRoom(roomId: string) {
    return [
      {
        id: 'item-1',
        roomId,
        title: 'Hetian Jade Pendant',
        status: 'bidding'
      }
    ];
  }
}
