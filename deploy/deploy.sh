#!/usr/bin/env bash
# Government Template Platform V3.0
# Gerege Systems Development Team & Claude AI, 2026
#
# Remote deploy step, run ON the server by the CD workflow (.github/workflows/deploy.yml)
# after the target commit is already checked out. Rebuilds images, restarts the
# compose stack, waits for health, and prunes dangling images. Idempotent — safe
# to re-run by hand: `bash deploy/deploy.sh`.
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "▶ Deploy commit: $(git rev-parse --short HEAD) — $(git log -1 --pretty=%s)"

# INTEGRATION_ENC_KEY — superadmin MFA-ийн TOTP secret болон integrations OAuth
# token-ийг AES-GCM-ээр шифрлэх түлхүүр. CD workflow нь GitHub secret-ээс энэ
# скриптэд дамжуулна. backend.env-д БАЙХГҮЙ тохиолдолд Л нэг удаа бичнэ
# (idempotent) — нэгэнт тавьсан түлхүүрийг дахин бичихгүй тул хэзээ ч
# өөрчлөгдөхгүй (өөрчилвөл өмнөх шифрлэсэн бүх өгөгдөл эвдэрнэ).
if [ -n "${INTEGRATION_ENC_KEY:-}" ] && ! grep -q '^INTEGRATION_ENC_KEY=' backend.env 2>/dev/null; then
  printf 'INTEGRATION_ENC_KEY=%s\n' "$INTEGRATION_ENC_KEY" >> backend.env
  echo "▶ INTEGRATION_ENC_KEY-г backend.env-д бичлээ (superadmin MFA идэвхжинэ)"
fi

# ── Юу ч өөрчлөгдөөгүй бол огт хөдлөхгүй ────────────────────────────────────
#
# Docker-ийн build нь ТОГТВОРТОЙ БИШ: ижил эх кодоос ч build бүрд ШИНЭ image ID
# гардаг (apk/wget/go build давхаргууд). Тиймээс `up -d` нь «image өөрчлөгдсөн»
# гэж үзээд контейнерийг ҮРГЭЛЖ дахин үүсгэдэг — код огт хөндөөгүй commit дээр
# ч секундын 502 (хэмжигдсэн 2026-07-27).
#
# Иймд хамгийн найдвартай хамгаалалт нь: HEAD өөрчлөгдөөгүй бол ЮУ Ч ХИЙХГҮЙ.
STAMP=".deployed-sha"
HEAD_SHA="$(git rev-parse HEAD)"
if [ "${FORCE_DEPLOY:-0}" != "1" ] && [ -f "$STAMP" ] && [ "$(cat "$STAMP")" = "$HEAD_SHA" ]; then
  echo "▶ Өөрчлөлт алга (${HEAD_SHA:0:12}) — build/up алгасав, тасалдал үүсгэхгүй."
  echo "  Албадах бол: FORCE_DEPLOY=1 bash deploy/deploy.sh"
  docker compose ps
  exit 0
fi

# Өгөгдлийн үйлчилгээг ЗОРИУД build хийхгүй. Тэдгээрийн image нь зөвхөн
# Dockerfile нь өөрчлөгдөхөд шинэчлэгдэх ёстой; deploy бүрд дахин build хийвэл
# шинэ image ID гарч db контейнер дахин үүсдэг — энэ нь бүх идэвхтэй DB
# холболтыг тасалдаг тул хамгийн үнэтэй тасалдал. Шаардлагатай үед гараар:
#   docker compose build db && docker compose up -d db
BUILD_SVCS="$(docker compose --profile migrate config --services \
              | grep -vxE 'db|redis|postgres' | tr '\n' ' ')"
echo "▶ Building images ($BUILD_SVCS)…"
# shellcheck disable=SC2086  # word splitting is intended
docker compose build $BUILD_SVCS

# Migration нь ТУСДАА алхам — `up -d`-ийн нэг хэсэг БИШ. Шалтгаан: өмнө нь
# `up -d` нь migrate-ыг дахин ажиллуулж, api+web-ийг ҮРГЭЛЖ дахин үүсгэдэг
# байсан — код огт хөндөөгүй commit дээр ч секундын 502 (хэмжигдсэн 07-27).
# Амжилтгүй бол deploy ЭНД зогсоно, api-д хүрэхгүй.
echo "▶ Running migrations…"
docker compose up -d --no-recreate db redis
# --no-deps: migrate-ийн depends_on-оор db-д гар хүрэхээс сэргийлнэ.
docker compose --profile migrate run --rm --no-deps migrate

echo "▶ Starting stack (migrate re-runs; applied migrations are skipped)…"
docker compose up -d --remove-orphans

# Wait until api + web report healthy (compose healthchecks). ~120s budget.
echo "▶ Waiting for containers to become healthy…"
deadline=$(( $(date +%s) + 150 ))
for svc in api web; do
  cid="$(docker compose ps -q "$svc")"
  if [ -z "$cid" ]; then echo "✖ service '$svc' has no container" >&2; exit 1; fi
  while true; do
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$cid" 2>/dev/null || echo unknown)"
    case "$status" in
      healthy|running) echo "  ✓ $svc: $status"; break ;;
      unhealthy|exited|dead) echo "✖ $svc became '$status'" >&2; docker logs --tail 40 "$cid" >&2; exit 1 ;;
    esac
    if [ "$(date +%s)" -ge "$deadline" ]; then
      echo "✖ timeout waiting for $svc (last: $status)" >&2; docker logs --tail 40 "$cid" >&2; exit 1
    fi
    sleep 3
  done
done

echo "▶ Pruning dangling images…"
docker image prune -f >/dev/null

echo "▶ Stack status:"
docker compose ps
echo "✅ Deploy complete."

# Амжилттай дууссан тул дараагийн no-op deploy алгасахад тамга үлдээнэ.
echo "$HEAD_SHA" > "$STAMP"
