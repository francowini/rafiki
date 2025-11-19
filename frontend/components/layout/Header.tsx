'use client';

import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/lib/auth-context';
import { UserMenu } from '@/components/auth/UserMenu';

export function Header() {
  const { isAuthenticated } = useAuth();

  return (
    <header className="border-b bg-white">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-2xl font-bold text-gray-900">
            Rafiki
          </Link>
          {isAuthenticated && (
            <nav className="flex gap-4">
              <Link href="/thinks">
                <Button variant="ghost">Thinks</Button>
              </Link>
            </nav>
          )}
        </div>
        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <UserMenu />
          ) : (
            <Link href="/login">
              <Button>Login</Button>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
