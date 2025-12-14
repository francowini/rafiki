# Frontend Errors to Avoid (TypeScript/React)

This document catalogs critical errors specific to frontend (TypeScript/React) development.

---

## Table of Contents

- [F1. Security: Markdown Injection in User Content](#f1-security-markdown-injection-in-user-content)
- [F2. Async: Stale Responses in useEffect](#f2-async-stale-responses-in-useeffect)
- [F3. Data: Silent Data Truncation](#f3-data-silent-data-truncation)
- [F4. State: useSyncExternalStore + useState Duplication](#f4-state-usesyncexternalstore--usestate-duplication)
- [F5. Accessibility: Clickable Div Without Keyboard Support](#f5-accessibility-clickable-div-without-keyboard-support)
- [F6. Storage: Unguarded localStorage JSON.parse](#f6-storage-unguarded-localstorage-jsonparse)
- [F7. Dates: Unvalidated Date Parsing](#f7-dates-unvalidated-date-parsing)

---

## F1. Security: Markdown Injection in User Content

### Severity: CRITICAL (Security Vulnerability)

### Problem

When generating Markdown from user-provided content, failing to escape special characters allows users to inject malicious Markdown.

### Bad Example

```typescript
// BAD: User content inserted directly into Markdown
function formatMoment(moment: ExportItem): string[] {
  const lines: string[] = [];
  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(moment.situation); // User could inject: "# HACKED [click here](evil.com)"
  }
  return lines;
}
```

### Good Example

```typescript
// GOOD: Escape Markdown special characters in user content
function escapeMarkdown(text: string): string {
  return text.replace(/([\\*_\[\]()#+-.,!`>|{}])/g, '\\$1');
}

function formatMoment(moment: ExportItem): string[] {
  const lines: string[] = [];
  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(escapeMarkdown(moment.situation)); // Safe
  }
  return lines;
}
```

### Checklist

- [ ] All user-provided content is escaped before insertion into Markdown
- [ ] Escape function covers all Markdown special characters

---

## F2. Async: Stale Responses in useEffect

### Severity: Major (Race Condition Bug)

### Problem

When fetching data in `useEffect`, rapid state changes can cause stale responses to overwrite newer data.

### Bad Example

```typescript
// BAD: No guard against stale responses
useEffect(() => {
  const fetchData = async () => {
    const response = await api.getData(filter);
    setData(response); // Stale response may overwrite newer data!
  };
  fetchData();
}, [filter]);
```

### Good Example

```typescript
// GOOD: Request ID guards against stale responses
const requestIdRef = useRef(0);

useEffect(() => {
  const currentRequestId = ++requestIdRef.current;

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const response = await api.getData(filter);
      if (currentRequestId === requestIdRef.current) {
        setData(response);
      }
    } catch (err) {
      if (currentRequestId === requestIdRef.current) {
        setError(err);
      }
    } finally {
      if (currentRequestId === requestIdRef.current) {
        setIsLoading(false);
      }
    }
  };

  fetchData();
}, [filter]);
```

### Alternative: AbortController

```typescript
useEffect(() => {
  const controller = new AbortController();

  const fetchData = async () => {
    try {
      const response = await api.getData(filter, { signal: controller.signal });
      setData(response);
    } catch (err) {
      if (!controller.signal.aborted) {
        setError(err);
      }
    }
  };

  fetchData();
  return () => controller.abort();
}, [filter]);
```

### Checklist

- [ ] All async useEffect hooks have a freshness guard (requestId or AbortController)
- [ ] State updates are conditional on the request still being current

---

## F3. Data: Silent Data Truncation

### Severity: Major (Data Loss / UX Issue)

### Problem

When exporting paginated data, if only the first page is fetched but the UI suggests a full export, users unknowingly receive partial data.

### Bad Example

```typescript
// BAD: Only fetches first page, no truncation warning
const handleExport = async () => {
  const response = await api.export.getItems({ startDate, endDate, rows: 100 });
  generateMarkdown(response.items); // User gets only first 100 items!
};
```

### Good Example

```typescript
// GOOD: Block export when results are truncated
const hasMoreItems = !!exportData && exportData.items.length < exportData.total;

{hasMoreItems && (
  <p className="text-sm text-amber-600">
    Showing {exportData.items.length} of {exportData.total} entries.
    Consider using a smaller date range.
  </p>
)}

<Button onClick={handleExport} disabled={hasMoreItems}>
  Export
</Button>
```

### Checklist

- [ ] Compare `items.length` vs `total` to detect truncation
- [ ] Either block export or show prominent warning when truncated

---

## F4. State: useSyncExternalStore + useState Duplication

### Severity: Major (State Sync Bug)

### Problem

When using `useSyncExternalStore` to read external state, creating a separate `useState` that copies the initial value creates a state duplication bug.

### Bad Example

```typescript
// BAD: useState duplicates external store state and never syncs
const initialCollapsed = useSyncExternalStore(subscribe, getSnapshot);
const [collapsed, setCollapsed] = useState(initialCollapsed); // Never updates!
```

### Good Example

```typescript
// GOOD: Use useSyncExternalStore directly, no useState duplication
const collapsed = useSyncExternalStore(subscribe, getSnapshot, () => false);

