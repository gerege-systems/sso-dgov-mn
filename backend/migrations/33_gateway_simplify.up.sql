-- Government Template Platform V3.0
-- API Gateway-г хялбарчилна: runtime proxy байхгүй тул функциональ бус
-- gateway_routes + gateway_policies-ийг хасна. gateway_request_logs-ыг route/
-- consumer холбоосоос салгаж, mock seed-ийг цэвэрлэнэ — цаашид middleware
-- жинхэнэ /api хүсэлтүүдийг (method/path/status/latency/ip) бичнэ.

-- request_logs-ийн route/consumer хамаарлыг салгана.
ALTER TABLE gateway_request_logs DROP CONSTRAINT IF EXISTS gateway_request_logs_route_id_fkey;
ALTER TABLE gateway_request_logs DROP CONSTRAINT IF EXISTS gateway_request_logs_consumer_id_fkey;
ALTER TABLE gateway_request_logs DROP COLUMN IF EXISTS route_id;
ALTER TABLE gateway_request_logs DROP COLUMN IF EXISTS consumer_id;

-- Mock seed хүсэлтүүдийг арилгана (цаашид зөвхөн бодит хүсэлт бичигдэнэ).
TRUNCATE TABLE gateway_request_logs;

-- Функциональ бус routes + policies-ийг хасна (policies → routes CASCADE).
DROP TABLE IF EXISTS gateway_policies;
DROP TABLE IF EXISTS gateway_routes;
