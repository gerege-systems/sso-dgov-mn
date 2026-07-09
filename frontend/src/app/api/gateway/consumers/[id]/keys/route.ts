import { authedFetch } from '@/lib/api';
import { proxyResult, readJson, checkOrigin, checkUUID } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// GET /api/gateway/consumers/{id}/keys — consumer-ийн API key-үүд.
export async function GET(_req: Request, props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const bad = checkUUID(params.id);
  if (bad) return bad;
  return proxyResult(await authedFetch(`/gateway/consumers/${params.id}/keys`, { method: 'GET' }));
}

// POST /api/gateway/consumers/{id}/keys — шинэ API key үүсгэх (plaintext нэг удаа буцна).
export async function POST(req: Request, props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const bad = checkOrigin(req) ?? checkUUID(params.id);
  if (bad) return bad;
  const body = await readJson(req);
  return proxyResult(await authedFetch(`/gateway/consumers/${params.id}/keys`, { method: 'POST', body: JSON.stringify(body) }));
}
