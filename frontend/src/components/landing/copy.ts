// Government Template Platform V3.0
// Gerege Systems Development Team & Claude AI, 2026
//
// DAN-Government SSO нүүр (landing) хуудасны маркетингийн текст — mn / en хосоор.
// Апп-ын үндсэн dict (lib/i18n.ts)-ийг бөглөхгүйн тулд landing-ийн урт мөрүүдийг
// энд төвлөрүүлэв. Бүх түлхүүр хоёр хэлэнд адил байх ёстой (i18n.ts-тэй нэг зарчим).

export interface LandingCopy {
  nav: { features: string; security: string; tech: string; login: string };
  hero: {
    badge: string;
    titleLead: string;
    titleAccent: string;
    titleTail: string;
    lede: string;
    ctaLogin: string;
    ctaExplore: string;
    stackLabel: string;
    stats: { value: string; label: string }[];
  };
  advantages: {
    heading: string;
    sub: string;
    eidTag: string;
    eidTitle: string;
    eidBody: string;
    googleTitle: string;
    googleBody: string;
    secTitle: string;
    secBody: string;
    ssoTitle: string;
    ssoBody: string;
    signTitle: string;
    signBody: string;
    consentTitle: string;
    consentBody: string;
  };
  tech: {
    heading: string;
    sub: string;
    backendTitle: string;
    backendBody: string;
    frontendTitle: string;
    frontendBody: string;
    aiTitle: string;
    aiBody: string;
    trustTitle: string;
    trustBadge: string;
    trustItems: string[];
  };
  everything: { heading: string; sub: string; items: { title: string; body: string }[] };
  cta: { title: string; sub: string; ctaLogin: string; ctaExplore: string; tagline: string };
  footer: { tagline: string; links: string[]; copyright: string };
}

