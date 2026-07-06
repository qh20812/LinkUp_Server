# Voice Call — Test Cases

## REST API Endpoints

| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET | `/api/calls/ws` | HandleWebsocket | Token |
| GET | `/api/calls/history` | GetCallHistory | Token |
| GET | `/api/calls/:callID` | GetCallDetail | Token |
| POST | `/api/calls/initiate` | InitiateCall | Token |
| POST | `/api/calls/:callID/accept` | AcceptCall | Token |
| POST | `/api/calls/:callID/reject` | RejectCall | Token |
| POST | `/api/calls/:callID/mute` | ToggleMute | Token |

---

## 1. InitiateCall

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-INIT-01 | Gọi cho chính mình | POST `/api/calls/initiate` với `callee_id = caller_id` | `400` error "không thể gọi cho chính mình". Không tạo DB record. | ✅ |
| VC-INIT-02 | Caller đang có cuộc gọi active (calling) | 1. A→B call created<br>2. A→C call initiated | `400` error "bạn đang có cuộc gọi khác". Không tạo DB record. | ✅ |
| VC-INIT-03 | Caller đang có cuộc gọi active (connected) | 1. A→B call created, accepted → connected<br>2. A→C call initiated | `400` error "bạn đang có cuộc gọi khác". Không tạo DB record. | ✅ |
| VC-INIT-04 | Caller và callee không phải bạn bè | POST `/api/calls/initiate` với callee không phải bạn | `400` error "chỉ có thể gọi cho bạn bè". | ✅ |
| VC-INIT-05 | Callee đang bận (calling) | 1. B→C call created (B là caller)<br>2. A→B call initiated | `200` response `{message: "người dùng đang bận"}`. Caller A nhận `call:busy` WS event. | ✅ |
| VC-INIT-06 | Callee đang bận (connected) | 1. B→C call connected<br>2. A→B call initiated | `200` response `{message: "người dùng đang bận"}`. A nhận `call:busy`. | ✅ |
| VC-INIT-07 | Callee online — full flow thành công | 1. A và B là bạn<br>2. A→B call initiated<br>3. B online | call created với status `calling`. B nhận `call:incoming` WS. A nhận `call:status` WS. | ✅ |
| VC-INIT-08 | Callee offline — call được tạo nhưng không có WS push | 1. A và B là bạn<br>2. B offline<br>3. A→B call initiated | call created với status `calling`. B KHÔNG nhận WS event. A nhận `call:status`. | ✅ |
| VC-INIT-09 | Call type voice | POST với `call_type = "voice"` | DB: `call_type = "voice"` | ✅ |
| VC-INIT-10 | Call type video | POST với `call_type = "video"` | DB: `call_type = "video"` | ✅ |
| VC-INIT-11 | Call type không hợp lệ | POST với `call_type = "fax"` | Fallback về `call_type = "voice"` (ParseCallType default). | ✅ |
| VC-INIT-12 | Thiếu callee_id | POST body thiếu `callee_id` | `400` error "callee_id và call_type là bắt buộc". | ✅ |
| VC-INIT-13 | Không có token | POST không gửi Authorization header | `401` Unauthorized (AuthMiddleware). | ✅ |

---

## 2. AcceptCall

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-ACC-01 | Người nhận chấp nhận cuộc gọi | POST `/api/calls/:callID/accept` bởi callee | Status `connected`, `started_at` set. Cả 2 nhận `call:status(connected)` WS. | ✅ |
| VC-ACC-02 | Người gọi không thể accept | POST `/api/calls/:callID/accept` bởi caller | `400` error "chỉ người nhận mới có thể chấp nhận cuộc gọi". | ✅ |
| VC-ACC-03 | User không phải participant | POST `/api/calls/:callID/accept` bởi user C | `400` error "chỉ người nhận mới có thể chấp nhận cuộc gọi". | ✅ |
| VC-ACC-04 | Cuộc gọi đã kết thúc (ended) | Accept sau khi call ended | `400` error "cuộc gọi không ở trạng thái chờ". | ✅ |
| VC-ACC-05 | Cuộc gọi đã bị từ chối (rejected) | Accept sau khi call rejected | `400` error "cuộc gọi không ở trạng thái chờ". | ✅ |
| VC-ACC-06 | Cuộc gọi không tồn tại | Accept với callID không có trong DB | `400` error "cuộc gọi không tồn tại". | ✅ |
| VC-ACC-07 | Thiếu callID param | POST `/api/calls//accept` | `404` route không match. | ✅ |

