import {
  ConnectedSocket,
  MessageBody,
  OnGatewayConnection,
  OnGatewayDisconnect,
  SubscribeMessage,
  WebSocketGateway,
  WebSocketServer
} from '@nestjs/websockets';
import { Server, Socket } from 'socket.io';

@WebSocketGateway({
  cors: {
    origin: '*'
  }
})
export class AuctionGateway
  implements OnGatewayConnection, OnGatewayDisconnect
{
  @WebSocketServer()
  server: Server;

  handleConnection(client: Socket) {
    client.emit('connected', {
      socketId: client.id
    });
  }

  handleDisconnect(client: Socket) {
    client.leaveAll();
  }

  @SubscribeMessage('join_room')
  handleJoinRoom(
    @ConnectedSocket() client: Socket,
    @MessageBody() payload: { roomId: string }
  ) {
    client.join(payload.roomId);
    client.emit('room_joined', payload);
  }

  @SubscribeMessage('leave_room')
  handleLeaveRoom(
    @ConnectedSocket() client: Socket,
    @MessageBody() payload: { roomId: string }
  ) {
    client.leave(payload.roomId);
    client.emit('room_left', payload);
  }

  broadcastPriceUpdate(roomId: string, data: Record<string, unknown>) {
    this.server.to(roomId).emit('auction_price_updated', data);
  }
}
