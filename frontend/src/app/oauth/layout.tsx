// OIDC provider (RP-facing) дэлгэцүүд — dan-ий ӨӨРИЙН дизайн (SigninShell: брэнд
// topbar + төвлөрсөн signin-card). dan-ий апп login (/login)-той ижил төрх.
import type { ReactNode } from 'react';
import SigninShell from '@gerege/ui-core/components/SigninShell';

export default function OAuthLayout({ children }: { children: ReactNode }) {
  return <SigninShell>{children}</SigninShell>;
}
