// Брэндийн нэрийг кодод шууд бичихээс сэргийлнэ.
//
// Платформын нэр ЗӨВХӨН `src/brand.config.ts` дотор байх ёстой (болон
// `src/components/landing/` — тэр нь зориудаар брэндийн маркетингийн текст).
//
// Яагаад Node дээр бичсэн бэ: энэ шалгалт `npm run build`-ийн нэг хэсэг тул
// Docker-ийн `node:20-alpine` дотор ч ажиллах ёстой. Тэнд bash байхгүй, busybox
// grep нь `-I` / `--exclude`-ийг дэмждэггүй. Node бол цорын ганц найдвартай орчин.
//
// Дэлгэрэнгүй: vision-gerege-mn/UI_CORE_PLAN.md §3.4

import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('..', import.meta.url).pathname;
const SRC = join(ROOT, 'src');

// Флот дахь бүх платформын нэр — аль нь ч кодод шууд байж болохгүй.
const PATTERN =
  /Gerege Template Platform V3\.0|Gerege App|Gerege POS|Gerege Kiosk|Gerege Wallet|Ring System|Government Template Platform/;

// Зөвшөөрөгдсөн: брэндийн тохиргоо ба landing-ийн маркетингийн текст.
const ALLOW = [/^brand\.config\.ts$/, /^components\/landing\//];

const TEXT = /\.(ts|tsx|js|jsx|mjs|cjs|css|json|md)$/;

function walk(dir, out = []) {
  for (const e of readdirSync(dir)) {
    const p = join(dir, e);
    if (statSync(p).isDirectory()) walk(p, out);
    else if (TEXT.test(e)) out.push(p);
  }
  return out;
}

const hits = [];
for (const file of walk(SRC)) {
  const rel = relative(SRC, file);
  if (ALLOW.some((re) => re.test(rel))) continue;
  const lines = readFileSync(file, 'utf8').split('\n');
  lines.forEach((line, i) => {
    if (PATTERN.test(line)) hits.push(`src/${rel}:${i + 1}:${line.trim()}`);
  });
}

if (hits.length > 0) {
  console.error('✗ Брэндийн нэр кодод шууд бичигдсэн байна:\n');
  for (const h of hits) console.error('    ' + h);
  console.error('\n  Засах: src/brand.config.ts -ийн brand.name / brand.short -ыг ашиглана.');
  console.error("         Хуудасны гарчигт: pageTitle('<хуудасны нэр>')");
  process.exit(1);
}

console.log('✓ Брэнд зөвхөн brand.config.ts дотор.');
