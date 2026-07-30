// Дундын багцын компонент ашигладаг CSS class энэ репод тодорхойлогдсон эсэх.
//
// Багц (`@gerege/ui-core`) нь CSS АГУУЛДАГГҮЙ — загвар нь репо бүрийн
// `globals.css`-д байрладаг. Иймд багц шинэ class нэрлэхэд, эсвэл репогийн CSS
// хуучирахад компонент ЧИМЭЭГҮЙ загваргүй болно: `<button>` нь хөтчийн анхдагч
// саарал, хүрээтэй чимэгтэй, `<table>` нь хүрээгүй, шахуу зурагдана. `tsc`,
// `lint`, `build` — бүгд ногоон хэвээр.
//
// Бодит тохиолдол (2026-07-30): хоёр SSO платформын `globals.css` эвхэгддэггүй
// ХУУЧИН цэсийнх байсан тул `.sidepanel__group-head` тодорхойлогдоогүй — зүүн
// цэсний бүлгийн толгой бүр анхдагч товч болж харагдаж байв.
//
// Яагаад Node дээр бичсэн бэ: `npm run build`-ийн нэг хэсэг тул Docker-ийн
// `node:20-alpine` дотор ч ажиллах ёстой (bash, GNU grep байхгүй).
//
// Дэлгэрэнгүй: vision-gerege-mn/UI_CORE_PLAN.md

import { readdirSync, readFileSync, statSync, existsSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('..', import.meta.url).pathname;
const PKG = join(ROOT, 'node_modules/@gerege/ui-core/src');
const APP_CSS_DIR = join(ROOT, 'src/app');

// Загварын нэр: block · block__elem · block--mod · block__elem--mod
const CLASS_RE = /^[a-z][a-z0-9]*(?:__[a-z0-9-]+)?(?:--[a-z0-9-]+)?$/;
// Зөвхөн ХУВИЛБАР (modifier) — хувьсагчид оноогдсон мөрийг барихад
const MODIFIER_RE = /^[a-z][a-z0-9]*(?:__[a-z0-9-]+)?--[a-z0-9-]+$/;

/**
 * Динамик нэр — багц `` `aichat__msg--${role}` `` гэж бичдэг тул бүтэн нэрийг
 * статикаар мэдэх боломжгүй. Тэдгээрийн БОДИТ хувилбарууд (`--user`, `--model`)
 * CSS-д байгаа эсэхийг гараар баталгаажуулсан; энд давтахгүй.
 */
const DYNAMIC_PREFIXES = ['aichat__msg--'];

function walk(dir, test, out = []) {
  if (!existsSync(dir)) return out;
  for (const e of readdirSync(dir)) {
    const p = join(dir, e);
    if (statSync(p).isDirectory()) walk(p, test, out);
    else if (test(e)) out.push(p);
  }
  return out;
}

// ── 1. Багцын компонентууд ямар class нэрлэдэг вэ ────────────────────────────
const used = new Map(); // class → эхний олдсон файл

function note(cls, file) {
  if (!CLASS_RE.test(cls)) return;
  if (cls.endsWith('--') || cls.endsWith('__')) return; // ${…}-ийн тасархай үлдэгдэл
  if (DYNAMIC_PREFIXES.some((p) => cls === p)) return;
  if (!used.has(cls)) used.set(cls, file);
}

for (const file of walk(PKG, (f) => f.endsWith('.tsx'))) {
  const src = readFileSync(file, 'utf8');
  const rel = relative(PKG, file);

  // className="…" · className={`…`} · className={'…'} — шууд хэлбэрүүд.
  // Template literal доторх ${…}-ийг зайгаар солино: түүнтэй залгаа тасархай
  // хэсэг («aichat__msg--») дээрх шалгуураар хаягдана.
  const attr = /className=(?:"([^"]*)"|\{`([^`]*)`\}|\{'([^']*)'\})/g;
  for (const m of src.matchAll(attr)) {
    const raw = (m[1] ?? m[2] ?? m[3]).replace(/\$\{[^}]*\}/g, ' ');
    for (const tok of raw.split(/\s+/)) note(tok, rel);
  }

  // Гурвалсан нөхцөлөөр хувьсагчид оноосон хувилбарууд:
  //   const klass = ok ? 'chip--success' : 'chip--pending';
  // Зөвхөн `--`-тэй мөрийг авна — өөр утгатай мөр ийм хэлбэртэй байх нь ховор.
  for (const m of src.matchAll(/'([a-z][a-z0-9]*(?:__[a-z0-9-]+)?--[a-z0-9-]+)'/g)) {
    if (MODIFIER_RE.test(m[1])) note(m[1], rel);
  }
}

// ── 2. Энэ репо юуг тодорхойлсон бэ ──────────────────────────────────────────
const defined = new Set();
for (const file of walk(APP_CSS_DIR, (f) => f.endsWith('.css'))) {
  const css = readFileSync(file, 'utf8');
  for (const m of css.matchAll(/\.([a-z][a-z0-9]*(?:__[a-z0-9-]+)?(?:--[a-z0-9-]+)?)/g)) {
    defined.add(m[1]);
  }
}

if (defined.size === 0) {
  console.error('✗ src/app дотор CSS олдсонгүй — энэ шалгалт утгагүй байна.');
  process.exit(1);
}

// ── 3. Тулгах ────────────────────────────────────────────────────────────────
const missing = [...used.keys()].filter((c) => !defined.has(c)).sort();

if (missing.length > 0) {
  console.error(`✗ Багцын ${missing.length} CSS class энэ репод тодорхойлогдоогүй:\n`);
  for (const c of missing) console.error(`    .${c}   (${used.get(c)})`);
  console.error('\n  Загваргүй class нь хөтчийн анхдагч чимэгээр зурагдана —');
  console.error('  товч саарал, хүснэгт хүрээгүй болно. Засах хоёр зам:');
  console.error('    1) src/app/globals.css -д дүрмийг нэмэх;');
  console.error('    2) багцад аль хэдийн БАЙГАА нэрийг хэрэглүүлэх (илүү дээр).');
  process.exit(1);
}

console.log(`✓ Багцын ${used.size} CSS class бүгд тодорхойлогдсон.`);
