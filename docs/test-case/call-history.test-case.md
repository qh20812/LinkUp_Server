# Call History — Test Cases

## REST API Endpoints

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| GET | `/api/calls/history` | GetCallHistory | Token | Lịch sử cuộc gọi (có filter, sort, pagination) |
| DELETE | `/api/calls/:callID/hide` | HideCall | Token | Ẩn cuộc gọi khỏi lịch sử (user-level soft delete) |
| GET | `/api/calls/missed/count` | GetMissedCallCount | Token | Đếm cuộc gọi nhỡ chưa đọc |
| POST | `/api/calls/missed/read` | MarkMissedRead | Token | Đánh dấu đã đọc cuộc gọi nhỡ |

---

## 1. GetCallHistory

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-HIST-01 | Lấy lịch sử mặc định | GET `/api/calls/history` | `200` response `{data: [...], total, limit: 20, offset: 0}`. Sort mặc định `created_at desc`. | ✅ |
| CH-HIST-02 | Phân trang | GET `/api/calls/history?limit=5&offset=10` | Trả đúng 5 records, offset từ vị trí 10. | ✅ |
| CH-HIST-03 | Limit > 100 | GET `/api/calls/history?limit=200` | Clamp về `limit=100`. | ✅ |
| CH-HIST-04 | Limit = 0 | GET `/api/calls/history?limit=0` | Clamp về `limit=100` (default capped). | ✅ |
| CH-HIST-05 | Limit < 0 | GET `/api/calls/history?limit=-5` | Clamp về `limit=100`. | ✅ |
| CH-HIST-06 | Offset âm | GET `/api/calls/history?offset=-1` | Clamp về `offset=0`. | ✅ |
| CH-HIST-07 | Filter theo call_type = voice | GET `/api/calls/history?type=voice` | Chỉ trả về cuộc gọi voice. | ✅ |
| CH-HIST-08 | Filter theo call_type = video | GET `/api/calls/history?type=video` | Chỉ trả về cuộc gọi video. | ✅ |
| CH-HIST-09 | Filter type không hợp lệ | GET `/api/calls/history?type=fax` | Type bị bỏ qua (treated as nil), trả tất cả. | ✅ |
| CH-HIST-10 | Filter theo status = missed | GET `/api/calls/history?status=missed` | Chỉ trả về cuộc gọi nhỡ. | ✅ |
| CH-HIST-11 | Filter theo status = ended | GET `/api/calls/history?status=ended` | Chỉ trả về cuộc gọi đã kết thúc. | ✅ |
| CH-HIST-12 | Filter status không hợp lệ | GET `/api/calls/history?status=deleted` | Status bị bỏ qua, trả tất cả. | ✅ |
| CH-HIST-13 | Sort theo duration | GET `/api/calls/history?sort=duration` | Kết quả sort theo duration tăng dần. | ✅ |
| CH-HIST-14 | Sort theo call_type | GET `/api/calls/history?sort=call_type` | Kết quả sort theo call_type. | ✅ |
| CH-HIST-15 | Sort column không hợp lệ | GET `/api/calls/history?sort=DROP TABLE` | Fallback về `sort=created_at`. | ✅ |
| CH-HIST-16 | Order asc | GET `/api/calls/history?order=asc` | Kết quả sort tăng dần. | ✅ |
| CH-HIST-17 | Order desc | GET `/api/calls/history?order=desc` | Kết quả sort giảm dần. | ✅ |
| CH-HIST-18 | Order không hợp lệ | GET `/api/calls/history?order=ascending` | Fallback về `order=desc`. | ✅ |
| CH-HIST-19 | User không có cuộc gọi nào | User mới, chưa có call history | `data: []`, `total: 0`. | ✅ |
| CH-HIST-20 | User có cả caller và callee | User A gọi B, B gọi A | Cả 2 cuộc gọi đều xuất hiện trong history của A. | ✅ |
| CH-HIST-21 | Cuộc gọi ẩn không xuất hiện | 1. A gọi B<br>2. A ẩn cuộc gọi `DELETE /:callID/hide`<br>3. A lấy history | Cuộc gọi đã ẩn KHÔNG xuất hiện trong kết quả. | ✅ |
| CH-HIST-22 | Mỗi user ẩn独立 | 1. A gọi B<br>2. A ẩn cuộc gọi<br>3. B lấy history của B | Cuộc gọi vẫn xuất hiện trong history của B (user-level soft delete). | ✅ |
| CH-HIST-23 | Direction logic — user là caller | A发起 call với B | direction = `outgoing` trong response của A. | ✅ |
| CH-HIST-24 | Direction logic — user là callee | B nhận call từ A | direction = `incoming` trong response của B. | ✅ |
| CH-HIST-25 | OtherUser profile batch load | A có 10 cuộc gọi với nhiều user khác nhau | Mỗi item có `other_user` với `id`, `display_name`, `avatar_url` đúng. Batch load 2-query (DISTINCT other_user_ids → FindByIDs). | ✅ |
| CH-HIST-26 | Không có token | GET `/api/calls/history` without Authorization | `401` Unauthorized. | ✅ |
| CH-HIST-27 | Kết hợp filter + sort + pagination | GET `/api/calls/history?type=voice&status=ended&sort=duration&order=asc&limit=3&offset=0` | Kết quả đúng filter, đúng sort, đúng phân trang. | ✅ |

