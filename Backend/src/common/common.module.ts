import { Module } from '@nestjs/common';
import { MockAuthGuard } from './guards/mock-auth.guard';

@Module({
  providers: [MockAuthGuard],
  exports: [MockAuthGuard]
})
export class CommonModule {}
