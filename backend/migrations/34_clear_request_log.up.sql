-- Government Template Platform V3.0
-- Хүсэлтийн лог-г зөвхөн RP-ийн gateway трафик бүртгэдэг болгосон (middleware).
-- Өмнө нь бичигдсэн DAN-ий first-party API мөрүүдийг (rbac/users/themes г.м.)
-- нэг удаа цэвэрлэнэ — цаашид зөвхөн /rp/sign, /api/v1/provider бүртгэгдэнэ.
TRUNCATE TABLE gateway_request_logs;
