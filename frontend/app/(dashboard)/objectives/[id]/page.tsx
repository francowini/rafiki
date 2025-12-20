'use client';

import { useParams, notFound } from 'next/navigation';
import { ObjectiveDetail } from '@/components/features/objectives/ObjectiveDetail';

export default function ObjectiveDetailPage() {
  const params = useParams();

  // Validate params.id: handle undefined, array, or invalid values
  const rawId = params?.id;
  const id = Array.isArray(rawId) ? rawId[0] : rawId;

  if (!id || typeof id !== 'string') {
    notFound();
  }

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      <ObjectiveDetail objectiveId={id} />
    </div>
  );
}
