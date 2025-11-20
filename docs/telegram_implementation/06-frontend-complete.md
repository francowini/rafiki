# Frontend Implementation - Complete Guide

**Task Category**: Frontend
**Estimated Time**: 17-23 hours total
**Prerequisites**: Backend API endpoints completed ([07-backend-api.md](./07-backend-api.md))
**Dependencies**: None

---

## Overview

This document combines all frontend implementation tasks:
1. Types & API Client (3-4h)
2. Generic Telegram Components (8-10h)
3. Settings Page Integration (4-6h)
4. Moment Card Updates (2-3h)

---

## Part 1: Types & API Client (3-4h)

### Task 1.1: Add Telegram Types

**File**: `frontend/lib/types.ts`

Add to end of file:

```typescript
// ============================================================================
// Telegram Integration Types
// ============================================================================

export type TelegramSourceType = "web" | "telegram";

// Connection status
export interface TelegramConnectionStatus {
  connected: boolean;
  telegramUserId?: string;
  telegramUsername?: string;
  connectedAt?: string;
}

// Link code for account linking
export interface TelegramLinkCode {
  linkCode: string;
  expiresAt: string;
  expiresInSeconds: number;
  botUsername: string;
}

// Extend existing Moment type
export interface Moment {
  // ... existing fields
  source?: TelegramSourceType; // NEW - optional for backward compatibility
  sourceMetadata?: {            // NEW
    messageId?: number;
    conversationDuration?: number;
  };
}
```

### Task 1.2: Update API Client

**File**: `frontend/lib/api.ts`

Add telegram namespace:

```typescript
export const api = {
  // ... existing methods (thinks, moments, users, etc.)

  telegram: {
    // Get connection status
    getStatus: async (): Promise<TelegramConnectionStatus> => {
      return fetchAPI<TelegramConnectionStatus>("/v1/telegram/status");
    },

    // Generate link code
    generateLinkCode: async (): Promise<TelegramLinkCode> => {
      return fetchAPI<TelegramLinkCode>("/v1/telegram/link", {
        method: "POST",
      });
    },

    // Disconnect account
    disconnect: async (): Promise<void> => {
      return fetchAPI<void>("/v1/telegram/disconnect", {
        method: "DELETE",
      });
    },
  },
};
```

**Checklist Part 1**:
- [ ] Add Telegram types to `types.ts`
- [ ] Extend `Moment` interface with source fields
- [ ] Add `telegram` namespace to `api.ts`
- [ ] Test API calls with Postman/curl

---

## Part 2: Generic Telegram Components (8-10h)

### Task 2.1: Create Telegram Context

**File**: `frontend/lib/telegram-context.tsx`

```typescript
"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { api } from "./api";
import type { TelegramConnectionStatus } from "./types";

interface TelegramContextValue {
  status: TelegramConnectionStatus | null;
  isConnected: boolean;
  isLoading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

const TelegramContext = createContext<TelegramContextValue | undefined>(undefined);

export function TelegramProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<TelegramConnectionStatus | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await api.telegram.getStatus();
      setStatus(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch status");
      console.error("Error fetching Telegram status:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  return (
    <TelegramContext.Provider
      value={{
        status,
        isConnected: status?.connected ?? false,
        isLoading,
        error,
        refresh,
      }}
    >
      {children}
    </TelegramContext.Provider>
  );
}

export function useTelegram() {
  const context = useContext(TelegramContext);
  if (!context) {
    throw new Error("useTelegram must be used within TelegramProvider");
  }
  return context;
}
```

### Task 2.2: Source Badge Component

**File**: `frontend/components/ui/source-badge.tsx`

```typescript
import { Badge } from "@/components/ui/badge";
import { MessageCircle } from "lucide-react";
import type { TelegramSourceType } from "@/lib/types";

interface SourceBadgeProps {
  source: TelegramSourceType;
  variant?: "full" | "compact" | "icon-only";
}

export function SourceBadge({ source, variant = "full" }: SourceBadgeProps) {
  if (source === "web") {
    return null; // Don't show badge for web-created moments
  }

  return (
    <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
      <MessageCircle className="h-3 w-3 mr-1" />
      {variant !== "icon-only" && <span>Telegram</span>}
    </Badge>
  );
}
```

### Task 2.3: Telegram Connection Card

**File**: `frontend/components/integrations/telegram/TelegramConnectionCard.tsx`

