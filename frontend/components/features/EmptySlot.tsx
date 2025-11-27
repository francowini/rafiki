'use client';

interface EmptySlotProps {
  slotNumber: number;
}

export function EmptySlot({ slotNumber }: EmptySlotProps) {
  return (
    <div className="flex gap-2 items-stretch">
      {/* Spacer for drag handle alignment */}
      <div className="w-9 flex-shrink-0" />

      {/* Empty slot card */}
      <div className="flex-1 h-24 bg-muted/20 border-2 border-dashed border-muted-foreground/20 rounded-lg flex items-center justify-center">
        <span className="text-sm text-muted-foreground">Slot #{slotNumber} - Empty</span>
      </div>
    </div>
  );
}
