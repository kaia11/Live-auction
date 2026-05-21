import { Controller, Get, Param } from '@nestjs/common';
import { AuctionSessionsService } from './auction-sessions.service';

@Controller('rooms/:roomId/current')
export class AuctionSessionsController {
  constructor(
    private readonly auctionSessionsService: AuctionSessionsService
  ) {}

  @Get()
  getCurrentSnapshot(@Param('roomId') roomId: string) {
    return this.auctionSessionsService.getCurrentSnapshot(roomId);
  }
}
