import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import LanguageManager from '@gerege/ui-core/components/admin/LanguageManager';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { isSuperAdmin } from '@gerege/ui-core/lib/types';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Хэл — Супер админ' };

export default async function AdminLanguagesPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/languages');
  // Зөвхөн super admin — хэл нэмэх/хасах нь БҮХ хэрэглэгчийн харах интерфейсийг
  // өөрчилдөг тул энгийн админд ч нээхгүй (backend мөн адил хаадаг).
  if (!isSuperAdmin(me.roleId)) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="langs.title" subKey="langs.sub" />
      <LanguageManager />
    </>
  );
}