```typescript
"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useTelegram } from "@/lib/telegram-context";
import { TelegramLinkingFlow } from "./TelegramLinkingFlow";
import { formatDistanceToNow } from "date-fns";

export function TelegramConnectionCard() {
  const { status, isConnected, refresh } = useTelegram();
  const [showLinking, setShowLinking] = useState(false);

  if (showLinking) {
    return (
      <TelegramLinkingFlow
        onComplete={() => {
          setShowLinking(false);
          refresh();
        }}
        onCancel={() => setShowLinking(false)}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Telegram Bot</CardTitle>
        <CardDescription>
          Record moments directly from Telegram
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isConnected ? (
          <ConnectedView status={status!} onDisconnect={refresh} />
        ) : (
          <DisconnectedView onConnect={() => setShowLinking(true)} />
        )}
      </CardContent>
    </Card>
  );
}

function ConnectedView({ status, onDisconnect }: any) {
  const [isDisconnecting, setIsDisconnecting] = useState(false);

  const handleDisconnect = async () => {
    if (!confirm("Are you sure you want to disconnect your Telegram account?")) {
      return;
    }

    try {
      setIsDisconnecting(true);
      await api.telegram.disconnect();
      onDisconnect();
    } catch (err) {
      console.error("Error disconnecting:", err);
      alert("Failed to disconnect. Please try again.");
    } finally {
      setIsDisconnecting(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <div className="h-3 w-3 bg-green-500 rounded-full" />
        <span className="font-medium">Connected as @{status.telegramUsername}</span>
      </div>
      {status.connectedAt && (
        <p className="text-sm text-muted-foreground">
          Connected {formatDistanceToNow(new Date(status.connectedAt), { addSuffix: true })}
        </p>
      )}
      <div className="space-y-2">
        <p className="text-sm">
          You can now create moments by messaging <strong>@rafiki_bot</strong> on Telegram.
        </p>
        <p className="text-sm text-muted-foreground">
          Send <code className="bg-muted px-1 py-0.5 rounded">/moment</code> to start.
        </p>
      </div>
      <Button
        variant="outline"
        onClick={handleDisconnect}
        disabled={isDisconnecting}
      >
        {isDisconnecting ? "Disconnecting..." : "Disconnect"}
      </Button>
    </div>
  );
}

function DisconnectedView({ onConnect }: { onConnect: () => void }) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Connect your Telegram account to create moments on-the-go without opening the web app.
      </p>
      <Button onClick={onConnect}>
        Setup Telegram Integration
      </Button>
    </div>
  );
}
```

### Task 2.4: Telegram Linking Flow

**File**: `frontend/components/integrations/telegram/TelegramLinkingFlow.tsx`

```typescript
"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import type { TelegramLinkCode } from "@/lib/types";
import QRCode from "qrcode.react"; // npm install qrcode.react

interface TelegramLinkingFlowProps {
  onComplete: () => void;
  onCancel: () => void;
}

export function TelegramLinkingFlow({ onComplete, onCancel }: TelegramLinkingFlowProps) {
  const [linkCode, setLinkCode] = useState<TelegramLinkCode | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeLeft, setTimeLeft] = useState(300); // 5 minutes in seconds

  useEffect(() => {
    generateCode();
  }, []);

  useEffect(() => {
    if (!linkCode) return;

    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          setError("Link code expired. Please try again.");
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [linkCode]);

  const generateCode = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const code = await api.telegram.generateLinkCode();
      setLinkCode(code);
      setTimeLeft(300); // Reset timer
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate code");
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <div>Generating link code...</div>;
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Error</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-destructive mb-4">{error}</p>
          <div className="flex gap-2">
            <Button onClick={generateCode}>Try Again</Button>
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;
  const deepLink = `https://t.me/${linkCode?.botUsername}?start=${linkCode?.linkCode}`;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Connect Telegram</CardTitle>
        <CardDescription>
          Scan the QR code or use the link code below
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* QR Code (Desktop) */}
        <div className="hidden md:flex flex-col items-center gap-4">
          <p className="text-sm font-medium">1. Scan this QR code with your phone</p>
          <QRCode value={deepLink} size={200} />
        </div>

        {/* Deep Link (Mobile) */}
        <div className="md:hidden flex flex-col gap-4">
          <p className="text-sm font-medium">1. Tap to open Telegram</p>
          <Button asChild className="w-full">
            <a href={deepLink}>Open @{linkCode?.botUsername}</a>
          </Button>
        </div>

        {/* Manual Instructions */}
        <div className="space-y-2">
          <p className="text-sm font-medium">Or manually:</p>
          <ol className="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
            <li>Open Telegram and search for <strong>@{linkCode?.botUsername}</strong></li>
            <li>
              Send this command:{" "}
              <code className="bg-muted px-2 py-1 rounded">/link {linkCode?.linkCode}</code>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigator.clipboard.writeText(`/link ${linkCode?.linkCode}`)}
              >
                Copy
              </Button>
            </li>
          </ol>
        </div>

        {/* Timer */}
        <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
          <span className="text-sm">Expires in:</span>
          <span className="text-sm font-mono font-medium">
            {minutes}:{seconds.toString().padStart(2, "0")}
          </span>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          <Button onClick={onComplete} className="flex-1">
            I've Completed Setup
          </Button>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
