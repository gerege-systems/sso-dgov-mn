import { authedFetch } from '@/lib/api';
import { proxyResult, checkOrigin, checkUUID } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// POST /api/gateway/keys/{keyId}/revoke — API key-г хүчингүй болгох.
export async function POST(req: Request, props: { params: Promise<{ keyId: string }> }) {
  const params = await props.params;
  const bad = checkOrigin(req) ?? checkUUID(params.keyId);
  if (bad) return bad;
  return proxyResult(await authedFetch(`/gateway/keys/${params.keyId}/revoke`, { method: 'POST' }));
}
