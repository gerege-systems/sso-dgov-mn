// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// eID PKI самбарын client-side туслахууд. getJSON нь HTTP статус алддаг тул
// PKI_READ эрхгүй (403)-ыг "эрх хүлээгдэж байна" төлөв болгон ялгахын тулд
// статусыг буцаадаг pkiGet-ийг эндээс хуваалцна (Dashboard + Profile хоёулаа).

export interface PkiSummary {
  certificates: { valid: number; revoked: number; expired: number; suspended: number; total: number };
  activity: { authentication: number; signature: number };
  devices_active: number;
  devices_total: number;
  representation_count: number;
}

export interface PkiCertItem {
  document_number: string;
  type: string;
  serial_number: string;
  certificate_level: string;
  status: string;
  not_before?: string;
  not_after?: string;
  issuer_dn?: string;
}

export interface PkiDeviceItem {
  document_number: string;
  platform?: string;
  active: boolean;
  enrolled_at?: string;
  deactivated_at?: string;
}

export interface PkiActItem {
  session_id?: string;
  flow: string;
  outcome: string;
  doc_text?: string;
  timestamp?: string;
}

/** pkiGet нь backend PKI endpoint-ыг дуудаж {status, data}-г бүрэн буцаана. */
export async function pkiGet<T>(path: string): Promise<{ status: number; data: T | null }> {
  const res = await fetch(path, { method: 'GET' });
  const body = await res.json().catch(() => null);
  return { status: body?.status ?? res.status, data: (body?.ok ? body.data : null) as T | null };
}
