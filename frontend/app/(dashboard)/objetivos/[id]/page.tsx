'use client';

import { useParams } from 'next/navigation';
import { ObjectiveDetail } from '@/components/features/objectives/ObjectiveDetail';

export default function ObjectiveDetailPage() {
  const params = useParams();
  const id = params.id as string;

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      <ObjectiveDetail objectiveId={id} />
    </div>
  );
}
