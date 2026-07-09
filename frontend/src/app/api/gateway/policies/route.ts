import { authedFetch } from '@/lib/api';
import { proxyResult, readJson, checkOrigin } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// GET /api/gateway/policies — policy (plugin)-ууд маршрутын нэртэйгээ.
export async function GET() {
  return proxyResult(await authedFetch('/gateway/policies', { method: 'GET' }));
}

// POST /api/gateway/policies — шинэ policy үүсгэх.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson(req);
  return proxyResult(await authedFetch('/gateway/policies', { method: 'POST', body: JSON.stringify(body) }));
}
