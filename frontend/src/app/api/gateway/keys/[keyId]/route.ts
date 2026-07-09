import { authedFetch } from '@/lib/api';
import { proxyResult, checkOrigin, checkUUID } from '@/lib/bff';

export const dynamic = 'force-dynamic';

// DELETE /api/gateway/keys/{keyId} — API key устгах.
export async function DELETE(req: Request, props: { params: Promise<{ keyId: string }> }) {
  const params = await props.params;
  const bad = checkOrigin(req) ?? checkUUID(params.keyId);
  if (bad) return bad;
  return proxyResult(await authedFetch(`/gateway/keys/${params.keyId}`, { method: 'DELETE' }));
}
