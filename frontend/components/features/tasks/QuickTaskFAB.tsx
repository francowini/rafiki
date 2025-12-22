'use client';

import { useState } from 'react';
import { usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import { QuickTaskSheet } from './QuickTaskSheet';
import { cn } from '@/lib/utils';

export function QuickTaskFAB() {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();

  // Hide on objective detail pages (they have their own "+ Nueva tarea" button)
  const isObjectiveDetailPage = /^\/objectives\/[^/]+$/.test(pathname);

  if (isObjectiveDetailPage) {
    return null;
  }

  return (
    <>
      <Button
        size="lg"
        className={cn(
          'fixed bottom-6 right-6 z-40',
          'h-12 px-5 rounded-full shadow-lg',
          'hover:scale-105 active:scale-95 transition-transform',
          'bg-primary hover:bg-primary/90',
          'flex items-center gap-2',
        )}
        onClick={() => setOpen(true)}
      >
        <Plus className="h-5 w-5" />
        <span className="font-medium">Nueva tarea</span>
      </Button>

      <QuickTaskSheet open={open} onOpenChange={setOpen} />
    </>
  );
}
