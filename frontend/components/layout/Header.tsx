'use client';

import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { UserMenu } from '@/components/auth/UserMenu';
import { MobileNav } from '@/components/layout/MobileNav';

export function Header() {
  const { isAuthenticated } = useAuth();

  return (
    <header className="border-b bg-white sticky top-0 z-10">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          {isAuthenticated && <MobileNav />}
          <Link href="/" className="text-2xl font-bold text-gray-900 lg:hidden">
            Rafiki
          </Link>
        </div>

        <div className="flex items-center gap-4">{isAuthenticated && <UserMenu />}</div>
      </div>
    </header>
  );
}
