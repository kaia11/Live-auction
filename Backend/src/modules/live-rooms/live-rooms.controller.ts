import { Controller, Get, Param } from '@nestjs/common';
import { LiveRoomsService } from './live-rooms.service';

@Controller('rooms')
export class LiveRoomsController {
  constructor(private readonly liveRoomsService: LiveRoomsService) {}

  @Get()
  findAll() {
    return this.liveRoomsService.findAll();
  }

  @Get(':roomId')
  findOne(@Param('roomId') roomId: string) {
    return this.liveRoomsService.findOne(roomId);
  }
}
