-- Government Template Platform V3.0
-- Landing navy дэвсгэрийн тохиргоог site_appearance-аас Themes систем рүү
-- (themes.config.appearance.colors.lpNavy) нүүлгэсэн тул 29-р шилжилтийн нэмсэн
-- landing_bg баганыг хасна. Landing navy-г одоо /admin/themes-ийн палитраас
-- (идэвхтэй theme) удирдана.
ALTER TABLE site_appearance
    DROP COLUMN IF EXISTS landing_bg;
