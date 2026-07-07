# Video Call — Test Cases

## REST API Endpoints

| Method | Path | Handler | Auth |
|--------|------|---------|------|
| POST | `/api/calls/:callID/video` | ToggleVideo | Token |
| GET | `/api/calls/ice-servers` | GetIceServers | Token |

**WS events (Phase 1)**:

| Event | Direction | Payload | Description |
|-------|-----------|---------|-------------|
| `call:video` | Server→Client | `{call_id, user_id, video_enabled}` | Broadcast when a participant toggles video |
| `call:video_toggle` | Client→Server | `{call_id, video_enabled}` | Request to toggle video (WS path) |

> **Lưu ý**: WS path (`call:video_toggle` trong `ws/client.go`) chỉ hoạt động khi `c.callService != nil`, tức là khi client kết nối qua `/api/calls/ws`. REST endpoint `/api/calls/:callID/video` luôn khả dụng.

---

## 1. ToggleVideo (POST /api/calls/:callID/video)

### 1.1. Success flows

| ID | Scenario | Steps | Expected Result |
|----|----------|-------|-----------------|
| VC-VID-01 | Caller bật video trong video call | 1. A→B video call connected<br>2. POST body `{"video_enabled":true}` với token A | `200` message "đã cập nhật trạng thái video". DB: `video_enabled_caller=true`. WS broadcast `call:video` → A, B. |
| VC-VID-02 | Caller tắt video | 1. A đã bật video (VC-VID-01)<br>2. POST body `{"video_enabled":false}` với token A | `200`. DB: `video_enabled_caller=false`. WS broadcast `call:video`. |
| VC-VID-03 | Callee bật video | POST body `{"video_enabled":true}` với token B | `200`. DB: `video_enabled_callee=true`. WS broadcast. |
| VC-VID-04 | Callee tắt video | POST body `{"video_enabled":false}` với token B | `200`. DB: `video_enabled_callee=false`. WS broadcast. |
| VC-VID-05 | Toggle nhiều lần liên tiếp | caller bật→tắt→bật video | Mỗi lần đều `200`, DB cập nhật đúng, WS broadcast mỗi lần. |

### 1.2. Error flows

| ID | Scenario | Steps | Expected Result |
|----|----------|-------|-----------------|
| VC-VID-06 | Gọi video toggle trên voice call | POST `video_enabled=true` trên call `call_type=voice` | `400` error "cuộc gọi không phải video call". DB không đổi. |
| VC-VID-07 | Call chưa connected (calling) | POST khi status=calling | `400` error "cuộc gọi không ở trạng thái kết nối". |
| VC-VID-08 | Call đã kết thúc | POST khi status=ended | `400` error "cuộc gọi không ở trạng thái kết nối" (hoặc "cuộc gọi không tồn tại"). |
| VC-VID-09 | Không phải participant (user C) | POST với token C | `400` error "không phải người tham gia cuộc gọi". |
| VC-VID-10 | Call không tồn tại | POST với callID giả | `400` error "cuộc gọi không tồn tại". |
| VC-VID-11 | Thiếu `callID` param | POST `/api/calls//video` | `404` (route không match) hoặc `400`. |
| VC-VID-12 | Thiếu `video_enabled` trong body | POST body `{}` | `400` error "video_enabled là bắt buộc". |
| VC-VID-13 | Sai kiểu dữ liệu | POST body `{"video_enabled":"yes"}` | `400` error "video_enabled là bắt buộc" (JSON parse fail). |
| VC-VID-14 | Không có token | POST không gửi Authorization | `401` Unauthorized (AuthMiddleware). |

### 1.3. State machine validation

| ID | Scenario | Steps | Expected Result |
|----|----------|-------|-----------------|
| VC-VID-15 | Gọi cho chính mình → video call → toggle | Initiate call self → video | `400` (không thể gọi mình) — không đến được toggle. |
| VC-VID-16 | Caller có active call khác | A đang trong call với C, toggle video call vs B | `400` (đang có cuộc gọi khác). |
| VC-VID-17 | Video toggle sau khi end call | connected → end call → toggle video | `400` (call không tồn tại hoặc đã kết thúc). |

---

## 2. GetIceServers (GET /api/calls/ice-servers)

