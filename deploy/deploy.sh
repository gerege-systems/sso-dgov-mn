#!/usr/bin/env bash
# sso.dgov.mn-ийг шинэчлэх. Серверийн /opt/sso-dgov-mn/deploy дотроос
# ажиллана.
#
#   ./deploy.sh
#
# Хажуугийн open.dgov.mn-ээс нэг ялгаа бий: тэр репод код байхгүй тул
# нийтлэгдсэн образ татдаг. Энд бүтээгдэхүүний бинар өөрөө баригдана —
# цөм нь go.mod-оор түгжигдэж ирдэг ч дөрвөн модуль эндхийнх. Тиймээс
# «шинэчлэх» гэдэг нь татах биш, барих.
#
# Бүрхүүл нь эсрэгээрээ нийтлэгдсэн образ (WEB_IMAGE) — distribution нь
# frontend-ээ fork хийдэггүй. GHCR-ийн пакет хаалттай тул тэр татахад
# токен хэрэгтэй бөгөөд серверт хадгалагддаггүй:
#
#   REGISTRY_USER=<github-хэрэглэгч> REGISTRY_TOKEN=<read:packages> ./deploy.sh
#
# Образ хост дээр аль хэдийн байвал энэ хоёр хувьсагч шаардлагагүй.
set -euo pipefail

APP_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$APP_DIR"

if [ ! -f .env ]; then
  echo "$APP_DIR/.env алга. .env.example-ээс хуулж, нууцуудыг нь бөглө." >&2
  exit 1
fi

if [ -n "${REGISTRY_TOKEN:-}" ]; then
  : "${REGISTRY_USER:?REGISTRY_TOKEN өгсөн бол REGISTRY_USER ч хэрэгтэй}"
  echo "$REGISTRY_TOKEN" | docker login ghcr.io -u "$REGISTRY_USER" --password-stdin
  trap 'docker logout ghcr.io >/dev/null 2>&1 || true' EXIT
fi

# Барилт эхлээд, тусад нь: ажиллаж байгаа контейнеруудыг шинэ образ бүрэн
# бэлэн болсны дараа л хөндөнө. Барилт унасны дараах татан буулгалт бол
# амжилтгүй deploy биш, тасалдал.
docker compose build backend
docker compose pull web

docker compose up -d --remove-orphans

# Шалгалт: API эрүүл, discovery нь өөрийгөө зөв нэрлэж байна, бүрхүүл
# ирж байна. Хоёр дахь нь сайн дурын биш — issuer буруу бол клиент бүр
# токеныг татгалзана, харин стек нь бүрэн эрүүл харагдана.
for i in $(seq 1 45); do
  curl -fsS http://127.0.0.1:8096/health >/dev/null 2>&1 && break
  [ "$i" -eq 45 ] && { echo "backend 90 секундэд эрүүл болсонгүй" >&2; docker compose logs --tail 40 backend >&2; exit 1; }
  sleep 2
done

origin="$(grep -E '^PUBLIC_ORIGIN=' .env | cut -d= -f2-)"
issuer="$(curl -fsS http://127.0.0.1:8096/.well-known/openid-configuration | sed -n 's/.*"issuer":"\([^"]*\)".*/\1/p')"
[ "$issuer" = "$origin" ] || { echo "issuer «$issuer» нь PUBLIC_ORIGIN «$origin»-тэй таарахгүй" >&2; exit 1; }

# Бүрхүүл backend эрүүл болсны дараа хэдэн секунд асдаг тул нэг удаагийн
# шалгалт эрүүл rollout-ыг унасан гэж дуудна.
brand="$(grep -E '^BRAND_NAME=' .env | cut -d= -f2-)"
for i in $(seq 1 45); do
  if body="$(curl -fsS http://127.0.0.1:3016/login 2>/dev/null)"; then
    [ -z "$brand" ] && break
    case "$body" in *"$brand"*) break ;; esac
  fi
  [ "$i" -eq 45 ] && { echo "бүрхүүл 90 секундэд «${brand:-хариу}» өгсөнгүй" >&2; exit 1; }
  sleep 2
done

echo "OK: $origin — $brand"
