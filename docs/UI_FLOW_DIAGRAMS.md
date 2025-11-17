# Telegram Integration - UI Flow Diagrams

## 1. Settings Page - Main Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Settings - Telegram Integration              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Status Card                                              │  │
│  │                                                          │  │
│  │  State: UNLINKED                                         │  │
│  │  ┌────────┐                                              │  │
│  │  │  (!)   │  Not linked to Telegram                      │  │
│  │  └────────┘                                              │  │
│  │                                                          │  │
│  │  [Link Telegram Account]                                 │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

                        ↓ User clicks "Link Telegram Account"

┌─────────────────────────────────────────────────────────────────┐
│                    Settings - Telegram Integration              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Link Code Display                                        │  │
│  │                                                          │  │
│  │              Your Link Code                              │  │
│  │                                                          │  │
│  │              ┌──────────┐                                │  │
│  │              │ AB12CD   │  ← Large, mono font            │  │
│  │              └──────────┘                                │  │
│  │                                                          │  │
│  │         Expires in ⏱️  4:32                               │  │
│  │                                                          │  │
│  │  ┌─────────────────────────────────────┐                │  │
│  │  │  [📱 Open Telegram]                 │  ← Mobile      │  │
│  │  └─────────────────────────────────────┘                │  │
│  │                                                          │  │
│  │  ┌──────────────────┐                                   │  │
│  │  │     QR CODE      │  ← Desktop only                   │  │
│  │  │   ▓▓▓▓▓▓▓▓▓▓▓    │                                   │  │
│  │  │   ▓▓▓▓▓▓▓▓▓▓▓    │                                   │  │
│  │  └──────────────────┘                                   │  │
│  │                                                          │  │
│  │  How to link:                                            │  │
│  │  1. Tap "Open Telegram" or scan QR code                  │  │
│  │  2. Send the code when prompted                          │  │
│  │  3. Wait for confirmation                                │  │
│  │                                                          │  │
│  │  [Cancel]                                                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

                        ↓ User completes linking in Telegram

┌─────────────────────────────────────────────────────────────────┐
│                    Settings - Telegram Integration              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Status Card                                              │  │
│  │                                                          │  │
│  │  State: LINKED                                           │  │
│  │  ┌────────┐                                              │  │
│  │  │   ✓    │  Linked to @username                         │  │
│  │  └────────┘  Connected 2 minutes ago                     │  │
│  │                                                          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Recent Imports (5)                            [View All] │  │
│  │                                                          │  │
│  │  ✓ Success  2 minutes ago      [View Moment]            │  │
│  │  ✓ Success  1 hour ago         [View Moment]            │  │
│  │  ✗ Failed   3 hours ago        [Retry]                  │  │
│  │     "Missing intensity value"                            │  │
│  │                                                          │  │
│  ��──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Danger Zone                                              │  │
│  │                                                          │  │
│  │  [Unlink Telegram Account]                               │  │
│  │                                                          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Mobile Flow - Step by Step

### Step 1: Initial State
```
┌─────────────────────────┐
│ ☰  Settings             │
├─────────────────────────┤
│                         │
│ Telegram Integration    │
│                         │
│ ┌─────────────────────┐ │
│ │   Status: OFF       │ │
│ │                     │ │
│ │   Not linked        │ │
│ │                     │ │
│ │ [Link Telegram]     │ │
│ └─────────────────────┘ │
│                         │
│                         │
│                         │
│                         │
│                         │
└─────────────────────────┘
```

### Step 2: Code Generation
```
┌─────────────────────────┐
│ ☰  Settings             │
├─────────────────────────┤
│                         │
│ Your Link Code          │
│                         │
│ ┌─────────────────────┐ │
│ │                     │ │
│ │     AB12CD          │ │
│ │                     │ │
│ └─────────────────────┘ │
│                         │
│ Expires in 4:32         │
│                         │
│ ┌─────────────────────┐ │
│ │ 📱 Open Telegram    │ │ ← Full width
│ └─────────────────────┘ │
│                         │
│ How to link (?)         │ ← Expandable
│                         │
│ [Cancel]                │
│                         │
└─────────────────────────┘
```

### Step 3: Waiting for Confirmation
```
┌─────────────────────────┐
│ ☰  Settings             │
├─────────────────────────┤
│                         │
│ Waiting for Telegram... │
│                         │
│ ┌─────────────────────┐ │
│ │                     │ │
│ │     ⏳ Loading      │ │
│ │                     │ │
│ └─────────────────────┘ │
│                         │
│ Checking status...      │
│                         │
│ This may take a moment  │
│                         │
│                         │
│ [Cancel]                │
│                         │
└─────────────────────────┘
```