```

**Checklist Part 2**:
- [ ] Create `telegram-context.tsx` with provider
- [ ] Create `source-badge.tsx` component
- [ ] Create `TelegramConnectionCard.tsx`
- [ ] Create `TelegramLinkingFlow.tsx`
- [ ] Install QR code library: `npm install qrcode.react`
- [ ] Test components in isolation

---

## Part 3: Settings Page Integration (4-6h)

### Task 3.1: Wrap App with Telegram Provider

**File**: `frontend/app/layout.tsx`

```typescript
import { TelegramProvider } from "@/lib/telegram-context";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <TelegramProvider>
            {children}
          </TelegramProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
```

### Task 3.2: Create Settings Page

**File**: `frontend/app/(dashboard)/settings/page.tsx`

```typescript
import { TelegramConnectionCard } from "@/components/integrations/telegram/TelegramConnectionCard";

export default function SettingsPage() {
  return (
    <div className="container max-w-4xl mx-auto py-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground mt-1">
          Manage your account and integrations
        </p>
      </div>

      {/* Telegram Integration */}
      <section>
        <h2 className="text-2xl font-semibold mb-4">Integrations</h2>
        <TelegramConnectionCard />
      </section>

      {/* Future: Other settings sections */}
    </div>
  );
}
```

### Task 3.3: Add Settings Link to Navigation

**File**: `frontend/components/layout/Navbar.tsx` (or wherever user menu is)

```typescript
<DropdownMenuItem asChild>
  <Link href="/settings">Settings</Link>
</DropdownMenuItem>
```

**Checklist Part 3**:
- [ ] Wrap app with `TelegramProvider` in layout
- [ ] Create `/settings` page
- [ ] Add Telegram section to settings
- [ ] Add settings link to navigation menu
- [ ] Test navigation and page load

---

## Part 4: Moment Card Updates (2-3h)

### Task 4.1: Update Moment Card

**File**: `frontend/components/features/MomentCard.tsx`

```typescript
import { SourceBadge } from "@/components/ui/source-badge";

export function MomentCard({ moment }: { moment: Moment }) {
  return (
    <Card
      className={cn(
        "hover:shadow-md transition-shadow",
        moment.source === "telegram" && "border-l-4 border-l-blue-400"
      )}
    >
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle>{formatDate(moment.momentDate)}</CardTitle>
            <CardDescription>{formatTime(moment.momentDate)}</CardDescription>
          </div>
          <div className="flex gap-2 items-center">
            {moment.source && <SourceBadge source={moment.source} />}
            <Badge className={getIntensityColor(moment.intensity)}>
              {moment.intensity}/10
            </Badge>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <p className="line-clamp-2">{moment.situation}</p>
      </CardContent>
      {/* ... rest of card */}
    </Card>
  );
}
```

**Checklist Part 4**:
- [ ] Update `MomentCard.tsx` to show source badge
- [ ] Add subtle border for Telegram moments
- [ ] Test with mix of web and Telegram moments
- [ ] Verify responsive design

---

## Testing Checklist

### Unit Tests
- [ ] Test `useTelegram` hook
- [ ] Test `SourceBadge` component with different sources
- [ ] Test `TelegramConnectionCard` connected/disconnected states
- [ ] Test `TelegramLinkingFlow` timer countdown

### Integration Tests
- [ ] Test full linking flow (generate code → link → refresh)
- [ ] Test disconnection flow
- [ ] Test moment list with mixed sources
- [ ] Test settings page navigation

### Visual Tests
- [ ] Verify QR code displays correctly on desktop
- [ ] Verify deep link works on mobile
- [ ] Verify badge colors and styles
- [ ] Verify countdown timer updates in real-time
- [ ] Verify responsive design (mobile/tablet/desktop)

### E2E Tests (Manual)
- [ ] Open settings page
- [ ] Generate link code
- [ ] Scan QR code on phone (or use deep link)
- [ ] Complete linking in Telegram
- [ ] Return to web app, verify connected status
- [ ] Create moment via Telegram
- [ ] Refresh moments page, verify badge appears
- [ ] Disconnect Telegram
- [ ] Verify status updates

---

## Troubleshooting

### QR Code Not Showing
- Check if `qrcode.react` is installed
- Verify deep link format is correct
- Check console for errors

### Link Code Expired Too Fast
- Backend: Check link code TTL (should be 5 minutes)
- Frontend: Verify timer countdown logic

### Badge Not Showing
- Check if moment has `source: "telegram"` field
- Verify API response includes source
- Check SourceBadge component logic

---

**Status**: ⏭️ Ready for Implementation
**Total Time**: 17-23 hours
**Next Task**: [07-devops-complete.md](./07-devops-complete.md) - DevOps & Deployment
