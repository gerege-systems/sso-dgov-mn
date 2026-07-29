import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import EidLogsView from '@gerege/ui-core/components/me/eid/EidLogsView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Үйл ажиллагаа') };

export default async function EidLogsPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/eid/logs');
  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="eid.logs.title" subKey="eid.logs.sub" />
      <EidLogsView show={!!me.eid} />
    </>
  );
}