### Step 4: Success Confirmation
```
┌─────────────────────────┐
│ ☰  Settings             │
├─────────────────────────┤
│                         │
│ ✓ Successfully Linked!  │
│                         │
│ ┌─────────────────────┐ │
│ │   Status: ON        │ │
│ │                     │ │
│ │   @username         │ │
│ │   Just now          │ │
│ └─────────────────────┘ │
│                         │
│ Recent Imports (0)      │
│                         │
│ You can now send        │
│ messages to @rafiki_bot │
│ to create moments       │
│                         │
└─────────────────────────┘
```

---

## 3. Error States

### Network Error
```
┌─────────────────────────────────────┐
│ Link Code Display                   │
├─────────────────────────────────────┤
│                                     │
│  ⚠️  Connection Failed              │
│                                     │
│  Could not generate link code.      │
│  Please check your internet         │
│  connection and try again.          │
│                                     │
│  [Retry]  [Cancel]                  │
│                                     │
└─────────────────────────────────────┘
```

### Code Expired
```
┌─────────────────────────────────────┐
│ Link Code Display                   │
├─────────────────────────────────────┤
│                                     │
│       AB12CD                        │
│                                     │
│  ⏱️  Code Expired                   │
│                                     │
│  This code is no longer valid.      │
│  Generate a new one to continue.    │
│                                     │
│  [Generate New Code]  [Cancel]      │
│                                     │
└─────────────────────────────────────┘
```

### Rate Limited
```
┌─────────────────────────────────────┐
│ Link Code Display                   │
├─────────────────────────────────────┤
│                                     │
│  ⏳ Too Many Attempts                │
│                                     │
│  You can try again in 28 seconds    │
│                                     │
│  [OK]                               │
│                                     │
└─────────────────────────────────────┘
```

### Import Failed
```
┌─────────────────────────────────────┐
│ Import History                      │
├─────────────────────────────────────┤
│                                     │
│  ✗ Failed  3 hours ago              │
│                                     │
│  Missing intensity value            │
│                                     │
│  💡 Tip: Make sure to answer all    │
│  questions in the conversation      │
│                                     │
│  [View in Telegram]                 │
│                                     │
└─────────────────────────────────────┘
```

---

## 4. Moments Page with Telegram Badge

```
┌──────────────────────────────────────────────────────────────────┐
│  Moments                                        🔄  [+ New]       │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────┐  ┌────────────────────┐  ┌─────────────┐│
│  │ 📱 Telegram  Nov 17│  │ 🌐 Web      Nov 16 │  │ 📱 Telegram ││
│  │                    │  │                    │  │             ││
│  │ Feeling anxious... │  │ Work stress...     │  │ Dinner with ││
│  │                    │  │                    │  │             ││
│  │ Intensity: 8/10    │  │ Intensity: 6/10    │  │ Intensity:  ││
│  │                    │  │                    │  │             ││
│  │ 2 hours ago        │  │ Yesterday          │  │ 3 days ago  ││
│  └────────────────────┘  └────────────────────┘  └─────────────┘│
│                                                                  │
│  ┌────────────────────┐  ┌────────────────────┐                 │
│  │ 🌐 Web      Nov 12 │  │ 📱 Telegram  Nov 10│                 │
│  │                    │  │                    │                 │
│  │ Family conflict... │  │ Lonely evening...  │                 │
│  │                    │  │                    │                 │
│  │ Intensity: 7/10    │  │ Intensity: 5/10    │                 │
│  │                    │  │                    │                 │
│  │ 5 days ago         │  │ 1 week ago         │                 │
│  └────────────────────┘  └────────────────────┘                 │
│                                                                  │
│                      [Previous]  Page 1/3  [Next]               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

Legend:
  📱 Telegram = Created via Telegram bot
  🌐 Web      = Created via web form
```

---

## 5. Unlink Confirmation Dialog

```
┌─────────────────────────────────────────────────┐
│                                                 │
│        Unlink Telegram Account?                 │
│                                                 │
│  Are you sure you want to unlink your           │
│  Telegram account (@username)?                  │
│                                                 │
│  • You won't be able to create moments via      │
│    Telegram                                     │
│  • Your existing moments will NOT be deleted    │
│  • You can re-link anytime                      │
│                                                 │
│              [Cancel]  [Unlink]                 │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## 6. Desktop Layout - Split View

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Settings - Telegram Integration                                             │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────────────────────┐  ┌─────────────────────────────────┐ │
│  │ Link Your Telegram                │  │ How It Works                    │ │
│  │                                   │  │                                 │ │
│  │ Your Code:                        │  │ 1. Get your unique link code    │ │
│  │                                   │  │                                 │ │
│  │       AB12CD                      │  │ 2. Open Telegram on your phone  │ │
│  │                                   │  │    or scan the QR code          │ │
│  │ Expires in 4:32                   │  │                                 │ │
│  │                                   │  │ 3. Send /start AB12CD to the    │ │
│  │ ┌──────────────┐                  │  │    @rafiki_moments_bot          │ │
│  │ │  QR CODE     │                  │  │                                 │ │
│  │ │  ▓▓▓▓▓▓▓▓▓▓▓  │                  │  │ 4. Start creating moments via   │ │
│  │ │  ▓▓▓▓▓▓▓▓▓▓▓  │                  │  │    conversational flow          │ │
│  │ │  ▓▓▓▓▓▓▓▓▓▓▓  │                  │  │                                 │ │
│  │ └──────────────┘                  │  │ ✓ Secure & Private              │ │
│  │                                   │  │ ✓ Your data stays encrypted     │ │
│  │ Scan with your phone camera       │  │ ✓ Unlink anytime                │ │
│  │                                   │  │                                 │ │
│  │ Or use this link:                 │  └─────────────────────────────────┘ │
│  │ https://t.me/rafiki_moments...    │                                     │
│  │                                   │                                     │
│  │ [Cancel]                          │                                     │
│  └───────────────────────────────────┘                                     │
│                                                                              │
└────────────────────────────────────────��─────────────────────────────────────┘
```

