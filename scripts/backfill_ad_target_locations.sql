-- Backfill ads.target_locations từ TÊN tỉnh (dữ liệu cũ) sang province ID,
-- khớp với profiles.current_province và danh mục 34 tỉnh/thành (post-2025).
--
-- KHÔNG chạy tự động. Quy trình:
--   1) Chạy phần DRY-RUN để xem ad nào còn lưu tên.
--   2) Chạy phần UPDATE trên môi trường staging trước.
--   3) Chạy trên production khi đã đối chiếu.
--
-- Ads mới (web hiện tại) đã gửi province ID nên không bị ảnh hưởng.

-- ─── DRY-RUN: liệt kê ad còn target_locations chứa tên ────────────────────
SELECT id, title, target_locations
FROM ads
WHERE JSON_SEARCH(target_locations, 'one', 'Hà Nội') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'TP.HCM') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'TP. Hồ Chí Minh') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'Hồ Chí Minh') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'Đà Nẵng') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'Hải Phòng') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'Cần Thơ') IS NOT NULL
   OR JSON_SEARCH(target_locations, 'one', 'Huế') IS NOT NULL;

-- ─── UPDATE: thay từng tên bằng ID (giữ nguyên các phần tử khác) ──────────
UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Hà Nội')), '01')
WHERE JSON_SEARCH(target_locations, 'one', 'Hà Nội') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Hồ Chí Minh')), '79')
WHERE JSON_SEARCH(target_locations, 'one', 'Hồ Chí Minh') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'TP. Hồ Chí Minh')), '79')
WHERE JSON_SEARCH(target_locations, 'one', 'TP. Hồ Chí Minh') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'TP.HCM')), '79')
WHERE JSON_SEARCH(target_locations, 'one', 'TP.HCM') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Đà Nẵng')), '48')
WHERE JSON_SEARCH(target_locations, 'one', 'Đà Nẵng') IS NOT NULL;

-- Quảng Nam (cũ) đã sáp nhập vào Đà Nẵng.
UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Quảng Nam')), '48')
WHERE JSON_SEARCH(target_locations, 'one', 'Quảng Nam') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Hải Phòng')), '31')
WHERE JSON_SEARCH(target_locations, 'one', 'Hải Phòng') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Cần Thơ')), '92')
WHERE JSON_SEARCH(target_locations, 'one', 'Cần Thơ') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Huế')), '46')
WHERE JSON_SEARCH(target_locations, 'one', 'Huế') IS NOT NULL;

-- Bình Dương / Bà Rịa - Vũng Tàu (cũ) đã sáp nhập vào Hồ Chí Minh.
UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Bình Dương')), '79')
WHERE JSON_SEARCH(target_locations, 'one', 'Bình Dương') IS NOT NULL;

UPDATE ads SET target_locations = JSON_REPLACE(target_locations,
  JSON_UNQUOTE(JSON_SEARCH(target_locations, 'one', 'Bà Rịa - Vũng Tàu')), '79')
WHERE JSON_SEARCH(target_locations, 'one', 'Bà Rịa - Vũng Tàu') IS NOT NULL;

-- Kiểm tra sau backfill: không còn phần tử nào không phải ID 2 chữ số.
-- SELECT id, target_locations FROM ads
-- WHERE JSON_SEARCH(target_locations, 'one', '%[^0-9]%') IS NOT NULL; -- tham khảo
