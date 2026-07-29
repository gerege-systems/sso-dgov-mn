import React from 'react';
import { redirect } from 'next/navigation';
import AiChatView from '@gerege/ui-core/components/me/AiChatView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('AI туслах') };

export default async function MeAiPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/ai');
  return <AiChatView />;
}
