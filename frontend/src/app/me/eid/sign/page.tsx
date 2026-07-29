import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import EidSignView from '@gerege/ui-core/components/me/eid/EidSignView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Гарын үсэг зурах') };

export default async function EidSignPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/eid/sign');
  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="eid.sign.title" subKey="eid.sign.sub" />
      <EidSignView />
    </>
  );
}
