import { IsNumber, IsString, Min } from 'class-validator';

export class CreateBidDto {
  @IsString()
  roomId: string;

  @IsString()
  itemId: string;

  @IsString()
  sessionId: string;

  @IsString()
  requestId: string;

  @IsNumber()
  @Min(0)
  bidPrice: number;
}
