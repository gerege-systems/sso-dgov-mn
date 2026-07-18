// Government Template Platform V3.0
// Gerege Systems Development Team & Claude AI, 2026
//
// Government SSO нүүр (landing) хуудасны маркетингийн текст — mn / en хосоор.
// Апп-ын үндсэн dict (lib/i18n.ts)-ийг бөглөхгүйн тулд landing-ийн урт мөрүүдийг
// энд төвлөрүүлэв. Бүх түлхүүр хоёр хэлэнд адил байх ёстой (i18n.ts-тэй нэг зарчим).

export interface LandingCopy {
  /** Брэнд нэр (nav + footer). Хоосон бол 'Government SSO'. Theme-ээр солино. */
  brand?: string;
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
  brand: 'Төрийн нэгдсэн нэвтрэлт',
  nav: { features: 'Боломжууд', security: 'Аюулгүй байдал', tech: 'Технологи', login: 'Нэвтрэх' },
  hero: {
    badge: 'Үндэсний цахим нэвтрэлт · eID',
    titleLead: 'Нэг нэвтрэлт —',
    titleAccent: 'бүх төрийн',
    titleTail: 'үйлчилгээ',
    lede:
      'Төрийн нэгдсэн нэвтрэлт нь үндэсний цахим үнэмлэг (eID)-т тулгуурласан төрийн нэгдсэн нэвтрэлт юм. Нэг удаа баталгаажуулснаар холбогдсон бүх төрийн болон хувийн үйлчилгээнд дахин нэвтрэлгүйгээр, найдвартай орно.',
    ctaLogin: 'eID-ээр нэвтрэх',
    ctaExplore: 'Боломжийг үзэх',
    stackLabel: 'Дэмждэг стандартууд',
    stats: [
      { value: 'eID', label: 'Цахим үнэмлэгээр баталгаажна' },
      { value: 'OAuth2 · OIDC', label: 'Олон улсын нээлттэй стандарт' },
      { value: 'SSO', label: 'Нэг нэвтрэлт — олон систем' },
    ],
  },
  advantages: {
    heading: 'Бүрэн гүйцэд, найдвартай нэвтрэлт',
    sub: 'Иргэн төвтэй, өндөр найдвартай төрийн нэвтрэлтийн онцлог шаардлагад нийцүүлэн эхнээс нь нямбай бүтээв.',
    eidTag: 'Баталгаажуулалт',
    eidTitle: 'Цахим үнэмлэгээр хормын дотор',
    eidBody:
      'Гар утасны цахим үнэмлэгийн апп руу шууд мэдэгдэл илгээх, эсвэл QR код уншуулан хормын дотор нэвтэрнэ. Нэг утсан дээр ч, компьютер-утас хослуулан ч ажиллана. Нэг ч нууц үг цээжлэх шаардлагагүй.',
    googleTitle: 'Google холболт',
    googleBody:
      'Цахим үнэмлэгээрээ нэг удаа баталгаажуулж Google дансаа холбоно. Улмаар нэг товшилтоор, аюулгүй байдлаа огт алдалгүйгээр хурдан нэвтэрнэ.',
    secTitle: 'Өндөр түвшний хамгаалалт',
    secBody:
      'Нэвтрэлтийн түлхүүр зөвхөн серверт хадгалагдаж, хөтчийн код руу хэзээ ч ил гардаггүй. Давхар хамгаалалт, өгөгдлийн мөр бүрийн хандалтын хяналт, хүсэлтийн ухаалаг хязгаарлалтаар бүрхэгдсэн.',
    ssoTitle: 'Бусад системд нэгдсэн нэвтрэлт (SSO)',
    ssoBody:
      'Government SSO-д холбогдсон аппликейшнүүд өөрсдийн нэвтрэлтийг Government SSO-гоор гүйцэтгүүлж, хэрэглэгчийн баталгаажсан мэдээллийг олон улсын нээлттэй стандартаар аюулгүй хүлээн авна. Хэрэглэгч нэг л удаа нэвтэрч, холбогдсон бүх системд орно.',
    signTitle: 'Цахим гарын үсгийн зуучлал',
    signBody:
      'Холбогдсон системүүд Government SSO-ийн итгэмжлэлээр дамжуулан баримт бичигт цахим гарын үсэг зуруулж чадна — өөрсдөө тусад нь гэрчилгээ эзэмших шаардлагагүйгээр.',
    consentTitle: 'Зөвшөөрлийг санана',
    consentBody:
      'Аппликейшн бүр таны зөвшөөрлийг зөвхөн анх удаа асууна. Дараа нь дахин төвөг учруулахгүй — жигд, тасралтгүй туршлага.',
  },
  tech: {
    heading: 'Орчин үеийн, найдвартай технологи дээр',
    sub: 'Хурд, найдвар, аюулгүй байдлыг эрхэмлэн сонгосон бүрэлдэхүүн.',
    backendTitle: 'Сервер тал — Go, Ory Hydra',
    backendBody:
      'Цэгцтэй бүтэцтэй Go сервер, PostgreSQL өгөгдлийн сан, Redis түргэн санах ой. Нэвтрэлтийн стандартуудыг олон улсад батжсан Ory Hydra хөдөлгүүрээр гүйцэтгэнэ.',
    frontendTitle: 'Хэрэглэгчийн тал — Next.js',
    frontendBody:
      'Хөтч зөвхөн өөрийн эх сурвалжтай харилцаж, серверийн тал нь дотоод системтэй холбогдоно. Нэвтрэлтийн түлхүүр хэрэглэгчийн код руу хэзээ ч гардаггүй.',
    aiTitle: 'Gemini хиймэл оюун туслах',
    aiBody:
      'Серверт ажилладаг ухаалаг туслах — асуултад хариулж, шаардлагатай үед монгол хэлээр найдвартай хариу өгнө.',
    trustTitle: 'Найдварын баталгаа',
    trustBadge: 'ИДЭВХТЭЙ',
    trustItems: [
      'Цахим үнэмлэгээр баталгаажуулалт',
      'Олон улсын нээлттэй стандарт',
      'Түлхүүр зөвхөн серверт нууцлагдана',
      'Мөр бүрийн хандалтын хяналт',
      'Олон давхар хамгаалалт, хязгаарлалт',
    ],
  },
  everything: {
    heading: 'Бүх боломж нэг дор',
    sub: 'Иргэн ба хөгжүүлэгчдэд эхний өдрөөс бэлэн.',
    items: [
      { title: 'Апп руу мэдэгдэл', body: 'Регистрийн дугаараар иргэний апп руу шууд илгээнэ.' },
      { title: 'QR болон шууд холбоос', body: 'Компьютер дээр QR, утсан дээр апп шууд нээгдэнэ.' },
      { title: 'Google холболт', body: 'Нэг удаа баталгаажуулснаар цаашид түргэн нэвтрэлт.' },
      { title: 'Баталгаат мэдээлэл', body: 'Холбогдсон систем баталгаажсан профайлыг стандартаар авна.' },
      { title: 'Зөвшөөрөл санах', body: 'Анх удаа л асууж, дараа нь жигд урсгал.' },
      { title: 'Хоёр хэл (mn/en)', body: 'Бүх дэлгэц монгол болон англиар.' },
      { title: 'Гэрэл / харанхуй горим', body: 'Системийн тохиргоонд тохирсон харагдац.' },
      { title: 'Олон давхар хамгаалалт', body: 'Хүсэлт бүрийг шалгаж, халдлагаас сэргийлнэ.' },
    ],
  },
  cta: {
    title: 'Одоо eID-ээр нэвтэрнэ үү',
    sub: 'Цахим үнэмлэгээ бэлдээд, нэг л баталгаажуулалтаар холбогдсон бүх үйлчилгээнд аюулгүй нэвтэрнэ.',
    ctaLogin: 'eID-ээр нэвтрэх',
    ctaExplore: 'Боломжийг үзэх',
    tagline: 'Цахим үнэмлэг · Нээлттэй стандарт · Найдвартай хамгаалалт',
  },
  footer: {
    tagline: 'Үндэсний цахим үнэмлэгт суурилсан төрийн нэгдсэн нэвтрэлт. Gerege Systems, 2026.',
    links: ['Үйлчилгээний нөхцөл', 'Нууцлалын бодлого', 'Холбоо барих'],
    copyright: '© 2026 Gerege Systems · Government SSO',
  },
};

