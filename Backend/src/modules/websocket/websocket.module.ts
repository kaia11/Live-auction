import { Module } from '@nestjs/common';
import { AuctionGateway } from './websocket.gateway';

@Module({
  providers: [AuctionGateway],
  exports: [AuctionGateway]
})
export class WebsocketModule {}
