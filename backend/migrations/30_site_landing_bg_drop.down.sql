-- Government Template Platform V3.0
-- 30_site_landing_bg_drop-ийн буцаалт — landing_bg баганыг сэргээнэ (29-тэй ижил).
ALTER TABLE site_appearance
    ADD COLUMN IF NOT EXISTS landing_bg TEXT NOT NULL DEFAULT '#0f1f39';
