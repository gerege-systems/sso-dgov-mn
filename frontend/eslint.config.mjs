// ESLint flat config (eslint 10 + eslint-config-next 16).
//
// eslint 10 нь `.eslintrc` форматыг огт дэмжихээ больсон тул өмнөх
// `.eslintrc.json` ({"extends": "next/core-web-vitals"})-ыг үүгээр сольсон.
//
// ХОЁР ТОХИРУУЛГА (eslint 10-ийн улмаас) — доорх тайлбарыг уншина уу:
//  1. `react/*` дүрмийг хасна: eslint-config-next 16 нь eslint-plugin-react
//     7.37.5-ыг авчирдаг бөгөөд түүний peer нь eslint ^9.7 хүртэл. eslint 10
//     дээр уг plugin дотроо унадаг (Components.js — TypeError). Бусад бүх Next
//     дүрэм (@next/next, react-hooks, jsx-a11y, import, typescript-eslint)
//     eslint 10 дээр асуудалгүй ажиллана.
//  2. `.js/.mjs`-д parser-ыг шууд зааж өгнө: eslint-config-next-ийн өөрийн
//     parser нь eslint 10-ийн шаарддаг scopeManager.addGlobals-гүй тул
//     `next.config.mjs` мэтийн файл дээр унадаг.
//
// eslint-plugin-react eslint 10-ыг дэмжмэгц эдгээрийг ХОЁУЛАНГ нь устгаж,
// `export default nextCoreWebVitals` болгож хялбарчилна.
import nextCoreWebVitals from 'eslint-config-next/core-web-vitals';
import tsParser from '@typescript-eslint/parser';

const stripReactPlugin = (cfg) => {
  const { react, ...plugins } = cfg.plugins ?? {};
  const rules = Object.fromEntries(
    Object.entries(cfg.rules ?? {}).filter(([name]) => !name.startsWith('react/')),
  );
  const out = { ...cfg };
  if (cfg.plugins) out.plugins = plugins;
  if (cfg.rules) out.rules = rules;
  return out;
};

// Массивыг шууд default болгож экспортлохгүй — eslint нь тохиргоог статикаар
// уншихын тулд нэрлэсэн хувьсагчийг илүүд үздэг (config-array дүрэм).
const config = [
  // Багцын өөрийн ignores нь .next/**, out/**, build/**, next-env.d.ts-ыг хамардаг.
  ...nextCoreWebVitals.map(stripReactPlugin),
  { ignores: ['node_modules/**', 'coverage/**', 'public/**'] },
  {
    files: ['**/*.{js,jsx,mjs,cjs}'],
    languageOptions: { parser: tsParser },
  },
  {
    rules: {
      // react-hooks 7-гийн шинэ (React Compiler эриний) дүрмүүд. Манай кодын
      // "effect дотор localStorage уншаад setState" хэв маяг нь SSR/hydration-д
      // зориулсан САНААТАЙ шийдэл (сервер дээр localStorage байхгүй) — 15 газарт
      // байгаа бөгөөд useSyncExternalStore руу шилжүүлэх нь тусдаа, hydration
      // эрсдэлтэй ажил. Тиймээс одоохондоо warning: CI-г зогсоохгүй ч харагдана.
      // TODO(frontend): эдгээрийг үе шаттайгаар зассаны дараа error болгох.
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/purity': 'warn',
    },
  },
];

export default config;
