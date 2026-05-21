import { Injectable } from '@nestjs/common';

@Injectable()
export class UsersService {
  getMockProfile() {
    return {
      id: 'mock-user-1',
      nickname: 'demo-user',
      role: 'user'
    };
  }
}