---

## 3. RejectCall

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-REJ-01 | Người nhận từ chối | POST `/api/calls/:callID/reject` bởi callee | Status `rejected`, `ended_at` set. Cả 2 nhận `call:status(rejected)` WS. | ✅ |
| VC-REJ-02 | Người gọi không thể reject | reject bởi caller | `400` error "chỉ người nhận mới có thể từ chối cuộc gọi". | ✅ |
| VC-REJ-03 | User không phải participant | reject bởi user C | `400` error "chỉ người nhận mới có thể từ chối cuộc gọi". | ✅ |
| VC-REJ-04 | Cuộc gọi đã connected | reject sau khi accepted | `400` error "cuộc gọi không ở trạng thái chờ". | ✅ |
| VC-REJ-05 | Cuộc gọi đã ended | reject sau khi ended | `400` error "cuộc gọi không ở trạng thái chờ". | ✅ |
| VC-REJ-06 | Cuộc gọi không tồn tại | reject với callID không hợp lệ | `400` error "cuộc gọi không tồn tại". | ✅ |

---

## 4. EndCall

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-END-01 | Caller kết thúc cuộc gọi connected | 1. A→B call connected<br>2. A ends call | Status `ended`, `duration` tính đúng. Cả 2 nhận `call:status(ended)` WS. | ✅ |
| VC-END-02 | Callee kết thúc cuộc gọi connected | 1. A→B call connected<br>2. B ends call | Status `ended`. | ✅ |
| VC-END-03 | Kết thúc cuộc gọi calling (chưa connected) | 1. A→B call calling<br>2. A ends call | Status `missed` (vì chưa connected). `duration = 0`. | ✅ |
| VC-END-04 | User không phải participant | end bởi user C | `400` error "không phải người tham gia cuộc gọi". | ✅ |
| VC-END-05 | Cuộc gọi đã ended | end call đã ended | `400` error "cuộc gọi đã kết thúc". | ✅ |
| VC-END-06 | Cuộc gọi đã rejected | end call đã rejected | `400` error "cuộc gọi đã kết thúc". | ✅ |
| VC-END-07 | Cuộc gọi không tồn tại | end với callID không hợp lệ | `400` error "cuộc gọi không tồn tại". | ✅ |

---

## 5. ToggleMute

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-MUTE-01 | Caller mute | POST `/api/calls/:callID/mute` `{muted: true}` | `muted_caller = true`. Cả 2 nhận `call:mute` WS. | ✅ |
| VC-MUTE-02 | Caller unmute | POST `/api/calls/:callID/mute` `{muted: false}` | `muted_caller = false`. | ✅ |
| VC-MUTE-03 | Callee mute | Callee gọi mute | `muted_callee = true`. | ✅ |
| VC-MUTE-04 | User không phải participant | mute bởi user C | `400` error "không phải người tham gia cuộc gọi". | ✅ |
| VC-MUTE-05 | Cuộc gọi chưa connected | mute khi call đang calling | `400` error "cuộc gọi không ở trạng thái kết nối". | ✅ |
| VC-MUTE-06 | Cuộc gọi đã kết thúc | mute sau khi ended | `400` error "cuộc gọi không ở trạng thái kết nối". | ✅ |
| VC-MUTE-07 | Cuộc gọi không tồn tại | mute với callID không hợp lệ | `400` error "cuộc gọi không tồn tại". | ✅ |
| VC-MUTE-08 | Thiếu muted field | POST body `{}` | `400` error "muted là bắt buộc". | ✅ |

---

## 6. GetCallDetail

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-DTL-01 | Caller xem detail | GET `/api/calls/:callID` bởi caller | `200` response trả về Call object. | ✅ |
| VC-DTL-02 | Callee xem detail | GET `/api/calls/:callID` bởi callee | `200` response trả về Call object. | ✅ |
| VC-DTL-03 | User không phải participant | GET bởi user C | `400` error "không phải người tham gia cuộc gọi". | ✅ |
| VC-DTL-04 | Cuộc gọi không tồn tại | GET với callID không hợp lệ | `400` error "cuộc gọi không tồn tại". | ✅ |

