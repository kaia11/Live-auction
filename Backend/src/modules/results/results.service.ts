import { Injectable } from '@nestjs/common';

@Injectable()
export class ResultsService {
  findOne(roomId: string, itemId: string) {
    return {
      roomId,
      itemId,
      resultStatus: 'pending'
    };
  }
}
