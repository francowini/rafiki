"use client";

import { useState } from "react";
import { ThinkForm } from "@/components/features/ThinkForm";
import { ThinkList } from "@/components/features/ThinkList";

export default function ThinksPage() {
  const [refreshKey, setRefreshKey] = useState(0);

  const handleThinkCreated = () => {
    // Trigger refresh of ThinkList
    setRefreshKey((prev) => prev + 1);
  };

  return (
    <div className="space-y-8">
      <div className="max-w-2xl">
        <ThinkForm onSuccess={handleThinkCreated} />
      </div>

      <div>
        <ThinkList refresh={refreshKey} />
      </div>
    </div>
  );
}