import { Controller, Get, Param } from '@nestjs/common';
import { AuctionItemsService } from './auction-items.service';

@Controller('rooms/:roomId/items')
export class AuctionItemsController {
  constructor(private readonly auctionItemsService: AuctionItemsService) {}

  @Get()
  findByRoom(@Param('roomId') roomId: string) {
    return this.auctionItemsService.findByRoom(roomId);
  }
}