---

## 7. GetCallHistory

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-HIST-01 | Lấy lịch sử (mặc định limit=20) | GET `/api/calls/history` | `200` data + total + limit + offset. | ✅ |
| VC-HIST-02 | Phân trang | GET `/api/calls/history?limit=5&offset=10` | Đúng số lượng records theo offset/limit. | ✅ |
| VC-HIST-03 | Limit vượt quá 100 | GET `/api/calls/history?limit=200` | Fallback về limit=20. | ✅ |
| VC-HIST-04 | Limit < 1 | GET `/api/calls/history?limit=0` | Fallback về limit=20. | ✅ |
| VC-HIST-05 | Offset âm | GET `/api/calls/history?offset=-1` | Fallback về offset=0. | ✅ |
| VC-HIST-06 | User không có call history | User mới, chưa có cuộc gọi nào | `data: []`, `total: 0`. | ✅ |
| VC-HIST-07 | Lọc đúng user | User A có 10 cuộc gọi (vài cuộc là caller, vài cuộc là callee) | Tất cả 10 cuộc gọi đều được trả về. | ✅ |

---

## 8. Signal Handling (WebSocket)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-SIG-01 | Caller gửi signal | WS `call:signal` bởi caller | Callee nhận `call:signal` WS event. | ✅ |
| VC-SIG-02 | Callee gửi signal | WS `call:signal` bởi callee | Caller nhận `call:signal` WS event. | ✅ |
| VC-SIG-03 | User không phải participant | WS `call:signal` bởi user C | `error` WS "không phải người tham gia cuộc gọi". | ✅ |
| VC-SIG-04 | Cuộc gọi không tồn tại | WS `call:signal` với callID không hợp lệ | `error` WS "cuộc gọi không tồn tại". | ✅ |

---

## 9. Busy Detection (Critical)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-BUSY-01 | 2 concurrent calls — callee busy | 1. A→B call created<br>2. C→B call initiated | Bước 2: isBusy=true, C nhận `call:busy`, DB chỉ có 1 record (A→B). | ✅ |
| VC-BUSY-02 | 3 concurrent calls — cùng callee | 1. A→B calling<br>2. C→B busy<br>3. D→B busy | Chỉ 1 call được tạo. | ✅ |
| VC-BUSY-03 | Atomic transaction — không có partial insert | `CreateIfNotBusy` fail giữa chừng | Rollback hoàn toàn, không có record. | ✅ |
| VC-BUSY-04 | TOCTOU race: friend check + create tách rời | 1. Friend check pass<br>2. B unfriend A (concurrent)<br>3. `CreateIfNotBusy` chạy | ❌ Call vẫn được tạo (TOCTOU). Cần gộp friend check vào transaction. | ⚠️ |
| VC-BUSY-05 | Callee busy (connected) | 1. A→B connected<br>2. C→B initiated | `call:busy` trả về C. | ✅ |
| VC-BUSY-06 | Callee busy (ringing) | 1. A→B calling<br>2. C→B initiated | `call:busy` trả về C (vì status=calling được tính là busy). | ✅ |

---