const mn: LandingCopy = {
  nav: { features: 'Онцлог', security: 'Аюулгүй байдал', tech: 'Технологи', login: 'Нэвтрэх' },
  hero: {
    badge: 'Үндэсний нэгдсэн нэвтрэлт · eID',
    titleLead: 'Нэг бүртгэл —',
    titleAccent: 'бүх төрийн',
    titleTail: 'үйлчилгээ',
    lede:
      'DAN-Government SSO нь үндэсний цахим үнэмлэх (eID)-д суурилсан нэгдсэн нэвтрэлтийн систем. Нэг л удаа баталгаажаад, холбогдсон бүх төрийн болон хувийн үйлчилгээнд аюулгүйгээр, дахин нэвтрэхгүйгээр орно.',
    ctaLogin: 'DAN-аар нэвтрэх',
    ctaExplore: 'Боломжийг үзэх',
    stackLabel: 'Дэмждэг стандарт',
    stats: [
      { value: 'eID', label: 'Цахим үнэмлэгээр танилт' },
      { value: 'OAuth2 · OIDC', label: 'Нээлттэй стандарт' },
      { value: 'SSO', label: 'Нэг нэвтрэлт — олон систем' },
    ],
  },
  advantages: {
    heading: 'Юуг ч дутуугүй, аюулгүй нэвтрэлт',
    sub: 'Иргэн төвтэй, өндөр аюулгүй байдалтай төрийн нэвтрэлтийн онцгой шаардлагыг эхнээс нь бодож зохион бүтээв.',
    eidTag: 'Танилт',
    eidTitle: 'eID цахим үнэмлэгээр хормын дотор',
    eidBody:
      'Гар утасны eID апп руу шууд илгээх (push) эсвэл QR уншуулах замаар — нэг төхөөрөмж дээр App2App, олон төхөөрөмж хооронд ч найдвартай, баталгаажсан нэвтрэлт. Нууц үг цээжлэх шаардлагагүй.',
    googleTitle: 'Google холболт',
    googleBody:
      'eID-ээр нэг удаа баталгаажуулан Google дансаа холбоод, дараа нь нэг товшилтоор түргэн нэвтэрнэ — аюулгүй байдлаа алдалгүйгээр хялбар.',
    secTitle: 'Аюулгүйгээр чангалсан',
    secBody:
      'httpOnly cookie дэх токен (browser JS-т хүрдэггүй), давхар CSRF хамгаалалт, мөрийн түвшний хамгаалалт (RLS), CSP/HSTS толгойнууд, IP тус бүрийн rate-limit.',
    ssoTitle: 'Гуравдагч системүүдийн SSO (OAuth2 / OIDC)',
    ssoBody:
      'Ory Hydra дээр суурилсан OpenID Connect үйлчилгээ үзүүлэгч. Холбогдсон аппликейшнүүд (RP) DAN-аар нэвтрэлтээ гүйцэтгэж, хэрэглэгчийн баталгаажсан мэдээллийг стандарт claim-аар хүлээн авна.',
    signTitle: 'Гарын үсгийн реле (Sign-relay)',
    signBody:
      'Гуравдагч RP-үүд DAN-ий eID RP итгэмжлэлээр дамжуулан баримтад цахим гарын үсэг зуруулах боломжтой — өөрсдөө eID гэрчилгээ эзэмших шаардлагагүйгээр.',
    consentTitle: 'Зөвшөөрлийг санадаг',
    consentBody:
      'Аппликейшн бүр эхний удаад л таны зөвшөөрлийг асууна. Дараа нь давтан асуухгүй — хэрэглэгчид жигд, тасралтгүй туршлага.',
  },
  tech: {
    heading: 'Батжсан, орчин үеийн технологи дээр',
    sub: 'Хурд, найдвар, аюулгүй байдлыг эхнээс нь бодсон бүрэлдэхүүн.',
    backendTitle: 'Go backend + Ory Hydra',
    backendBody:
      'Clean Architecture-тай Go (chi · net/http) backend, pgx-ээр PostgreSQL, Redis кэш. OAuth2/OIDC-ийг батжсан Ory Hydra хөдөлгүүрээр гүйцэтгэнэ.',
    frontendTitle: 'Next.js Frontend (BFF)',
    frontendBody:
      'Browser зөвхөн ижил-эх (same-origin) Next.js route-той харилцаж, backend руу серверийн талд proxy хийнэ. Токен client JS-т хэзээ ч хүрэхгүй.',
    aiTitle: 'Gemini AI туслах',
    aiBody:
      'SDK-гүй REST pipeline, серверт ажилладаг tools, DB-ээр тохируулах scope/зааварчилгаа, доголдоход монгол хэлээр найдвартай fallback.',
    trustTitle: 'Итгэлийн баталгаа',
    trustBadge: 'ПРОДАКШН',
    trustItems: [
      'eID цахим үнэмлэгээр танилт',
      'OAuth2 / OpenID Connect стандарт',
      'httpOnly cookie · CSRF хамгаалалт',
      'PostgreSQL RLS мөрийн хамгаалалт',
      'CSP · HSTS · rate-limit',
    ],
  },
  everything: {
    heading: 'Бүх боломж нэг дор',
    sub: 'Иргэн ба хөгжүүлэгчдэд эхний өдрөөс бэлэн.',
    items: [
      { title: 'eID push нэвтрэлт', body: 'РД-ээр иргэний апп руу шууд илгээж зөвшөөрүүлнэ.' },
      { title: 'QR · App2App', body: 'Desktop дээр QR, утсан дээр deep-link — хоёр талдаа.' },
      { title: 'Google холбоос', body: 'eID-баталгаажсан эхний холболтоор түргэн нэвтрэлт.' },
      { title: 'OIDC claim-ууд', body: 'RP-үүд баталгаажсан профайлыг стандартаар хүлээн авна.' },
      { title: 'Зөвшөөрөл санах', body: 'Эхний удаад л асууж, дараа нь жигд урсгал.' },
      { title: 'Хоёр хэл (mn/en)', body: 'Бүх дэлгэц монгол болон англиар.' },
      { title: 'Гэрэл / харанхуй', body: 'Системийн загварт тохирсон theme.' },
      { title: 'Аюулгүйн толгойнууд', body: 'CSP, HSTS, COOP/COEP, origin шалгалт.' },
    ],
  },
  cta: {
    title: 'Одоо DAN-аар нэвтэрнэ үү',
    sub: 'Цахим үнэмлэгээ бэлдээд, нэг л баталгаажуулалтаар холбогдсон бүх үйлчилгээнд аюулгүйгээр орно.',
    ctaLogin: 'DAN-аар нэвтрэх',
    ctaExplore: 'Онцлогийг үзэх',
    tagline: 'eID суурьтай · Нээлттэй стандарт · Аюулгүйгээр зохион бүтээсэн',
  },
  footer: {
    tagline: 'Үндэсний цахим үнэмлэгт суурилсан төрийн нэгдсэн нэвтрэлт. Gerege Systems, 2026.',
    links: ['Үйлчилгээний нөхцөл', 'Нууцлалын бодлого', 'Холбоо барих'],
    copyright: '© 2026 Gerege Systems · DAN-Government SSO',
  },
};