const en: LandingCopy = {
  brand: 'Government SSO',
  nav: { features: 'Features', security: 'Security', tech: 'Technology', login: 'Sign in' },
  hero: {
    badge: 'National Single Sign-On · eID',
    titleLead: 'One identity —',
    titleAccent: 'every government',
    titleTail: 'service',
    lede:
      'Government SSO is a national single sign-on built on the electronic ID (eID). Verify once, then access every connected government and private service securely — without signing in again.',
    ctaLogin: 'Sign in with eID',
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
      'An OpenID Connect provider built on Ory Hydra. Relying applications (RPs) delegate sign-in to Government SSO and receive verified user data as standard claims.',
    signTitle: 'Signature relay',
    signBody:
      'Third-party RPs can have documents e-signed through Government SSO’s eID RP credentials — without holding their own eID certificates.',
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
    title: 'Sign in with eID now',
    sub: 'Have your electronic ID ready and, with a single verification, access every connected service securely.',
    ctaLogin: 'Sign in with eID',
    ctaExplore: 'Explore features',
    tagline: 'eID-based · Open standards · Secure by design',
  },
  footer: {
    tagline: 'National single sign-on built on the electronic ID. Gerege Systems, 2026.',
    links: ['Terms of Service', 'Privacy Policy', 'Contact'],
    copyright: '© 2026 Gerege Systems · Government SSO',
  },
};

export const landingCopy: Record<'mn' | 'en', LandingCopy> = { mn, en };
