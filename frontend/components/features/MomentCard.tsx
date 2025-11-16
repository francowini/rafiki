"use client";

import { Moment } from "@/lib/types";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Trash2, Edit } from "lucide-react";

interface MomentCardProps {
  moment: Moment;
  onClick?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function MomentCard({ moment, onClick, onEdit, onDelete }: MomentCardProps) {
  const momentDate = new Date(moment.momentDate);
  const dateStr = momentDate.toLocaleDateString("en-US", {
    day: "numeric",
    month: "long",
    year: "numeric",
  });
  const timeStr = momentDate.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
  });

  // Truncate situation for preview
  const situationPreview = moment.situation.length > 100
    ? moment.situation.slice(0, 100) + "..."
    : moment.situation;

  // Intensity color
  const getIntensityColor = (intensity: number) => {
    if (intensity >= 8) return "bg-red-100 text-red-800 border-red-300";
    if (intensity >= 5) return "bg-yellow-100 text-yellow-800 border-yellow-300";
    return "bg-green-100 text-green-800 border-green-300";
  };

  return (
    <Card
      className="hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="text-base font-medium">
              {dateStr}
            </CardTitle>
            <CardDescription className="text-sm">
              {timeStr}
            </CardDescription>
          </div>
          <Badge
            variant="outline"
            className={getIntensityColor(moment.intensity)}
          >
            {moment.intensity}/10
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm text-muted-foreground line-clamp-3">
          {situationPreview}
        </p>

        <div className="flex gap-2 pt-2" onClick={(e) => e.stopPropagation()}>
          {onEdit && (
            <Button
              size="sm"
              variant="outline"
              onClick={onEdit}
              className="flex-1"
            >
              <Edit className="h-4 w-4 mr-1" />
              Edit
            </Button>
          )}
          {onDelete && (
            <Button
              size="sm"
              variant="outline"
              onClick={onDelete}
              className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
