# Component Implementation Examples

This document provides complete, production-ready React component examples for the Telegram integration.

---

## 1. Updated Types (`lib/types.ts`)

```typescript
// Add to /Users/francowini/Documents/rafiki/frontend/lib/types.ts

// ============================================================================
// Telegram Integration Types
// ============================================================================

export type MomentSource = "web" | "telegram" | "api";

export interface TelegramStatus {
  linked: boolean;
  userId?: string;          // User's Telegram ID (if linked)
  username?: string;        // Telegram username (if linked)
  linkedAt?: string;        // ISO 8601 timestamp
}

export interface TelegramLinkCode {
  code: string;
  expiresAt: string;        // ISO 8601 timestamp
  deepLink: string;         // t.me link with pre-filled code
}

export interface TelegramUnlinkRequest {
  confirm: boolean;
}

export interface TelegramImport {
  id: string;
  momentId: string;
  status: "success" | "failed" | "partial";
  telegramMessageId: string;
  importedAt: string;       // ISO 8601
  errorMessage?: string;
}

export interface TelegramImportListResponse {
  items: TelegramImport[];
  total: number;
  page: number;
  rowsPerPage: number;
}

export type TelegramErrorType = "parse" | "auth" | "network" | "rate_limit" | "validation";

export interface TelegramError {
  type: TelegramErrorType;
  message: string;
  field?: string;
  retryAfter?: number;      // seconds
}

// Update existing Moment interface
export interface Moment {
  id: string;
  momentDate: string;
  situation: string;
  thoughts: string;
  physicalSymptoms: string;
  behavior: string;
  consequences: string;
  valuesReflection: string;
  intensity: number;
  dateCreated: string;
  dateUpdated: string;

  // New fields for Telegram integration
  source?: MomentSource;
  telegramMessageId?: string;
}
```

---

## 2. Updated API Client (`lib/api.ts`)

```typescript
// Add to /Users/francowini/Documents/rafiki/frontend/lib/api.ts

import {
  TelegramStatus,
  TelegramLinkCode,
  TelegramUnlinkRequest,
  TelegramImportListResponse,
} from "./types";

export const api = {
  // ... existing code (thinks, moments, health)

  telegram: {
    /**
     * Get Telegram link status
     */
    getStatus: async (): Promise<TelegramStatus> => {
      return fetchAPI<TelegramStatus>("/v1/telegram/status");
    },

    /**
     * Generate a new link code
     */
    generateLinkCode: async (): Promise<TelegramLinkCode> => {
      return fetchAPI<TelegramLinkCode>("/v1/telegram/link-code", {
        method: "POST",
      });
    },

    /**
     * Unlink Telegram account
     */
    unlink: async (): Promise<void> => {
      return fetchAPI<void>("/v1/telegram/unlink", {
        method: "POST",
        body: JSON.stringify({ confirm: true } as TelegramUnlinkRequest),
      });
    },

    /**
     * Get import history
     */
    getImportHistory: async (params?: {
      page?: number;
      rows?: number;
    }): Promise<TelegramImportListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.set("page", params.page.toString());
      if (params?.rows) queryParams.set("rows", params.rows.toString());

      const query = queryParams.toString();
      return fetchAPI<TelegramImportListResponse>(
        `/v1/telegram/import-history${query ? `?${query}` : ""}`
      );
    },
  },
};
```

---

## 3. Custom Hook: `usePageVisibility`

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/hooks/usePageVisibility.ts

import { useState, useEffect } from 'react';

/**
 * Hook to detect when the page is visible or hidden.
 * Useful for pausing/resuming polling or refreshing data when user returns.
 */
