import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { ScheduleModule } from '@nestjs/schedule';
import { AuctionItemsModule } from './modules/auction-items/auction-items.module';
import { AuctionSessionsModule } from './modules/auction-sessions/auction-sessions.module';
import { BidsModule } from './modules/bids/bids.module';
import { CommonModule } from './common/common.module';
import { LiveRoomsModule } from './modules/live-rooms/live-rooms.module';
import { LogsModule } from './modules/logs/logs.module';
import { PrismaModule } from './prisma/prisma.module';
import { RedisModule } from './redis/redis.module';
import { ResultsModule } from './modules/results/results.module';
import { UsersModule } from './modules/users/users.module';
import { WebsocketModule } from './modules/websocket/websocket.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env'
    }),
    ScheduleModule.forRoot(),
    CommonModule,
    PrismaModule,
    RedisModule,
    UsersModule,
    LiveRoomsModule,
    AuctionItemsModule,
    AuctionSessionsModule,
    BidsModule,
    ResultsModule,
    WebsocketModule,
    LogsModule
  ]
})
export class AppModule {}
