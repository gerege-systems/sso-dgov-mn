#!/usr/bin/env bash
# Шинэ байрлуулалтад эхний байгууллага, эхний админыг өгнө.
#
#   ./first-admin.sh <и-мэйл> <нууц үг> ["Байгууллагын нэр"] [slug]
#
# Яагаад SQL вэ: бүртгүүлэх дэлгэц гэж байхгүй (`/api/v1/auth/register`
# байхгүй), баримтжсан зам болох control plane нь өөрийн vhost,
# CONTROL_PLANE_HOST, TOTP бүртгэл шаарддаг — эхний хүнийг оруулахын тулд
# өөр бүтээгдэхүүн босгох хэрэг гардаг. open.dgov.mn ч мөн ингэж эхэлсэн.
#
# Дахин ажиллуулж болно: гурван INSERT бүгд ON CONFLICT DO NOTHING.
set -euo pipefail

email="${1:?и-мэйл}"; password="${2:?нууц үг}"
tenant_name="${3:-Цахим Засаг}"; tenant_slug="${4:-dgov}"

APP_DIR="$(cd "$(dirname "$0")" && pwd)"

docker exec -i dgov_sso_postgres psql -v ON_ERROR_STOP=1 -U postgres -d platform_db \
  -v email="$email" -v pass="$password" -v tname="$tenant_name" -v tslug="$tenant_slug" <<'SQL'
-- Миграцууд үүнийг суулгадаггүй. Нууц үгийн hash-ийг сангийн дотор үүсгэж,
-- ил бичвэрийг psql-ийн түүх, процессын жагсаалтад гаргахгүй байхын тулд.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO tenants (slug, name) VALUES (:'tslug', :'tname')
  ON CONFLICT (slug) DO NOTHING;

-- gen_salt('bf', 10) нь Go-гийн bcrypt уншдаг `$2a$10$` хэлбэрийг өгнө.
INSERT INTO users (email, password_hash, name)
  VALUES (:'email', crypt(:'pass', gen_salt('bf', 10)), :'email')
  ON CONFLICT (email) DO NOTHING;

INSERT INTO memberships (tenant_id, user_id)
  SELECT t.id, u.id FROM tenants t, users u
   WHERE t.slug = :'tslug' AND u.email = :'email'
  ON CONFLICT (tenant_id, user_id) DO NOTHING;

-- Тусдаа өгүүлбэр байх нь заавал: `admin` рольд tenants дээрх AFTER INSERT
-- trigger үүсгэдэг тул тенантыг үүсгэсэн өгүүлбэр дотроос харагдахгүй — нэг
-- CTE дотор бичвэл юу ч оруулалгүй чимээгүй амжилттай болно.
INSERT INTO membership_roles (membership_id, role_id)
  SELECT m.id, r.id
    FROM memberships m
    JOIN tenants t ON t.id = m.tenant_id AND t.slug = :'tslug'
    JOIN users   u ON u.id = m.user_id   AND u.email = :'email'
    JOIN roles   r ON r.tenant_id = t.id AND r.code = 'admin'
  ON CONFLICT DO NOTHING;

DROP EXTENSION pgcrypto;
SQL

# Шалгалт: бичсэн зүйл нь нэвтэрч чадах эсэх. Cookie-гоор баталгаажсан
# хүсэлт Origin эсвэл Sec-Fetch-Site авчрах ёстой тул header нь сайн дурын
# биш — байхгүй бол 403 нь эрхийн алдаа мэт харагдана, тийм биш.
origin="$(grep -E '^PUBLIC_ORIGIN=' "$APP_DIR/.env" | cut -d= -f2-)"
code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "http://127.0.0.1:8096/api/v1/auth/login" \
  -H 'Content-Type: application/json' -H "Origin: ${origin}" \
  -d "{\"email\":\"${email}\",\"password\":\"${password}\"}")
[ "$code" = "200" ] || { echo "нэвтрэлт $code буцаав — админ бэлэн биш" >&2; exit 1; }
echo "OK: ${email} → ${tenant_name} (${tenant_slug})"