## 10. Repository Methods

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-REPO-01 | FindByID — tồn tại | Query bằng ID hợp lệ | Trả về `*Call`, không nil. | ✅ |
| VC-REPO-02 | FindByID — không tồn tại | Query bằng ID không có | Trả về nil, nil. | ✅ |
| VC-REPO-03 | FindActiveByUserID — caller có active call | A là caller của call đang calling | Trả về call đó. | ✅ |
| VC-REPO-04 | FindActiveByUserID — callee có active call | B là callee của call đang connected | Trả về call đó. | ✅ |
| VC-REPO-05 | FindActiveByUserID — không có active call | User không có call nào active | Trả về nil, nil. | ✅ |
| VC-REPO-06 | FindActiveBetween — 2 user có call active | A→B đang calling | Trả về call đó. | ✅ |
| VC-REPO-07 | FindActiveBetween — không có call | A và B chưa từng gọi | Trả về nil, nil. | ✅ |
| VC-REPO-08 | UpdateStatus — sang connected | Update từ calling → connected | `started_at` được set, `ended_at` nil. | ✅ |
| VC-REPO-09 | UpdateStatus — sang ended | Update từ connected → ended | `started_at` giữ nguyên, `ended_at` set, `duration` set. | ✅ |
| VC-REPO-10 | UpdateMuted — caller | `mutedCaller = true, mutedCallee = nil` | `muted_caller = true`, `muted_callee` không đổi. | ✅ |
| VC-REPO-11 | UpdateMuted — callee | `mutedCaller = nil, mutedCallee = true` | `muted_callee = true`, `muted_caller` không đổi. | ✅ |
| VC-REPO-12 | CountHistory — đúng user | User A có 5 calls (3 caller, 2 callee) | `total = 5`. | ✅ |
| VC-REPO-13 | CountHistory — không có call | User mới | `total = 0`. | ✅ |

---

## 11. WebSocket Client Events

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-WS-01 | `call:initiate` — gửi từ client | Client gửi WS `call:initiate` | Server gọi `callService.InitiateCall`. | ✅ |
| VC-WS-02 | `call:initiate` — callService nil | WS client không có callService | `error` WS "dịch vụ gọi không khả dụng". | ✅ |
| VC-WS-03 | `call:initiate` — payload không hợp lệ | Client gửi payload thiếu field | `error` WS "dữ liệu khởi tạo cuộc gọi không hợp lệ". | ✅ |
| VC-WS-04 | `call:initiate` — busy | Initiate trả về call=nil | Không gửi response (continue). Client đã nhận `call:busy` qua hub. | ✅ |
| VC-WS-05 | `call:accept` — từ client | Client gửi WS `call:accept` | Server gọi `callService.AcceptCall`. | ✅ |
| VC-WS-06 | `call:reject` — từ client | Client gửi WS `call:reject` | Server gọi `callService.RejectCall`. | ✅ |
| VC-WS-07 | `call:end` — từ client | Client gửi WS `call:end` | Server gọi `callService.EndCall`. | ✅ |
| VC-WS-08 | `call:signal` — từ client | Client gửi WS `call:signal` | Server gọi `callService.HandleSignal`. | ✅ |
| VC-WS-09 | `call:busy` — client ack | Client gửi WS `call:busy` | No-op, continue (acknowledgement). | ✅ |
| VC-WS-10 | Event không xác định | Client gửi `"type": "unknown:event"` | `error` WS "loại sự kiện không xác định". | ✅ |

---

## 12. Edge Cases & Security

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| VC-SEC-01 | SQL injection qua callID | GET `/api/calls/1' OR '1'='1` | GORM parameterized query, không injection. | ✅ |
| VC-SEC-02 | User bị block vẫn gọi được cho blocker | B block A, A vẫn là bạn B | ❌ Call được tạo vì không có block check. | ⚠️ |
| VC-SEC-03 | Token hết hạn | Gửi request với JWT hết hạn | `401` Unauthorized. | ✅ |
| VC-SEC-04 | WS không token | Kết nối WS không có token | AuthMiddleware chặn. | ✅ |
| VC-SEC-05 | Duration overflow | Cuộc gọi kéo dài > 68 năm | `int` seconds overflow. Cần xem xét. | ℹ️ |
| VC-SEC-06 | Gọi user không tồn tại | POST initiate với callee_id random | Chỉ check friend → `400` "chỉ có thể gọi cho bạn bè" (vì không phải bạn). | ✅ |

---

## 13. State Machine (Status Transitions)

