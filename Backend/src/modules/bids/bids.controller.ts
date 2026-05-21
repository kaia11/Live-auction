import { Body, Controller, Post } from '@nestjs/common';
import { CreateBidDto } from './dto/create-bid.dto';
import { BidsService } from './bids.service';

@Controller('bids')
export class BidsController {
  constructor(private readonly bidsService: BidsService) {}

  @Post()
  create(@Body() createBidDto: CreateBidDto) {
    return this.bidsService.create(createBidDto);
  }
}
