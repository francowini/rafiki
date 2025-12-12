'use client';

import { useState, useSyncExternalStore } from 'react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { Home, Heart, Sparkles, Brain, Clock, ChevronLeft } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

const navItems = [
  { label: 'Dashboard', icon: Home, href: '/', color: 'text-gray-700' },
  { label: 'Values', icon: Heart, href: '/values', color: 'text-rose-600' },
  { label: 'Life Visions', icon: Sparkles, href: '/life-visions', color: 'text-purple-600' },
  { label: 'Thinks', icon: Brain, href: '/thinks', color: 'text-blue-600' },
  { label: 'Moments', icon: Clock, href: '/momentos', color: 'text-teal-600' },
];

function getSidebarCollapsed(): boolean {
  if (typeof window === 'undefined') return false;
  const saved = localStorage.getItem('sidebar-collapsed');
  return saved !== null ? JSON.parse(saved) : false;
}

function subscribeSidebarState(callback: () => void) {
  window.addEventListener('storage', callback);
  return () => window.removeEventListener('storage', callback);
}

export function AppSidebar() {
  const initialCollapsed = useSyncExternalStore(subscribeSidebarState, getSidebarCollapsed, () => false);
  const [collapsed, setCollapsed] = useState(initialCollapsed);
  const pathname = usePathname();

  const toggleCollapsed = () => {
    const newState = !collapsed;
    setCollapsed(newState);
    localStorage.setItem('sidebar-collapsed', JSON.stringify(newState));
  };

  return (
    <aside
      className={cn(
        'hidden lg:flex flex-col border-r bg-white h-screen sticky top-0 transition-all duration-300',
        collapsed ? 'w-16' : 'w-64',
      )}
    >
      <div className="p-4 border-b">
        <Link href="/" className="flex items-center gap-3">
          {collapsed ? (
            <span className="text-2xl font-bold text-gray-900">R</span>
          ) : (
            <span className="text-2xl font-bold text-gray-900">Rafiki</span>
          )}
        </Link>
      </div>

      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded-md transition-colors',
                isActive
                  ? 'bg-gray-100 text-gray-900 font-medium'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
                collapsed && 'justify-center',
              )}
              title={collapsed ? item.label : undefined}
            >
              <Icon className={cn('h-5 w-5 flex-shrink-0', item.color)} />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t">
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleCollapsed}
          className={cn('w-full', collapsed && 'justify-center')}
        >
          <ChevronLeft
            className={cn('h-4 w-4 transition-transform', collapsed && 'rotate-180')}
          />
          {!collapsed && <span className="ml-2">Collapse</span>}
        </Button>
      </div>
    </aside>
  );
}