---

## 2. HideCall

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-HIDE-01 | Ẩn cuộc gọi thành công | DELETE `/api/calls/:callID/hide` (user là participant) | `200` `{message: "đã ẩn cuộc gọi"}`. Tạo bản ghi `call_hidden`. | ✅ |
| CH-HIDE-02 | Ẩn cuộc gọi đã ẩn trước đó | DELETE `/api/calls/:callID/hide` (đã có trong call_hidden) | `200` (idempotent — không lỗi, không duplicate record). | ✅ |
| CH-HIDE-03 | User không phải participant | DELETE `/api/calls/:callID/hide` bởi user C (không phải caller/callee) | `400` error "không phải người tham gia cuộc gọi". | ✅ |
| CH-HIDE-04 | Cuộc gọi không tồn tại | DELETE `/api/calls/:callID/hide` với callID không hợp lệ | `400` error "cuộc gọi không tồn tại". | ✅ |
| CH-HIDE-05 | Thiếu callID param | DELETE `/api/calls//hide` | `404` route không match (parameter binding fail). | ✅ |
| CH-HIDE-06 | Không có token | DELETE without Authorization | `401` Unauthorized. | ✅ |

---

## 3. GetMissedCallCount

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-MISS-01 | Đếm cuộc gọi nhỡ mặc định | GET `/api/calls/missed/count` | `200` `{count: N}`. Đếm từ `last_read_missed_at` hoặc `created_at`. | ✅ |
| CH-MISS-02 | Không có cuộc gọi nhỡ | User không có call status=missed | `count: 0`. | ✅ |
| CH-MISS-03 | Có 3 cuộc gọi nhỡ | 3 calls với status=missed, created_at > last_read_missed_at | `count: 3`. | ✅ |
| CH-MISS-04 | Đếm bằng >= last_read_missed_at | `last_read_missed_at = 2026-07-01`<br>Calls: `2026-07-01` (missed), `2026-07-02` (missed) | `count: 2` (>=, không phải >). | ✅ |
| CH-MISS-05 | Cuộc gọi ended không tính | Calls: 1 missed, 1 ended | `count: 1`. | ✅ |
| CH-MISS-06 | Cuộc gọi rejected không tính | Calls: 1 missed, 1 rejected | `count: 1`. | ✅ |
| CH-MISS-07 | Ẩn cuộc gọi nhỡ vẫn tính | 1 missed call + user đã hide call đó | `count: 1` (hide không ảnh hưởng missed count). | ✅ |
| CH-MISS-08 | Sau MarkMissedRead, count = 0 | 1. Có 2 missed calls<br>2. POST `/calls/missed/read`<br>3. GET `/calls/missed/count` | `count: 0`. | ✅ |
| CH-MISS-09 | Không có token | GET without Authorization | `401` Unauthorized. | ✅ |