const en: LandingCopy = {
  nav: { features: 'Features', security: 'Security', tech: 'Technology', login: 'Sign in' },
  hero: {
    badge: 'National Single Sign-On · eID',
    titleLead: 'One identity —',
    titleAccent: 'every government',
    titleTail: 'service',
    lede:
      'DAN-Government SSO is a national single sign-on built on the electronic ID (eID). Verify once, then access every connected government and private service securely — without signing in again.',
    ctaLogin: 'Sign in with DAN',
    ctaExplore: 'Explore features',
    stackLabel: 'Standards supported',
    stats: [
      { value: 'eID', label: 'Electronic-ID identity' },
      { value: 'OAuth2 · OIDC', label: 'Open standards' },
      { value: 'SSO', label: 'One login — many systems' },
    ],
  },
  advantages: {
    heading: 'Complete, secure sign-in',
    sub: 'Engineered from the ground up for the demands of citizen-centric, high-assurance government access.',
    eidTag: 'Authentication',
    eidTitle: 'Instant eID sign-in',
    eidBody:
      'Push straight to the eID app or scan a QR — App2App on one device, reliable cross-device flows on two. Verified identity with no passwords to remember.',
    googleTitle: 'Google linking',
    googleBody:
      'Link your Google account once behind an eID verification, then sign in with a single tap — convenience without giving up assurance.',
    secTitle: 'Security hardened',
    secBody:
      'Tokens in httpOnly cookies (never exposed to browser JS), double CSRF defense, row-level security (RLS), CSP/HSTS headers and per-IP rate limiting.',
    ssoTitle: 'SSO for third parties (OAuth2 / OIDC)',
    ssoBody:
      'An OpenID Connect provider built on Ory Hydra. Relying applications (RPs) delegate sign-in to DAN and receive verified user data as standard claims.',
    signTitle: 'Signature relay',
    signBody:
      'Third-party RPs can have documents e-signed through DAN’s eID RP credentials — without holding their own eID certificates.',
    consentTitle: 'Remembers consent',
    consentBody:
      'Each application asks for your consent only the first time. After that it never re-prompts — a smooth, uninterrupted experience.',
  },
  tech: {
    heading: 'On a modern, proven stack',
    sub: 'Components chosen for speed, reliability and security from day one.',
    backendTitle: 'Go backend + Ory Hydra',
    backendBody:
      'A Clean-Architecture Go (chi · net/http) backend with pgx over PostgreSQL and Redis caching. OAuth2/OIDC is powered by the battle-tested Ory Hydra engine.',
    frontendTitle: 'Next.js frontend (BFF)',
    frontendBody:
      'The browser talks only to same-origin Next.js routes, which proxy to the backend server-side. Tokens never reach client JS.',
    aiTitle: 'Gemini AI assistant',
    aiBody:
      'An SDK-free REST pipeline with server-side tools, DB-configurable scope/instructions and a resilient Mongolian fallback on failure.',
    trustTitle: 'Trust guarantees',
    trustBadge: 'PRODUCTION',
    trustItems: [
      'eID electronic-ID identity',
      'OAuth2 / OpenID Connect standard',
      'httpOnly cookies · CSRF defense',
      'PostgreSQL row-level security',
      'CSP · HSTS · rate limiting',
    ],
  },
  everything: {
    heading: 'Every capability in one place',
    sub: 'Ready for citizens and developers from day one.',
    items: [
      { title: 'eID push sign-in', body: 'Push to the citizen app by ID number and approve.' },
      { title: 'QR · App2App', body: 'QR on desktop, deep-link on mobile — both directions.' },
      { title: 'Google linking', body: 'Fast sign-in after an eID-verified first link.' },
      { title: 'OIDC claims', body: 'RPs receive the verified profile as standard claims.' },
      { title: 'Consent memory', body: 'Asked once, then a seamless flow afterward.' },
      { title: 'Bilingual (mn/en)', body: 'Every screen in Mongolian and English.' },
      { title: 'Light / dark', body: 'A theme that follows the system design.' },
      { title: 'Security headers', body: 'CSP, HSTS, COOP/COEP and origin checks.' },
    ],
  },
  cta: {
    title: 'Sign in with DAN now',
    sub: 'Have your electronic ID ready and, with a single verification, access every connected service securely.',
    ctaLogin: 'Sign in with DAN',
    ctaExplore: 'Explore features',
    tagline: 'eID-based · Open standards · Secure by design',
  },
  footer: {
    tagline: 'National single sign-on built on the electronic ID. Gerege Systems, 2026.',
    links: ['Terms of Service', 'Privacy Policy', 'Contact'],
    copyright: '© 2026 Gerege Systems · DAN-Government SSO',
  },
};

export const landingCopy: Record<'mn' | 'en', LandingCopy> = { mn, en };