| ID | Scenario | Config | Expected Result |
|----|----------|--------|-----------------|
| VC-ICE-01 | Có STUN, không TURN | `ICE_SERVER_URLS=stun:stun.l.google.com:19302`, TURN để trống | `200` `{"ice_servers":[{"urls":"stun:stun.l.google.com:19302"}]}` |
| VC-ICE-02 | Nhiều STUN | `ICE_SERVER_URLS=stun:a:1,stun:b:2,stun:c:3` | `200` 3 STUN servers |
| VC-ICE-03 | STUN + TURN | STUN + `TURN_SERVER_URL`, `TURN_USERNAME`, `TURN_CREDENTIAL` | `200` list STUN + TURN có credentials |
| VC-ICE-04 | Không có config nào | Tất cả để trống | `200` `{"ice_servers":[]}` |
| VC-ICE-05 | STUN có khoảng trắng | `ICE_SERVER_URLS= stun:a:1 , stun:b:2 ` | `200` 2 server, URLs đã trim |
| VC-ICE-06 | Empty entries | `ICE_SERVER_URLS=stun:a:1,,stun:b:2` | `200` 2 server (skip empty) |
| VC-ICE-07 | Yêu cầu auth | Gọi không token | `401` Unauthorized (AuthMiddleware) |
| VC-ICE-08 | TURN thiếu username | Có TURN_URL nhưng username rỗng | `200` TURN vẫn trả về với username="" |
| VC-ICE-09 | Invalid URL format | `ICE_SERVER_URLS=google.com` (thiếu prefix stun:) | `200` — server không validate format |

---

## 3. WS Events

### 3.1. call:video (Server → Client)

| ID | Scenario | Steps | Expected WS |
|----|----------|-------|-------------|
| VC-WS-01 | Caller bật video | REST toggle caller=true | A, B nhận `{"type":"call:video","payload":{"call_id":"...","user_id":"caller-uuid","video_enabled":true}}` |
| VC-WS-02 | Callee tắt video | REST toggle callee=false | A, B nhận `{"type":"call:video","payload":{"call_id":"...","user_id":"callee-uuid","video_enabled":false}}` |
| VC-WS-03 | WS không kết nối | Toggle qua REST khi 1 user offline | User online vẫn nhận WS broadcast (xử lý bên trong hub bỏ qua offline user) |
| VC-WS-04 | Nhận đúng call | Cùng lúc 2 call khác nhau, toggle call-1 | Chỉ participant call-1 nhận event |

### 3.2. call:video_toggle (Client → Server) — WS path

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| VC-WS-05 | Gửi toggle qua WS (callService != nil) | Server gọi `callService.ToggleVideo`, broadcast `call:video` |
| VC-WS-06 | Gửi toggle qua WS nhánh chat (service != nil) | Không liên quan — event đi qua `call:video_toggle` case, check `callService != nil` |
| VC-WS-07 | Payload thiếu `call_id` | Error "dữ liệu video toggle không hợp lệ" |
| VC-WS-08 | Payload thiếu `video_enabled` | Error "dữ liệu video toggle không hợp lệ" |

---

## 4. WebRTC Integration

| ID | Scenario | Steps | Expected Result |
|----|----------|-------|-----------------|
| VC-WEBRTC-01 | Khởi tạo video call từ frontend | 1. Fetch ICE servers từ `/api/calls/ice-servers` (cần token)<br>2. Tạo RTCPeerConnection với ICE servers<br>3. Initiate video call | Peer connection created với video track |
| VC-WEBRTC-02 | Video toggle từ frontend | 1. Đang trong video call connected<br>2. User bấm nút tắt camera<br>3. POST `/api/calls/:callID/video` `{video_enabled:false}`<br>4. Frontend disable local video track | Local video track disabled, remote peer nhận `call:video` và update UI |
| VC-WEBRTC-03 | ICE server fallback | Server trả về `{"ice_servers":[]}` | Browser tự dùng STUN mặc định, vẫn kết nối được trong LAN |
| VC-WEBRTC-04 | TURN relay | Cấu hình TURN, peer ở 2 mạng khác nhau | Kết nối qua TURN relay (candidate type `relay`) |

---

## 5. Database schema

Migration cần thêm 2 cột vào bảng `calls`:

```sql
ALTER TABLE calls 
  ADD COLUMN video_enabled_caller BOOLEAN DEFAULT FALSE,
  ADD COLUMN video_enabled_callee BOOLEAN DEFAULT FALSE;
```

Kiểm tra:

```sql
-- Verify 2 cột mới
SELECT video_enabled_caller, video_enabled_callee FROM calls WHERE id = '<call_id>';

-- Verify default value với call cũ
SELECT COUNT(*) FROM calls WHERE video_enabled_caller IS NULL;
```

---

## 6. Edge cases tổng hợp

| ID | Scenario | Steps | Expected |
|----|----------|-------|----------|
| VC-EDGE-01 | Toggle video → ngay lập tức end call | Toggle caller=true → end call ngay sau đó | Dữ liệu DB: video_enabled_caller=true. Call đã ended. |
| VC-EDGE-02 | Cả 2 cùng toggle cùng lúc | A và B cùng POST toggle gần đồng thời | Cả 2 DB cập nhật đúng (cập nhật khác cột). WS broadcast 2 lần. |
| VC-EDGE-03 | Voice call upgrade? | Hiện tại không có endpoint upgrade call_type | Chỉ cho toggle trên `call_type=video` |
| VC-EDGE-04 | Concurrent toggle | Gửi 10 request toggle liên tiếp cùng lúc | Tất cả 200, DB ghi nhận giá trị cuối cùng |