---

## 4. MarkMissedRead

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-READ-01 | Đánh dấu đã đọc thành công | POST `/api/calls/missed/read` | `200` `{message: "đã đánh dấu đã đọc"}`. `profiles.last_read_missed_at` update về `NOW()`. | ✅ |
| CH-READ-02 | Sau đọc, missed count = 0 | 1. POST `/calls/missed/read`<br>2. GET `/calls/missed/count` | `count: 0`. | ✅ |
| CH-READ-03 | Missed call mới sau read | 1. POST `/calls/missed/read`<br>2. Cuộc gọi nhỡ mới arrive<br>3. GET `/calls/missed/count` | `count: 1` (cuộc gọi mới sau thời điểm read). | ✅ |
| CH-READ-04 | Gọi nhiều lần | 1. POST `/calls/missed/read`<br>2. POST `/calls/missed/read` lần nữa | Không lỗi. `last_read_missed_at` cập nhật lại (idempotent). | ✅ |
| CH-READ-05 | Không có token | POST without Authorization | `401` Unauthorized. | ✅ |

---

## 5. DTO Serialization

### 5.1 CallHistoryItem

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-DTO-01 | JSON roundtrip đầy đủ | Marshal → Unmarshal `CallHistoryItem` có đầy đủ fields | Tất cả fields giữ nguyên giá trị. | ✅ |
| CH-DTO-02 | JSON roundtrip missed call | Marshal `CallHistoryItem` với `status: "missed"`, `is_missed: true` | JSON chứa `"is_missed":true`, `"status":"missed"`, `"direction":"incoming"`. | ✅ |
| CH-DTO-03 | Omit nil timestamps | Marshal `CallHistoryItem` với `started_at: nil`, `ended_at: nil` | JSON KHÔNG chứa `"started_at"` hay `"ended_at"` (omitempty). | ✅ |
| CH-DTO-04 | Include non-nil timestamps | Marshal với `started_at` và `ended_at` có giá trị | JSON chứa cả 2 field. | ✅ |

### 5.2 UserBrief

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-DTO-05 | JSON roundtrip | Marshal → Unmarshal `UserBrief` với đầy đủ fields | `id`, `display_name`, `avatar_url` giữ nguyên. | ✅ |
| CH-DTO-06 | Empty fields | Marshal `UserBrief{}` | JSON chứa 3 field rỗng/null. | ✅ |

### 5.3 CallMissedPayload

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-DTO-07 | JSON roundtrip | Marshal → Unmarshal `CallMissedPayload` | `call_id`, `caller_id`, `timestamp` giữ nguyên. | ✅ |
| CH-DTO-08 | Field count = 3 | Marshal và kiểm tra raw JSON | Đúng 3 field: `call_id`, `caller_id`, `timestamp`. | ✅ |

### 5.4 CallHistoryQuery

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-DTO-09 | Zero value defaults | Tạo `CallHistoryQuery{}` | `limit=0`, `offset=0`, `type=nil`, `status=nil`. | ✅ |
| CH-DTO-10 | With values | Tạo với limit=10, offset=5, type=video, status=missed | Tất cả fields đúng giá trị. | ✅ |

---

## 6. Query Whitelist Logic (queryToFilter)