export function usePageVisibility(): boolean {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    // Initial check
    setIsVisible(document.visibilityState === 'visible');

    const handleVisibilityChange = () => {
      setIsVisible(document.visibilityState === 'visible');
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  return isVisible;
}
```

---

## 4. Custom Hook: `useCountdown`

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/hooks/useCountdown.ts

import { useState, useEffect } from 'react';

interface CountdownResult {
  timeLeft: number;      // seconds remaining
  isExpired: boolean;
  formatted: string;     // "4:32" format
}

/**
 * Hook to manage countdown timer.
 *
 * @param expiresAt - ISO 8601 timestamp of expiration
 * @returns Countdown state with formatted time
 */
export function useCountdown(expiresAt: string | null): CountdownResult {
  const [timeLeft, setTimeLeft] = useState(0);

  useEffect(() => {
    if (!expiresAt) {
      setTimeLeft(0);
      return;
    }

    const calculateTimeLeft = () => {
      const now = new Date().getTime();
      const expiry = new Date(expiresAt).getTime();
      const diff = Math.max(0, Math.floor((expiry - now) / 1000));
      setTimeLeft(diff);
    };

    // Initial calculation
    calculateTimeLeft();

    // Update every second
    const interval = setInterval(calculateTimeLeft, 1000);

    return () => clearInterval(interval);
  }, [expiresAt]);

  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;
  const formatted = `${minutes}:${seconds.toString().padStart(2, '0')}`;

  return {
    timeLeft,
    isExpired: timeLeft === 0,
    formatted,
  };
}
```

---

## 5. Component: `CountdownTimer`

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/components/features/CountdownTimer.tsx

"use client";

import { useCountdown } from "@/hooks/useCountdown";
import { cn } from "@/lib/utils";

interface CountdownTimerProps {
  expiresAt: string;
  className?: string;
  onExpire?: () => void;
}

export function CountdownTimer({ expiresAt, className, onExpire }: CountdownTimerProps) {
  const { formatted, isExpired, timeLeft } = useCountdown(expiresAt);

  // Call onExpire callback when timer hits zero
  React.useEffect(() => {
    if (isExpired && onExpire) {
      onExpire();
    }
  }, [isExpired, onExpire]);

  return (
    <span
      className={cn(
        "font-mono",
        timeLeft < 60 && "text-destructive font-bold", // Red when < 1 minute
        className
      )}
      aria-live="polite"
      aria-atomic="true"
    >
      {formatted}
    </span>
  );
}
```

---

## 6. Component: `LinkCodeDisplay`

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/components/features/LinkCodeDisplay.tsx

"use client";

import { useState } from "react";
import QRCode from "qrcode.react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CountdownTimer } from "./CountdownTimer";
import { MessageCircle, Copy, Check } from "lucide-react";
import { TelegramLinkCode } from "@/lib/types";

interface LinkCodeDisplayProps {
  linkCode: TelegramLinkCode;
  onCancel: () => void;
  onExpire: () => void;
}

export function LinkCodeDisplay({ linkCode, onCancel, onExpire }: LinkCodeDisplayProps) {
  const [copied, setCopied] = useState(false);

  const handleCopyCode = async () => {
    await navigator.clipboard.writeText(linkCode.code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Link Your Telegram</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Code Display */}
        <div className="bg-muted p-8 rounded-lg text-center">
          <p className="text-sm text-muted-foreground mb-3">Your link code</p>
          <div className="flex items-center justify-center gap-2">
            <p className="text-5xl font-mono font-bold tracking-wider select-all">
              {linkCode.code}
            </p>
            <Button
              variant="ghost"
              size="icon"
              onClick={handleCopyCode}
              className="ml-2"
            >
              {copied ? (
                <Check className="h-5 w-5 text-green-600" />
              ) : (
                <Copy className="h-5 w-5" />
              )}
            </Button>
          </div>
          <p className="text-sm text-muted-foreground mt-4">
            Expires in{" "}
            <CountdownTimer expiresAt={linkCode.expiresAt} onExpire={onExpire} />
          </p>
        </div>

        {/* Mobile: Deep Link Button */}
        <div className="sm:hidden">
          <Button
            onClick={() => window.open(linkCode.deepLink, "_blank")}
            className="w-full h-12 text-base"
            size="lg"
          >
            <MessageCircle className="mr-2 h-5 w-5" />
            Open Telegram
          </Button>
          <p className="text-xs text-center text-muted-foreground mt-2">
            This will open the Telegram app
          </p>
        </div>

        {/* Desktop: QR Code */}
        <div className="hidden sm:flex flex-col items-center gap-4">
          <QRCode value={linkCode.deepLink} size={200} level="M" />
          <p className="text-sm text-muted-foreground text-center max-w-xs">
            Scan this QR code with your phone's camera to open Telegram
          </p>
          <p className="text-xs text-muted-foreground">
            Or click:{" "}
            <a
              href={linkCode.deepLink}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary underline"
            >
              Open Telegram Desktop
            </a>
          </p>
        </div>

        {/* Instructions */}
        <Alert className="bg-blue-50 border-blue-200">
          <AlertDescription>
            <p className="font-medium mb-2">How to link:</p>
            <ol className="text-sm space-y-1 list-decimal list-inside">
              <li>Open Telegram using the button or QR code above</li>
              <li>The bot will ask for your code - send it</li>
              <li>Wait for confirmation (this page will update automatically)</li>
            </ol>
          </AlertDescription>
        </Alert>

        {/* Actions */}
        <div className="flex gap-3">
          <Button variant="outline" onClick={onCancel} className="flex-1">
            Cancel
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## 7. Component: `TelegramStatusCard`

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/components/features/TelegramStatusCard.tsx

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { TelegramStatus } from "@/lib/types";
import { CheckCircle, MessageCircle, AlertCircle } from "lucide-react";
import { formatDistanceToNow } from "date-fns";

interface TelegramStatusCardProps {
  status: TelegramStatus;
  onLink: () => void;
  onUnlink: () => void;
}

export function TelegramStatusCard({ status, onLink, onUnlink }: TelegramStatusCardProps) {
  if (status.linked) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-xl">Telegram Status</CardTitle>
            <Badge className="bg-green-100 text-green-800 border-green-200">
              Linked
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-start gap-3">
            <CheckCircle className="h-8 w-8 text-green-600 mt-1 flex-shrink-0" />
            <div className="flex-1">
              <p className="font-medium">
                Connected to @{status.username || "unknown"}
              </p>
              <p className="text-sm text-muted-foreground">
                {status.linkedAt
                  ? `Linked ${formatDistanceToNow(new Date(status.linkedAt))} ago`
                  : "Linked recently"}
              </p>
            </div>
          </div>

          <div className="bg-blue-50 p-4 rounded-lg">
            <p className="text-sm text-blue-900">
              You can now create moments by chatting with{" "}
              <span className="font-medium">@rafiki_moments_bot</span> on Telegram.
            </p>
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={onUnlink}
            className="w-full text-destructive hover:text-destructive"
          >
            Unlink Account
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl">Telegram Status</CardTitle>
          <Badge variant="outline">Not Linked</Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="text-center py-6">
          <MessageCircle className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground mb-1">
            Link your Telegram account to create moments on the go
          </p>
          <p className="text-sm text-muted-foreground">
            Use the Telegram bot for a guided conversation
          </p>
        </div>

        <Button onClick={onLink} className="w-full" size="lg">
          Link Telegram Account
        </Button>
      </CardContent>
    </Card>
  );
}
```

---

## 8. Component: `TelegramImportHistory`

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/components/features/TelegramImportHistory.tsx

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { api } from "@/lib/api";
import { TelegramImport } from "@/lib/types";
import { formatDistanceToNow } from "date-fns";
import { CheckCircle, XCircle, ExternalLink, AlertCircle } from "lucide-react";

interface TelegramImportHistoryProps {
  limit?: number;
}

export function TelegramImportHistory({ limit = 10 }: TelegramImportHistoryProps) {
  const [imports, setImports] = useState<TelegramImport[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadImports();
  }, []);

  const loadImports = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.telegram.getImportHistory({
        page: 1,
        rows: limit,
      });
      setImports(response.items);
    } catch (err: any) {
      setError(err.message || "Failed to load import history");
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Recent Imports</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-20 w-full" />
          ))}
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Recent Imports</CardTitle>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
          <Button onClick={loadImports} variant="outline" className="mt-4 w-full">
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (imports.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Recent Imports</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <p className="text-sm text-muted-foreground">
              No imports yet. Start a conversation with the bot on Telegram!
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Recent Imports ({imports.length})</CardTitle>
          <Button variant="ghost" size="sm" onClick={loadImports}>
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {imports.map((importItem) => (
          <ImportItem key={importItem.id} import={importItem} />
        ))}
      </CardContent>
    </Card>
  );
}

