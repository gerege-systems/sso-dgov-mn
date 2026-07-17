-- Government Template Platform V3.0
-- Landing (нүүр) хуудасны navy дэвсгэр өнгийг админ 'settings.manage' эрхээр
-- тохируулах боломж. site_appearance singleton мөрөнд '#rrggbb' hex багана нэмнэ.
-- Хоосон/буруу үед frontend нь built-in default (#0f1f39)-ыг хэрэглэнэ. Утгын
-- баталгаажуулалт (hex) usecase/handler давхаргад — 27-р шилжилтийн accent-тэй
-- ижил зарчим. RLS хэрэггүй (нийтийн config); app зөвхөн UPDATE хийдэг.
ALTER TABLE site_appearance
    ADD COLUMN IF NOT EXISTS landing_bg TEXT NOT NULL DEFAULT '#0f1f39';
