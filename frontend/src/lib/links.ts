// Gerege Systems Development Team & Claude AI, 2026
//
// Гадаад холбоосуудын нэг эх сурвалж. Баримт бичгийн (docs) сайтын хаяг —
// NEXT_PUBLIC_DOCS_URL орчны хувьсагчаар дарж болно; байхгүй бол GitHub Pages
// дээрх нийтийн баримт бичгийн сайт руу заана.

/** Нийтийн техник баримт бичгийн сайт (MkDocs · GitHub Pages). */
export const DOCS_URL =
  process.env.NEXT_PUBLIC_DOCS_URL ??
  'https://gerege-systems.github.io/sso-dgov-mn-documentation/';
