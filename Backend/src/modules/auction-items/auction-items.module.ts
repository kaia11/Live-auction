import { Module } from '@nestjs/common';
import { AuctionItemsController } from './auction-items.controller';
import { AuctionItemsService } from './auction-items.service';

@Module({
  controllers: [AuctionItemsController],
  providers: [AuctionItemsService],
  exports: [AuctionItemsService]
})
export class AuctionItemsModule {}