function ImportItem({ import: importItem }: { import: TelegramImport }) {
  const isSuccess = importItem.status === "success";
  const isFailed = importItem.status === "failed";

  return (
    <div
      className={cn(
        "border rounded-lg p-4",
        isFailed && "border-destructive bg-destructive/5"
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3 flex-1">
          {isSuccess ? (
            <CheckCircle className="h-5 w-5 text-green-600 mt-0.5 flex-shrink-0" />
          ) : (
            <XCircle className="h-5 w-5 text-destructive mt-0.5 flex-shrink-0" />
          )}

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <Badge
                variant={isSuccess ? "default" : "destructive"}
                className="text-xs"
              >
                {importItem.status}
              </Badge>
              <span className="text-xs text-muted-foreground">
                {formatDistanceToNow(new Date(importItem.importedAt))} ago
              </span>
            </div>

            {isFailed && importItem.errorMessage && (
              <p className="text-sm text-destructive mt-2">
                {importItem.errorMessage}
              </p>
            )}
          </div>
        </div>

        {isSuccess && (
          <Link href={`/momentos/${importItem.momentId}`}>
            <Button variant="ghost" size="sm" className="flex-shrink-0">
              View
              <ExternalLink className="ml-1 h-3 w-3" />
            </Button>
          </Link>
        )}
      </div>
    </div>
  );
}

// Helper function (already exists in lib/utils.ts, but shown for completeness)
function cn(...inputs: any[]) {
  // Implementation from lib/utils.ts
  return inputs.filter(Boolean).join(" ");
}
```

---

## 9. Main Settings Page Component

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/app/(dashboard)/settings/page.tsx

"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api";
import { TelegramStatus, TelegramLinkCode } from "@/lib/types";
import { TelegramStatusCard } from "@/components/features/TelegramStatusCard";
import { LinkCodeDisplay } from "@/components/features/LinkCodeDisplay";
import { TelegramImportHistory } from "@/components/features/TelegramImportHistory";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle } from "lucide-react";

type LinkingState = "idle" | "generating" | "waiting" | "linked" | "error";

export default function SettingsPage() {
  const [status, setStatus] = useState<TelegramStatus | null>(null);
  const [linkCode, setLinkCode] = useState<TelegramLinkCode | null>(null);
  const [linkingState, setLinkingState] = useState<LinkingState>("idle");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showUnlinkDialog, setShowUnlinkDialog] = useState(false);

  // Load initial status
  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const statusData = await api.telegram.getStatus();
      setStatus(statusData);
      setLinkingState(statusData.linked ? "linked" : "idle");
    } catch (err: any) {
      setError(err.message || "Failed to load Telegram status");
      setLinkingState("error");
    } finally {
      setIsLoading(false);
    }
  };

  // Start linking process
  const handleStartLink = async () => {
    setLinkingState("generating");
    setError(null);

    try {
      const code = await api.telegram.generateLinkCode();
      setLinkCode(code);
      setLinkingState("waiting");

      // Start polling for link confirmation
      startPollingStatus();
    } catch (err: any) {
      setError(err.message || "Failed to generate link code");
      setLinkingState("error");
    }
  };

  // Poll status while waiting for user to link in Telegram
  const startPollingStatus = useCallback(() => {
    const pollInterval = setInterval(async () => {
      try {
        const statusData = await api.telegram.getStatus();

        if (statusData.linked) {
          setStatus(statusData);
          setLinkingState("linked");
          setLinkCode(null);
          clearInterval(pollInterval);
        }
      } catch (err) {
        // Silently fail - user can retry
        console.error("Poll error:", err);
      }
    }, 3000); // Poll every 3 seconds

    // Clean up after 5 minutes (code expiry)
    setTimeout(() => {
      clearInterval(pollInterval);
    }, 5 * 60 * 1000);

    return pollInterval;
  }, []);

  // Handle code expiry
  const handleCodeExpire = () => {
    setLinkCode(null);
    setLinkingState("idle");
    setError("Link code expired. Please generate a new one.");
  };

  // Handle cancel
  const handleCancelLink = () => {
    setLinkCode(null);
    setLinkingState("idle");
  };

  // Handle unlink
  const handleUnlink = async () => {
    try {
      await api.telegram.unlink();
      setStatus({ linked: false });
      setLinkingState("idle");
      setShowUnlinkDialog(false);
    } catch (err: any) {
      setError(err.message || "Failed to unlink account");
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto py-8 px-4 max-w-4xl">
        <h1 className="text-3xl font-bold tracking-tight mb-8">Settings</h1>
        <div className="space-y-6">
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 px-4 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground mt-1">
          Manage your Telegram integration
        </p>
      </div>

      <div className="space-y-6">
        {/* Error Alert */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Status Card or Link Code Display */}
        {linkingState === "waiting" && linkCode ? (
          <LinkCodeDisplay
            linkCode={linkCode}
            onCancel={handleCancelLink}
            onExpire={handleCodeExpire}
          />
        ) : (
          status && (
            <TelegramStatusCard
              status={status}
              onLink={handleStartLink}
              onUnlink={() => setShowUnlinkDialog(true)}
            />
          )
        )}

        {/* Import History (only shown when linked) */}
        {status?.linked && <TelegramImportHistory limit={10} />}
      </div>

      {/* Unlink Confirmation Dialog */}
      <AlertDialog open={showUnlinkDialog} onOpenChange={setShowUnlinkDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Unlink Telegram Account?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to unlink your Telegram account (@
              {status?.username || "unknown"})?
              <ul className="list-disc list-inside mt-4 space-y-1 text-sm">
                <li>You won't be able to create moments via Telegram</li>
                <li>Your existing moments will NOT be deleted</li>
                <li>You can re-link anytime</li>
              </ul>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleUnlink}
              className="bg-destructive hover:bg-destructive/90"
            >
              Unlink
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
```

---

## 10. Updated MomentCard with Source Badge

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/components/features/MomentCard.tsx
// UPDATE existing component

"use client";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Moment } from "@/lib/types";
import { formatDistanceToNow, format } from "date-fns";
import { MessageCircle, Globe, MoreVertical, Edit, Trash } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface MomentCardProps {
  moment: Moment;
  onClick?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function MomentCard({ moment, onClick, onEdit, onDelete }: MomentCardProps) {
  // Determine source icon and label
  const source = moment.source || "web"; // Default to "web" for backward compatibility
  const sourceIcon = source === "telegram" ? MessageCircle : Globe;
  const sourceLabel = source === "telegram" ? "Telegram" : "Web";
  const sourceBadgeClass =
    source === "telegram"
      ? "bg-blue-100 text-blue-800 border-blue-200"
      : "bg-gray-100 text-gray-700 border-gray-200";

  return (
    <Card className="cursor-pointer hover:shadow-md transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between mb-2">
          <Badge variant="outline" className={`gap-1 ${sourceBadgeClass}`}>
            {React.createElement(sourceIcon, { className: "h-3 w-3" })}
            {sourceLabel}
          </Badge>

          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">
              {format(new Date(moment.momentDate), "MMM d, yyyy")}
            </span>

            {/* Actions Menu */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={onEdit}>
                  <Edit className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={onDelete}
                  className="text-destructive"
                >
                  <Trash className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        <h3
          className="font-semibold text-lg line-clamp-2 cursor-pointer"
          onClick={onClick}
        >
          {moment.situation}
        </h3>
      </CardHeader>

      <CardContent onClick={onClick}>
        <p className="text-sm text-muted-foreground line-clamp-3 mb-4">
          {moment.thoughts}
        </p>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium">Intensity:</span>
            <Badge
              variant="outline"
              className={
                moment.intensity >= 7
                  ? "bg-red-100 text-red-800 border-red-200"
                  : moment.intensity >= 4
                  ? "bg-yellow-100 text-yellow-800 border-yellow-200"
                  : "bg-green-100 text-green-800 border-green-200"
              }
            >
              {moment.intensity}/10
            </Badge>
          </div>

          <span className="text-xs text-muted-foreground">
            {formatDistanceToNow(new Date(moment.dateCreated))} ago
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## 11. Updated Header with Settings Link

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/components/layout/Header.tsx
// UPDATE existing component

'use client';

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth-context";
import { UserMenu } from "@/components/auth/UserMenu";
import { Settings } from "lucide-react";

export function Header() {
  const { isAuthenticated } = useAuth();

  return (
    <header className="border-b bg-white">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-2xl font-bold text-gray-900">
            Rafiki
          </Link>
          {isAuthenticated && (
            <nav className="flex gap-4">
              <Link href="/thinks">
                <Button variant="ghost">Thinks</Button>
              </Link>
              <Link href="/settings">
                <Button variant="ghost">
                  <Settings className="h-4 w-4 mr-2" />
                  Settings
                </Button>
              </Link>
            </nav>
          )}
        </div>
        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <UserMenu />
          ) : (
            <Link href="/login">
              <Button>Login</Button>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
```

---

## 12. Package.json Dependencies Update

```json
// File: /Users/francowini/Documents/rafiki/frontend/package.json
// ADD these dependencies

{
  "dependencies": {
    // ... existing dependencies
    "qrcode.react": "^3.1.0",
    "date-fns": "^3.0.0"
  },
  "devDependencies": {
    // ... existing devDependencies
    "@types/qrcode.react": "^1.0.5"
  }
}
```

**Installation command**:
```bash
cd /Users/francowini/Documents/rafiki/frontend
npm install qrcode.react date-fns
npm install -D @types/qrcode.react
```

---

## 13. Usage Example: Complete Flow

```typescript
// Example: How a user experiences the flow

// 1. User navigates to /settings
// → SettingsPage component loads
// → Calls api.telegram.getStatus()
// → Shows TelegramStatusCard with "Not Linked" state

// 2. User clicks "Link Telegram Account"
// → SettingsPage.handleStartLink() called
// → Calls api.telegram.generateLinkCode()
// → Receives { code: "AB12CD", expiresAt: "...", deepLink: "..." }
// → Shows LinkCodeDisplay component

// 3. User sees countdown timer (4:32 remaining)
// → CountdownTimer component renders
// → Updates every second
// → Warns when < 1 minute (red text)

// 4. User clicks "Open Telegram" (mobile) or scans QR (desktop)
// → Opens t.me/rafiki_moments_bot?start=AB12CD
// → User sends /start AB12CD in Telegram

// 5. Backend validates and links account
// → SettingsPage is polling api.telegram.getStatus() every 3 seconds
// → Receives { linked: true, username: "john_doe", ... }
// → Updates UI to show "Linked" state
// → Stops polling

// 6. User can now see import history
// → TelegramImportHistory component loads
// → Calls api.telegram.getImportHistory()
// → Shows list of recent imports

// 7. User creates moment via Telegram
// → Backend processes and saves moment
// → User opens /momentos page
// → Page Visibility API triggers refresh
// → New moment appears with Telegram badge
```

---

## 14. Testing Utilities

```typescript
// File: /Users/francowini/Documents/rafiki/frontend/__tests__/telegram.test.tsx
// Example test structure (using Jest + React Testing Library)

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { TelegramStatusCard } from '@/components/features/TelegramStatusCard';

const server = setupServer(
  rest.get('/v1/telegram/status', (req, res, ctx) => {
    return res(ctx.json({ linked: false }));
  }),
  rest.post('/v1/telegram/link-code', (req, res, ctx) => {
    return res(
      ctx.json({
        code: 'AB12CD',
        expiresAt: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
        deepLink: 'https://t.me/rafiki_moments_bot?start=AB12CD',
      })
    );
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('TelegramStatusCard', () => {
  it('shows unlinked state correctly', () => {
    const status = { linked: false };
    render(<TelegramStatusCard status={status} onLink={jest.fn()} onUnlink={jest.fn()} />);

    expect(screen.getByText('Not Linked')).toBeInTheDocument();
    expect(screen.getByText('Link Telegram Account')).toBeInTheDocument();
  });

  it('shows linked state with username', () => {
    const status = {
      linked: true,
      username: 'john_doe',
      linkedAt: new Date().toISOString(),
    };

    render(<TelegramStatusCard status={status} onLink={jest.fn()} onUnlink={jest.fn()} />);

    expect(screen.getByText('Linked')).toBeInTheDocument();
    expect(screen.getByText(/Connected to @john_doe/)).toBeInTheDocument();
  });

  it('calls onLink when link button is clicked', async () => {
    const onLink = jest.fn();
    const status = { linked: false };

    render(<TelegramStatusCard status={status} onLink={onLink} onUnlink={jest.fn()} />);

    const linkButton = screen.getByText('Link Telegram Account');
    await userEvent.click(linkButton);

    expect(onLink).toHaveBeenCalledTimes(1);
  });
});
```

---

**Document Version**: 1.0
**Last Updated**: 2025-11-17

**Note**: All components follow the existing Rafiki frontend patterns:
- Radix UI + shadcn/ui components
- Tailwind CSS for styling
- React Hook Form + Zod for forms (where applicable)
- Date-fns for date formatting
- Next.js 16 App Router conventions
