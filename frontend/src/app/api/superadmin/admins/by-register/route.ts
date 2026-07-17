import { authedFetch } from '@/lib/api';
import { proxyResult, readJson, checkOrigin } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/superadmin/admins/by-register — DAN-д бүртгэлтэй байгаа хэрэглэгчийг
// регистрийн дугаараар нь админ болгох. Register DAN-д бүртгэлгүй бол backend 404.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson(req);
  return proxyResult(
    await authedFetch('/superadmin/admins/by-register', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  );
}
