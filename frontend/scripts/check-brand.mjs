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

// Флот дахь БУСАД платформын нэр — аль нь ч кодод шууд байж болохгүй.
const FLEET = [
  'Gerege Template Platform V3.0',
  'Government Template Platform V3.0',
  'Gerege Platform V3.0',
  'Gerege Developer Portal V3.0',
  'Government Developer Portal V3.0',
  'Gerege Wallet V1.0',
  'Gerege App',
  'Gerege POS',
  'Gerege Kiosk',
  'Gerege SSO',
  'Ring System',
  'Төрийн нэгдсэн нэвтрэлт',
  'Төрийн үйлчилгээний «Хурдан» платформ',
];

// ЭНЭ платформын нэрийг `brand.config.ts`-ээс УНШИНА — гараар давхардуулбал
// репо бүрт зөрч, өөрийн нэрээ хамгаалдаггүй болно (2026-07-30-нд 14 репогийн
// 9 нь тийм байв; иймээс «Ring System» гэсэн нэр өөр платформ дээр амьд
// үлдсэн). Уншиж чадахгүй бол ЗОГСОНО — чимээгүй сул хаалт байхаас дээр.
const cfg = readFileSync(join(SRC, 'brand.config.ts'), 'utf8');
const own = [...cfg.matchAll(/^\s*name:\s*'([^']+)'/gm)].map((m) => m[1]);
if (own.length === 0) {
  console.error('✗ brand.config.ts-ээс brand.name уншигдсангүй.');
  process.exit(1);
}

const esc = (x) => x.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
const PATTERN = new RegExp([...new Set([...FLEET, ...own])].map(esc).join('|'));

// Зөвшөөрөгдсөн — эдгээр нь БҮГД платформын өөрийн АГУУЛГА бөгөөд дундын
// кодод хэзээ ч ордоггүй. Брэндийн нэрийг агуулах нь тэдний зорилго:
//
//   brand.config.ts       — тодорхойлолт өөрөө;
//   components/landing/** — маркетингийн текст;
//   lib/*I18n.ts          — платформын толь (гар утасны апп-ыг нэрээр заадаг);
//   lib/apiCatalog.ts     — API каталог: БУСАД платформын бүтээгдэхүүнийг
//                           нэрлэн танилцуулах нь түүний цорын ганц зорилго.
const ALLOW = [
  /^brand\.config\.ts$/,
  /^components\/landing\//,
  /^lib\/[a-z]+I18n\.ts$/,
  /^lib\/apiCatalog\.ts$/,
];

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
