// Applications (OAuth2 client — Ory Hydra) BFF (/api/applications/*) хариунуудын
// TypeScript хэлбэрүүд. API-Gateway-ийн consumer ба SSO-ийн RP бүртгэлийг нэгтгэсэн
// нэгдмэл загвар.

export type AppType = 'web' | 'spa' | 'native' | 'm2m';

export interface Application {
  id: string;
  client_id: string;
  name: string;
  app_type: AppType;
  tags: string[];
  redirect_uris: string[];
  enabled: boolean;
  service_ids: string[];
  /** Зөвхөн create/rotate хариунд ирнэ (нэг удаа харагдана). */
  secret?: string;
  created_at: string;
  updated_at?: string;
}

/** app_type сонголтын mn/en шошго. */
export const APP_TYPES: { value: AppType; mn: string; en: string }[] = [
  { value: 'web', mn: 'Веб апп', en: 'Web app' },
  { value: 'spa', mn: 'SPA', en: 'SPA' },
  { value: 'native', mn: 'Мобайл/Native', en: 'Mobile / Native' },
  { value: 'm2m', mn: 'Сервер-хоорондын (M2M)', en: 'Server-to-server (M2M)' },
];

/** app_type нь redirect URI шаарддаг (OAuth redirect flow) эсэх. */
export function needsRedirect(t: AppType): boolean {
  return t === 'web' || t === 'spa' || t === 'native';
}

/** app_type нь нууц түлхүүртэй (confidential) эсэх — secret rotate зөвшөөрөгдөнө. */
export function isConfidential(t: AppType): boolean {
  return t === 'web' || t === 'm2m';
}
