-- SSO нэвтрэлт бол СУУРЬ (built-in) service — бүртгэгдсэн апп бүрт автоматаар
-- үйлчилдэг (base OIDC scope: openid/profile/email). Тиймээс grant/checkbox-оор
-- олгодог "dan-sso" gateway service нь шаардлагагүй, төөрөгдүүлдэг тул хасна.
-- (eid-proxy/eid-org-proxy/eid-sign зэрэг нэмэлт service-үүд нь олголттой хэвээр.)
DELETE FROM gateway_services WHERE name = 'dan-sso';