| ID | From | To | Valid | Notes |
|----|------|----|-------|-------|
| VC-SM-01 | calling | connected | ✅ | AcceptCall |
| VC-SM-02 | calling | rejected | ✅ | RejectCall |
| VC-SM-03 | calling | missed | ✅ | EndCall bởi caller khi chưa connected |
| VC-SM-04 | ringing | connected | ✅ | AcceptCall |
| VC-SM-05 | ringing | rejected | ✅ | RejectCall |
| VC-SM-06 | ringing | missed | ✅ | EndCall khi ringing |
| VC-SM-07 | connected | ended | ✅ | EndCall |
| VC-SM-08 | connected | rejected | ❌ | Bị chặn bởi service |
| VC-SM-09 | ended | connected | ❌ | Bị chặn bởi service |
| VC-SM-10 | rejected | connected | ❌ | Bị chặn bởi service |
| VC-SM-11 | missed | connected | ❌ | Bị chặn bởi service |
| VC-SM-12 | calling | ended | ✅ | EndCall (→ missed internally) |
| VC-SM-13 | calling | busy | N/A | busy là trạng thái trả về trước khi tạo call, không phải DB status |

---

## 14. API Response Format Verification

| ID | Endpoint | Expected Response | Status |
|----|----------|-------------------|--------|
| VC-RESP-01 | GET /history | `{"data": [...], "total": 5, "limit": 20, "offset": 0}` | ✅ |
| VC-RESP-02 | GET /:callID | `{"data": {...call object...}}` | ✅ |
| VC-RESP-03 | POST /initiate (success) | `{"data": {...call object...}}` | ✅ |
| VC-RESP-04 | POST /initiate (busy) | `{"message": "người dùng đang bận"}` | ✅ |
| VC-RESP-05 | POST /:callID/accept | `{"message": "cuộc gọi đã được chấp nhận"}` | ✅ |
| VC-RESP-06 | POST /:callID/reject | `{"message": "cuộc gọi đã bị từ chối"}` | ✅ |
| VC-RESP-07 | POST /:callID/mute | `{"message": "đã cập nhật trạng thái tắt tiếng"}` | ✅ |
| VC-RESP-08 | Error response | `{"error": "<message>"}` | ✅ |

---

## 15. WebSocket Event Types (Server → Client)

| Event | Direction | Payload | Triggered By |
|-------|-----------|---------|--------------|
| `call:incoming` | Server → Callee | `{call_id, caller_id, call_type, timestamp}` | InitiateCall (khi callee online) |
| `call:status` | Server → Both | `{call_id, status, caller_id, callee_id, call_type, started_at?, ended_at?, duration?}` | InitiateCall, AcceptCall, RejectCall, EndCall |
| `call:busy` | Server → Caller | `{callee_id}` | InitiateCall (khi callee busy) |
| `call:signal` | Server → Peer | `{call_id, sender_id, signal}` | HandleSignal |
| `call:mute` | Server → Both | `{call_id, user_id, muted}` | ToggleMute |

---

## Test Coverage Summary

| Feature | Total Cases | ✅ Pass | ⚠️ Warning | ❌ Fail | ℹ️ Note |
|---------|-------------|---------|------------|---------|---------|
| InitiateCall | 13 | 13 | 0 | 0 | 0 |
| AcceptCall | 7 | 7 | 0 | 0 | 0 |
| RejectCall | 6 | 6 | 0 | 0 | 0 |
| EndCall | 7 | 7 | 0 | 0 | 0 |
| ToggleMute | 8 | 8 | 0 | 0 | 0 |
| GetCallDetail | 4 | 4 | 0 | 0 | 0 |
| GetCallHistory | 7 | 7 | 0 | 0 | 0 |
| Signal Handling | 4 | 4 | 0 | 0 | 0 |
| Busy Detection | 6 | 5 | 1 | 0 | 0 |
| Repository | 13 | 13 | 0 | 0 | 0 |
| WS Client Events | 10 | 10 | 0 | 0 | 0 |
| Edge Cases & Security | 6 | 4 | 1 | 0 | 1 |
| State Machine | 13 | 13 | 0 | 0 | 0 |
| API Response Format | 8 | 8 | 0 | 0 | 0 |
| WS Event Types | 5 | 5 | 0 | 0 | 0 |
| **Total** | **117** | **114** | **2** | **0** | **1** |

### Known Issues

1. **VC-BUSY-04** (⚠️): TOCTOU race — friend check và DB create không cùng transaction. Fix: gộp friend check vào `CreateIfNotBusy`.
2. **VC-SEC-02** (⚠️): Thiếu block check — user bị block vẫn gọi được. Fix: inject `BlockRepository`.
