USE auction_live;

-- Demo seed passwords are inserted as plaintext bootstrap values and
-- migrated to bcrypt hashes during backend startup before auth is served.
INSERT INTO users (id, username, password, nickname, avatar, role)
VALUES
  ('user-001', 'viewer_demo', '123456', '阿宁', 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&q=80', 'viewer'),
  ('user-002', 'viewer_guest', '123456', '小满', 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200&q=80', 'viewer'),
  ('user-003', 'viewer_vip', '123456', '阿青', 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=200&q=80', 'viewer'),
  ('anchor-001', 'anchor_admin', '123456', '主播小玉', 'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=200&q=80', 'anchor'),
  ('admin-001', 'admin_root', '123456', '运营管理员', 'https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?w=200&q=80', 'admin')
ON DUPLICATE KEY UPDATE username = VALUES(username), password = VALUES(password), nickname = VALUES(nickname), avatar = VALUES(avatar), role = VALUES(role);

INSERT INTO live_rooms (
  id, title, cover_image, video_url, status, anchor_user_id, anchor_name, online_count, thumbnail, current_session_id
)
VALUES (
  'room-001',
  '古风首饰直播竞拍间',
  'https://images.unsplash.com/photo-1617038220319-276d3cfab638?w=1200&q=80',
  'https://www.w3schools.com/html/mov_bbb.mp4',
  'live',
  'anchor-001',
  '主播小玉',
  1288,
  'https://images.unsplash.com/photo-1512436991641-6745cdb1723f?w=200&q=80',
  'session-001'
)
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  cover_image = VALUES(cover_image),
  video_url = VALUES(video_url),
  status = VALUES(status),
  anchor_user_id = VALUES(anchor_user_id),
  anchor_name = VALUES(anchor_name),
  online_count = VALUES(online_count),
  thumbnail = VALUES(thumbnail),
  current_session_id = VALUES(current_session_id);

INSERT INTO auction_items (
  id, room_id, title, cover_image, description, start_price, increment_step, ceiling_price,
  duration_seconds, extension_seconds, extension_trigger_seconds, queue_status
)
VALUES
  (
    'item-001',
    'room-001',
    '和田玉吊坠',
    'https://images.unsplash.com/photo-1512436991641-6745cdb1723f?w=800&q=80',
    '直播竞拍样例拍品，当前使用假视频占位进行联调。',
    0, 5, 999,
    120, 30, 30, 'active'
  ),
  (
    'item-002',
    'room-001',
    '鎏金花丝耳坠',
    'https://images.unsplash.com/photo-1617038220319-276d3cfab638?w=800&q=80',
    '待上场拍品，用于后续拍品队列与切场开发。',
    0, 10, NULL,
    120, 30, 30, 'queued'
  )
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  cover_image = VALUES(cover_image),
  description = VALUES(description),
  start_price = VALUES(start_price),
  increment_step = VALUES(increment_step),
  ceiling_price = VALUES(ceiling_price),
  duration_seconds = VALUES(duration_seconds),
  extension_seconds = VALUES(extension_seconds),
  extension_trigger_seconds = VALUES(extension_trigger_seconds),
  queue_status = VALUES(queue_status);

INSERT INTO room_item_queue (room_id, item_id, sort_order)
VALUES
  ('room-001', 'item-001', 1),
  ('room-001', 'item-002', 2)
ON DUPLICATE KEY UPDATE sort_order = VALUES(sort_order);

INSERT INTO auction_sessions (
  id, room_id, item_id, status, current_price, leader_user_id, end_time,
  participant_count, increment_step, extension_seconds, extension_trigger_seconds,
  ceiling_price, supports_auto_proxy
)
VALUES
  (
    'session-001',
    'room-001',
    'item-001',
    'bidding',
    135,
    'user-003',
    DATE_ADD(NOW(), INTERVAL 2 MINUTE),
    3,
    5,
    30,
    30,
    999,
    1
  ),
  (
    'session-002',
    'room-001',
    'item-002',
    'pending',
    0,
    '',
    NULL,
    0,
    10,
    30,
    30,
    NULL,
    1
  )
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  current_price = VALUES(current_price),
  leader_user_id = VALUES(leader_user_id),
  end_time = VALUES(end_time),
  participant_count = VALUES(participant_count),
  increment_step = VALUES(increment_step),
  extension_seconds = VALUES(extension_seconds),
  extension_trigger_seconds = VALUES(extension_trigger_seconds),
  ceiling_price = VALUES(ceiling_price),
  supports_auto_proxy = VALUES(supports_auto_proxy);

INSERT INTO bids (id, session_id, room_id, item_id, user_id, bid_price, request_id, rank_after, status, create_time)
VALUES
  ('bid-001', 'session-001', 'room-001', 'item-001', 'user-001', 125, 'req-001', 3, 'accepted', DATE_SUB(NOW(), INTERVAL 90 SECOND)),
  ('bid-002', 'session-001', 'room-001', 'item-001', 'user-002', 130, 'req-002', 2, 'accepted', DATE_SUB(NOW(), INTERVAL 60 SECOND)),
  ('bid-003', 'session-001', 'room-001', 'item-001', 'user-003', 135, 'req-003', 1, 'accepted', DATE_SUB(NOW(), INTERVAL 30 SECOND))
ON DUPLICATE KEY UPDATE
  bid_price = VALUES(bid_price),
  rank_after = VALUES(rank_after),
  status = VALUES(status),
  create_time = VALUES(create_time);