| ID | Scenario | Input | Expected Output | Status |
|----|----------|-------|-----------------|--------|
| CH-WL-01 | Defaults | sort="", order="", type=nil, status=nil | sort=`created_at`, order=`desc`, type=nil, status=nil | ✅ |
| CH-WL-02 | Type voice | type=`"voice"` | type=`"voice"` | ✅ |
| CH-WL-03 | Type video | type=`"video"` | type=`"video"` | ✅ |
| CH-WL-04 | Type invalid | type=`"fax"` | type=nil (rejected) | ✅ |
| CH-WL-05 | Status missed | status=`"missed"` | status=`"missed"` | ✅ |
| CH-WL-06 | Status ended | status=`"ended"` | status=`"ended"` | ✅ |
| CH-WL-07 | Status rejected | status=`"rejected"` | status=`"rejected"` | ✅ |
| CH-WL-08 | Status invalid | status=`"deleted"` | status=nil (rejected) | ✅ |
| CH-WL-09 | Sort: created_at | sort=`"created_at"` | sort=`"created_at"` | ✅ |
| CH-WL-10 | Sort: duration | sort=`"duration"` | sort=`"duration"` | ✅ |
| CH-WL-11 | Sort: call_type | sort=`"call_type"` | sort=`"call_type"` | ✅ |
| CH-WL-12 | Sort: status | sort=`"status"` | sort=`"status"` | ✅ |
| CH-WL-13 | Sort: uppercase | sort=`"CREATED_AT"` | sort=`"created_at"` (lowercased) | ✅ |
| CH-WL-14 | Sort: invalid column | sort=`"invalid_column"` | sort=`"created_at"` (fallback) | ✅ |
| CH-WL-15 | Sort: empty string | sort=`""` | sort=`"created_at"` (fallback) | ✅ |
| CH-WL-16 | Sort: id | sort=`"id"` | sort=`"created_at"` (not in whitelist) | ✅ |
| CH-WL-17 | Order: asc | order=`"asc"` | order=`"asc"` | ✅ |
| CH-WL-18 | Order: desc | order=`"desc"` | order=`"desc"` | ✅ |
| CH-WL-19 | Order: uppercase | order=`"ASC"` | order=`"asc"` (lowercased) | ✅ |
| CH-WL-20 | Order: invalid | order=`"invalid"` | order=`"desc"` (fallback) | ✅ |
| CH-WL-21 | Order: empty | order=`""` | order=`"desc"` (fallback) | ✅ |

---

## 7. SQL Injection Prevention

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-INJ-01 | Sort: SQL injection semicolon | sort=`"created_at; DROP TABLE calls"` | sort=`"created_at"` (rejected, fallback) | ✅ |
| CH-INJ-02 | Sort: SQL injection OR 1=1 | sort=`"1=1 OR 1"` | sort=`"created_at"` (rejected, fallback) | ✅ |
| CH-INJ-03 | Order: SQL injection semicolon | order=`"desc; SELECT * FROM users"` | order=`"desc"` (rejected, fallback) | ✅ |
| CH-INJ-04 | Sort: Unicode bypass | sort=`"créated_at"` | sort=`"created_at"` (rejected, fallback) | ✅ |
| CH-INJ-05 | Sort: Null byte | sort=`"created_at\x00"` | sort=`"created_at"` (rejected, fallback) | ✅ |

---

## 8. Direction & OtherUser Logic

| ID | Scenario | userID | callerID | calleeID | Expected | Status |
|----|----------|--------|----------|----------|----------|--------|
| CH-DIR-01 | User is caller | user-a | user-a | user-b | direction=`outgoing`, other_user=user-b | ✅ |
| CH-DIR-02 | User is callee | user-b | user-a | user-b | direction=`incoming`, other_user=user-a | ✅ |
| CH-DIR-03 | User is caller (different) | user-x | user-x | user-y | direction=`outgoing`, other_user=user-y | ✅ |
| CH-DIR-04 | User is callee (different) | user-y | user-x | user-y | direction=`incoming`, other_user=user-x | ✅ |

---

## 9. IsMissed Logic

| ID | Status | Expected IsMissed | Status |
|----|--------|-------------------|--------|
| CH-MSD-01 | `missed` | `true` | ✅ |
| CH-MSD-02 | `ended` | `false` | ✅ |
| CH-MSD-03 | `rejected` | `false` | ✅ |
| CH-MSD-04 | `connected` | `false` | ✅ |
| CH-MSD-05 | `calling` | `false` | ✅ |

---

## 10. CallHidden Model

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-MOD-01 | Primary keys | Tạo `CallHidden{CallID, UserID}` | Cả 2 fields đúng giá trị (composite PK). | ✅ |
| CH-MOD-02 | JSON serialization | Marshal → Unmarshal | `call_id`, `user_id`, `created_at` giữ nguyên. | ✅ |

---

