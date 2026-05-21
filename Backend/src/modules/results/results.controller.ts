import { Controller, Get, Param } from '@nestjs/common';
import { ResultsService } from './results.service';

@Controller('rooms/:roomId/results')
export class ResultsController {
  constructor(private readonly resultsService: ResultsService) {}

  @Get(':itemId')
  findOne(@Param('roomId') roomId: string, @Param('itemId') itemId: string) {
    return this.resultsService.findOne(roomId, itemId);
  }
}
