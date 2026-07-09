import { authedFetch } from '@/lib/api';
import { proxyResult, readJson, checkOrigin, checkUUID } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// PUT /api/gateway/policies/{id} — policy шинэчлэх.
export async function PUT(req: Request, props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const bad = checkOrigin(req) ?? checkUUID(params.id);
  if (bad) return bad;
  const body = await readJson(req);
  return proxyResult(await authedFetch(`/gateway/policies/${params.id}`, { method: 'PUT', body: JSON.stringify(body) }));
}

// DELETE /api/gateway/policies/{id} — policy устгах.
export async function DELETE(req: Request, props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const bad = checkOrigin(req) ?? checkUUID(params.id);
  if (bad) return bad;
  return proxyResult(await authedFetch(`/gateway/policies/${params.id}`, { method: 'DELETE' }));
}
