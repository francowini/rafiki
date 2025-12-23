export interface ProgressResult {
  percentage: number;
  isInverse: boolean;
  isNegativeProgress: boolean;
  displayText: string;
}

export function calculateProgress(
  begin: number | undefined,
  current: number,
  target: number,
): ProgressResult {
  const startValue = begin ?? 0;
  const isInverse = startValue > target;

  if (isInverse) {
    // Decrease goal (weight loss: 80 -> 70)
    const totalRange = startValue - target;

    // Guard against division by zero (startValue === target)
    if (totalRange === 0) {
      return {
        percentage: 100,
        isInverse: true,
        isNegativeProgress: current > startValue,
        displayText: 'Meta alcanzada',
      };
    }

    const progress = startValue - current;
    const percentage = Math.min(100, Math.max(0, (progress / totalRange) * 100));
    const isNegativeProgress = current > startValue;

    // Cap displayed progress at totalRange when overshooting (current below target)
    const displayProgress = Math.min(Math.abs(progress), totalRange);

    return {
      percentage: Math.round(percentage),
      isInverse: true,
      isNegativeProgress,
      displayText: isNegativeProgress
        ? `+${current - startValue} desde inicio`
        : `${displayProgress} reducido de ${totalRange}`,
    };
  } else {
    // Increase goal (books: 0 -> 12)
    const totalRange = target - startValue;

    // Guard against division by zero (startValue === target)
    if (totalRange === 0) {
      return {
        percentage: 100,
        isInverse: false,
        isNegativeProgress: current < startValue,
        displayText: 'Meta alcanzada',
      };
    }

    const progress = current - startValue;
    const percentage = Math.min(100, Math.max(0, (progress / totalRange) * 100));
    const isNegativeProgress = current < startValue;

    return {
      percentage: Math.round(percentage),
      isInverse: false,
      isNegativeProgress,
      displayText: isNegativeProgress
        ? `${startValue - current} por debajo del inicio`
        : `${progress} de ${totalRange}`,
    };
  }
}

export function getProgressColors(isNegativeProgress: boolean) {
  if (isNegativeProgress) {
    return {
      barColor: 'bg-orange-500',
      textColor: 'text-orange-700',
      bgColor: 'bg-orange-50',
    };
  }
  return {
    barColor: 'bg-blue-500',
    textColor: 'text-blue-700',
    bgColor: 'bg-blue-50',
  };
}