## 11. Service Layer — EndCall sends call:missed

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| CH-SVC-01 | EndCall khi chưa connected → missed WS | 1. A gọi B (calling)<br>2. A end call<br>3. B đang online | Call status = `missed`. B nhận `call:missed` WS event với `{call_id, caller_id, timestamp}`. | ✅ |
| CH-SVC-02 | EndCall khi connected → ended (không missed WS) | 1. A gọi B → connected<br>2. A end call | Call status = `ended`. KHÔNG có `call:missed` WS event. | ✅ |

---

## 12. WebSocket Event Types (Server → Client)

| Event | Direction | Payload | Triggered By |
|-------|-----------|---------|--------------|
| `call:missed` | Server → Callee | `{call_id, caller_id, timestamp}` | EndCall (khi status chuyển sang missed) |

---

## 13. API Response Format Verification

| ID | Endpoint | Expected Response | Status |
|----|----------|-------------------|--------|
| CH-FMT-01 | GET `/calls/history` | `{"data": [{id, other_user, call_type, direction, status, is_missed, duration, started_at?, ended_at?, created_at}], "total": N, "limit": N, "offset": N}` | ✅ |
| CH-FMT-02 | DELETE `/calls/:callID/hide` | `{"message": "đã ẩn cuộc gọi"}` | ✅ |
| CH-FMT-03 | GET `/calls/missed/count` | `{"count": N}` | ✅ |
| CH-FMT-04 | POST `/calls/missed/read` | `{"message": "đã đánh dấu đã đọc"}` | ✅ |
| CH-FMT-05 | Error response | `{"error": "<message>"}` | ✅ |

---

## 14. Response Envelope Structure

| ID | Field | Type | Description | Status |
|----|-------|------|-------------|--------|
| CH-ENV-01 | `data` | `[]CallHistoryItem` | Danh sách cuộc gọi | ✅ |
| CH-ENV-02 | `total` | `int` | Tổng số records (trước pagination) | ✅ |
| CH-ENV-03 | `limit` | `int` | Số lượng records mỗi trang | ✅ |
| CH-ENV-04 | `offset` | `int` | Vị trí bắt đầu | ✅ |

---

## Test Coverage Summary

| Feature | Total Cases | ✅ Pass | Status |
|---------|-------------|---------|--------|
| GetCallHistory | 27 | 27 | ✅ |
| HideCall | 6 | 6 | ✅ |
| GetMissedCallCount | 9 | 9 | ✅ |
| MarkMissedRead | 5 | 5 | ✅ |
| DTO Serialization | 10 | 10 | ✅ |
| Query Whitelist | 21 | 21 | ✅ |
| SQL Injection | 5 | 5 | ✅ |
| Direction Logic | 4 | 4 | ✅ |
| IsMissed Logic | 5 | 5 | ✅ |
| CallHidden Model | 2 | 2 | ✅ |
| Service (call:missed) | 2 | 2 | ✅ |
| WS Event Types | 1 | 1 | ✅ |
| API Response Format | 5 | 5 | ✅ |
| Envelope Structure | 4 | 4 | ✅ |
| **Total** | **110** | **110** | ✅ |

---

## Test File

`tests/call/call_history_test.go` — 29 unit tests (validation-only, no DB)

```
TestCallHistoryItemJSON
TestCallHistoryItemMissedCall
TestCallHistoryItemOmitsNilTimestamps
TestUserBriefJSON
TestUserBriefEmptyFields
TestCallMissedPayloadJSON
TestCallMissedPayloadFieldCount
TestDirectionLogic (4 subtests)
TestOtherUserIDLogic (2 subtests)
TestIsMissedLogic (5 subtests)
TestCallHiddenPrimaryKeys
TestCallHiddenJSON
TestCallHistoryQueryDefaults
TestCallHistoryQueryWithValues
TestQueryToFilterDefaults
TestQueryToFilterWhitelistType
TestQueryToFilterWhitelistStatus
TestQueryToFilterWhitelistSort (10 subtests)
TestQueryToFilterWhitelistOrder (7 subtests)
TestQueryToFilterSQLInjection (5 subtests)
TestCallHistoryResponseEnvelope
```
