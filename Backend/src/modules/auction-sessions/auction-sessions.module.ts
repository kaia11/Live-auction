import { Module } from '@nestjs/common';
import { AuctionSessionsController } from './auction-sessions.controller';
import { AuctionSessionsService } from './auction-sessions.service';

@Module({
  controllers: [AuctionSessionsController],
  providers: [AuctionSessionsService],
  exports: [AuctionSessionsService]
})
export class AuctionSessionsModule {}
