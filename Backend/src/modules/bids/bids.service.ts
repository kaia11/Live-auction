import { Injectable } from '@nestjs/common';
import { CreateBidDto } from './dto/create-bid.dto';

@Injectable()
export class BidsService {
  create(createBidDto: CreateBidDto) {
    return {
      accepted: true,
      message: 'Bid request accepted by skeleton service',
      payload: createBidDto
    };
  }
}