---

## 7. Import History - Expanded View

```
┌──────────────────────────────────────────────────────────────────┐
│  Import History                                      [← Back]     │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Showing last 30 days                                            │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ✓ Success                               2 minutes ago    │   │
│  │                                                          │   │
│  │ Moment: "Feeling anxious about work"                     │   │
│  │ Intensity: 8/10                          [View Moment]   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ✓ Success                               1 hour ago       │   │
│  │                                                          │   │
│  │ Moment: "Difficult conversation with friend"             │   │
│  │ Intensity: 6/10                          [View Moment]   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ✗ Failed                                3 hours ago      │   │
│  │                                                          │   │
│  │ Error: Missing intensity value                           │   │
│  │                                                          │   │
│  │ 💡 The bot asked for intensity (0-10) but didn't         │   │
│  │    receive a valid number.                               │   │
│  │                                                          │   │
│  │ [View in Telegram]                                       │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ✓ Success                               Yesterday        │   │
│  │                                                          │   │
│  │ Moment: "Stress about finances"                          │   │
│  │ Intensity: 7/10                          [View Moment]   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│                     [Previous]  Page 1/2  [Next]                │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 8. Loading States

### Generating Code
```
┌─────────────────────────┐
│                         │
│   Generating code...    │
│                         │
│   ⏳                    │
│                         │
└─────────────────────────┘
```

### Waiting for Link
```
┌─────────────────────────┐
│                         │
│   Waiting for           │
│   Telegram...           │
│                         │
│   ⏳ Checking status    │
│                         │
│   This may take         │
│   a moment              │
│                         │
└─────────────────────────┘
```

### Unlinking
```
┌─────────────────────────┐
│                         │
│   Unlinking...          │
│                         │
│   ⏳                    │
│                         │
└─────────────────────────┘
```

---

## 9. Navigation Flow

```
Dashboard (/)
    │
    ├─→ Moments (/momentos)
    │       │
    │       ├─→ Moment Detail (modal)
    │       │       └─→ Shows source badge
    │       │
    │       └─→ [🔄 Refresh] button
    │
    ├─→ Thinks (/thinks)
    │
    └─→ Settings (/settings)  ← NEW
            │
            ├─→ Telegram Integration (tab/section)
            │       │
            │       ├─→ Status Card
            │       │
            │       ├─→ Link Flow
            │       │   ├─→ Generate Code
            │       │   ├─→ Show Code + Timer
            │       │   ├─→ Poll Status
            │       │   └─→ Success/Error
            │       │
            │       ├─→ Import History
            │       │   └─→ List of recent imports
            │       │
            │       └─→ Unlink
            │           └─→ Confirmation dialog
            │
            └─→ Other settings (future)
```

---

## 10. Responsive Breakpoints

### Mobile (< 640px)
- Single column layout
- Full-width buttons (min 44px height)
- Deep link primary action
- QR code hidden
- Collapsed sections by default

### Tablet (640px - 1024px)
- 2-column layout for cards
- Medium buttons
- Show both deep link and QR
- Partially expanded sections

### Desktop (> 1024px)
- Side-by-side layout (code + instructions)
- QR code prominent
- Deep link as secondary action
- All sections expanded
- Hover states on cards

---

## 11. Color Coding

### Status Indicators
```
✓ Success    → Green (#10B981)
✗ Failed     → Red (#EF4444)
⏳ Pending   → Yellow (#F59E0B)
ℹ Info       → Blue (#3B82F6)
```

### Source Badges
```
📱 Telegram  → Blue background (#DBEAFE), Blue text (#1E40AF)
🌐 Web       → Gray background (#F3F4F6), Gray text (#4B5563)
```

### Urgency States
```
Normal       → Default colors
Warning      → Yellow tint (#FEF3C7)
Error        → Red tint (#FEE2E2)
Success      → Green tint (#D1FAE5)
```

---

**Document Version**: 1.0
**Last Updated**: 2025-11-17
