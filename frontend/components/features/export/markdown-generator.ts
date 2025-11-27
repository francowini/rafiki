import { ExportItem, ThinkCategory } from '@/lib/types';
import { formatDateForMarkdown, formatDateTimeForMarkdown } from '@/lib/date-utils';

export function generateMarkdown(items: ExportItem[], startDate: Date, endDate: Date): string {
  const lines: string[] = [];

  // Header
  lines.push('# Weekly Journal Export');
  lines.push('');
  lines.push(`**Period:** ${formatDateForMarkdown(startDate)} - ${formatDateForMarkdown(endDate)}`);
  lines.push('');
  lines.push(`**Total entries:** ${items.length}`);
  lines.push('');
  lines.push('---');
  lines.push('');

  // Separate moments and thinks
  const moments = items.filter((item) => item.itemType === 'moment');
  const thinks = items.filter((item) => item.itemType === 'think');

  // Moments section
  if (moments.length > 0) {
    lines.push('## Moments');
    lines.push('');
    lines.push(
      `Recorded difficult moments for reflection and pattern recognition. Total: ${moments.length}`,
    );
    lines.push('');

    const sortedMoments = [...moments].sort((a, b) => {
      const dateA = new Date(a.itemDate);
      const dateB = new Date(b.itemDate);
      return dateB.getTime() - dateA.getTime();
    });

    sortedMoments.forEach((moment, index) => {
      lines.push(...formatMoment(moment, index + 1));
    });
  }

  // Thinks section
  if (thinks.length > 0) {
    lines.push('## Thoughts & Notes');
    lines.push('');
    lines.push(`Captured thoughts, ideas, and reflections. Total: ${thinks.length}`);
    lines.push('');

    const thinksByCategory = groupThinksByCategory(thinks);

    Object.entries(thinksByCategory).forEach(([category, categoryThinks]) => {
      lines.push(`### ${formatCategory(category as ThinkCategory)}`);
      lines.push('');

      const sortedThinks = [...categoryThinks].sort((a, b) => {
        const dateA = new Date(a.dateCreated);
        const dateB = new Date(b.dateCreated);
        return dateB.getTime() - dateA.getTime();
      });

      sortedThinks.forEach((think, index) => {
        lines.push(...formatThink(think, index + 1));
      });
    });
  }

  // Footer
  lines.push('---');
  lines.push('');
  lines.push(`*Exported on ${formatDateTimeForMarkdown(new Date())}*`);
  lines.push('');

  return lines.join('\n');
}

function formatMoment(moment: ExportItem, number: number): string[] {
  const lines: string[] = [];
  const momentDate = new Date(moment.itemDate);

  lines.push(
    `### ${number}. ${formatDateTimeForMarkdown(momentDate)} - Intensity: ${moment.intensity}/10`,
  );
  lines.push('');

  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(moment.situation);
    lines.push('');
  }

  if (moment.thoughts) {
    lines.push('**Thoughts:**');
    lines.push(moment.thoughts);
    lines.push('');
  }

  if (moment.physicalSymptoms) {
    lines.push('**Physical Symptoms:**');
    lines.push(moment.physicalSymptoms);
    lines.push('');
  }

  if (moment.behavior) {
    lines.push('**Behavior:**');
    lines.push(moment.behavior);
    lines.push('');
  }

  if (moment.consequences) {
    lines.push('**Consequences:**');
    lines.push(moment.consequences);
    lines.push('');
  }

  if (moment.valuesReflection) {
    lines.push('**Values Reflection:**');
    lines.push(moment.valuesReflection);
    lines.push('');
  }

  return lines;
}

function formatThink(think: ExportItem, number: number): string[] {
  const lines: string[] = [];
  const thinkDate = new Date(think.dateCreated);

  lines.push(`#### ${number}. ${formatDateTimeForMarkdown(thinkDate)}`);
  lines.push('');
  lines.push(think.content || '');
  lines.push('');

  return lines;
}

function groupThinksByCategory(thinks: ExportItem[]): Record<string, ExportItem[]> {
  return thinks.reduce(
    (acc, think) => {
      const category = think.category || 'personal';
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(think);
      return acc;
    },
    {} as Record<string, ExportItem[]>,
  );
}

function formatCategory(category: ThinkCategory): string {
  const categoryMap: Record<ThinkCategory, string> = {
    personal: 'Personal',
    work: 'Work',
    ideas: 'Ideas',
    learning: 'Learning',
    reflection: 'Reflection',
  };
  return categoryMap[category] || category;
}
