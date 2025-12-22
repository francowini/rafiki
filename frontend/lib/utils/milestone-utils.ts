export interface MilestoneInfo {
  milestone: 25 | 50 | 75 | 100;
  title: string;
  description: string;
}

export function detectMilestone(
  previousProgress: number,
  newProgress: number,
  targetMetric: number,
): MilestoneInfo | null {
  const prevPercentage = (previousProgress / targetMetric) * 100;
  const newPercentage = (newProgress / targetMetric) * 100;

  const milestones: Array<{ threshold: number; title: string; description: string }> = [
    { threshold: 25, title: '25% completado', description: 'Buen comienzo' },
    { threshold: 50, title: '50% completado', description: 'A mitad de camino' },
    { threshold: 75, title: '75% completado', description: 'Casi lo logras' },
    { threshold: 100, title: 'Objetivo completado', description: 'Lo lograste' },
  ];

  for (const { threshold, title, description } of milestones) {
    if (prevPercentage < threshold && newPercentage >= threshold) {
      return { milestone: threshold as 25 | 50 | 75 | 100, title, description };
    }
  }

  return null;
}
