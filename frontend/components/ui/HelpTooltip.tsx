import { HelpCircle } from 'lucide-react';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

interface HelpTooltipProps {
  content: string;
  className?: string;
}

export function HelpTooltip({ content, className }: HelpTooltipProps) {
  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>
          {/* F5 Compliance: Use <button> for keyboard accessibility */}
          <button
            type="button"
            className={`inline-flex items-center justify-center rounded-sm focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 ${className || ''}`}
            aria-label="Ayuda"
          >
            <HelpCircle className="h-4 w-4 text-gray-400 hover:text-gray-600 transition-colors" />
          </button>
        </TooltipTrigger>
        <TooltipContent className="max-w-xs">
          <p className="text-sm">{content}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