const toggleCollapsed = useCallback(() => {
  const newState = !getSnapshot();
  localStorage.setItem(STORAGE_KEY, JSON.stringify(newState));
  window.dispatchEvent(new Event('sidebar-state-change'));
}, []);
```

### Checklist

- [ ] Never copy `useSyncExternalStore` result into `useState`
- [ ] Use custom events to trigger same-tab updates

---

## F5. Accessibility: Clickable Div Without Keyboard Support

### Severity: Major (Accessibility Violation)

### Problem

Using a `<div>` with `onClick` without proper ARIA attributes and keyboard handlers makes the component inaccessible.

### Bad Example

```typescript
// BAD: Clickable div without accessibility
<div className="cursor-pointer" onClick={() => setExpanded(!expanded)}>
  <h3>{title}</h3>
</div>
```

### Good Example (with role and keyboard)

```typescript
// GOOD: Proper accessibility for clickable element
const handleKeyDown = (e: React.KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault();
    setExpanded(!expanded);
  }
};

<div
  role="button"
  tabIndex={0}
  aria-expanded={expanded}
  onClick={() => setExpanded(!expanded)}
  onKeyDown={handleKeyDown}
>
  <h3>{title}</h3>
</div>
```

### Better Example (semantic button)

```typescript
// BETTER: Use semantic HTML when possible
<button
  type="button"
  aria-expanded={expanded}
  className="w-full text-left"
  onClick={() => setExpanded(!expanded)}
>
  <h3>{title}</h3>
</button>
```

### Checklist

- [ ] Clickable divs have `role="button"` and `tabIndex={0}`
- [ ] Include `onKeyDown` handler for Enter and Space keys
- [ ] Use `aria-expanded` for expandable elements
- [ ] Prefer semantic `<button>` when styling allows

---

## F6. Storage: Unguarded localStorage JSON.parse

### Severity: Major (Runtime Exception)

### Problem

Calling `JSON.parse()` on localStorage values without error handling can throw exceptions when the stored data is corrupted.

### Bad Example

```typescript
// BAD: JSON.parse can throw on corrupted data
function getPreference(): boolean {
  const saved = localStorage.getItem('preference');
  return saved !== null ? JSON.parse(saved) : false; // Throws if corrupted
}
```

### Good Example

```typescript
// GOOD: Wrap in try/catch and clean up corrupted values
function getPreference(): boolean {
  if (typeof window === 'undefined') return false;

  try {
    const saved = localStorage.getItem('preference');
    return saved !== null ? JSON.parse(saved) : false;
  } catch {
    localStorage.removeItem('preference');
    return false;
  }
}
```

### Checklist

- [ ] All `JSON.parse(localStorage.getItem(...))` wrapped in try/catch
- [ ] Return sensible default on parse error
- [ ] Optionally remove corrupted key

---

## F7. Dates: Unvalidated Date Parsing

### Severity: Major (UI Crash / Bad UX)

### Problem

Calling `toLocaleDateString()` on an invalid Date object displays "Invalid Date" to users.

### Bad Example

```typescript
// BAD: No validation - shows "Invalid Date" if malformed
const momentDate = new Date(moment.momentDate);
const dateStr = momentDate.toLocaleDateString('en-US', { ... });
```

### Good Example

```typescript
// GOOD: Validate date before formatting
function formatDate(dateValue: string): { dateStr: string; timeStr: string } {
  const date = new Date(dateValue);

  if (isNaN(date.getTime())) {
    return { dateStr: 'Invalid date', timeStr: '' };
  }

  const dateStr = date.toLocaleDateString('en-US', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
  const timeStr = date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });

  return { dateStr, timeStr };
}
```

### Checklist

- [ ] All Date parsing validates with `isNaN(date.getTime())`
- [ ] Provide user-friendly fallback for invalid dates
- [ ] Consider extracting to reusable utility function

---

## Quick Reference Checklist

- [ ] **Markdown**: User content is escaped before insertion into Markdown
- [ ] **Async**: useEffect with fetch uses requestId or AbortController
- [ ] **Pagination**: Exports handle truncation (warn or fetch all pages)
- [ ] **External Store**: Never copy `useSyncExternalStore` result into `useState`
- [ ] **Accessibility**: Clickable divs have `role="button"`, `tabIndex={0}`, keyboard handlers
- [ ] **Accessibility**: Prefer semantic `<button>` elements
- [ ] **localStorage**: Wrap `JSON.parse(localStorage.getItem(...))` in try/catch
- [ ] **Dates**: Validate dates with `isNaN(date.getTime())` before formatting
