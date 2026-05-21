import { Injectable } from '@nestjs/common';

@Injectable()
export class LiveRoomsService {
  findAll() {
    return [
      {
        id: 'room-1',
        title: 'Jade Auction Room',
        status: 'live'
      }
    ];
  }

  findOne(roomId: string) {
    return {
      id: roomId,
      title: 'Jade Auction Room',
      status: 'live'
    };
  }
}
