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

When generating Markdown from user-provided content, failing to escape special characters allows users to inject malicious Markdown structure. Additionally, if your Markdown renderer permits raw HTML or unsafe link protocols (e.g., `javascript:`, `data:`), escaping Markdown alone does NOT prevent XSS attacks.

### Security Layers

1. **Markdown Structure Injection**: Escaping special characters prevents users from creating headings, links, images, etc.
2. **XSS via HTML**: Many Markdown renderers allow raw HTML - use the renderer's safe mode or a sanitizer
3. **Unsafe Links**: Links with `javascript:` or `data:` protocols can execute code - whitelist protocols (http, https, mailto)

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

### Simple Example (Markdown Structure Only)

```typescript
// SIMPLE EXAMPLE: Basic escape for common Markdown characters
// NOTE: This is a minimal, non-exhaustive escape. It does NOT handle:
// - GFM extensions (tables, strikethrough ~, task lists)
// - Autolinks, code block language specifiers
// - Parser-specific syntax variations
// For production, use a well-maintained library (see recommendations below)
function escapeMarkdownBasic(text: string): string {
  return text.replace(/([\\*_\[\]()#+-.,!`>|{}])/g, '\\$1');
}

function formatMoment(moment: ExportItem): string[] {
  const lines: string[] = [];
  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(escapeMarkdownBasic(moment.situation)); // Prevents structure injection
  }
  return lines;
}
```

### Production Recommendations

For robust security, use established libraries:

- **Markdown escaping**: Use your Markdown parser's API (e.g., `markdown-it` plugins, `rehype-sanitize`)
- **HTML sanitization**: `DOMPurify`, `sanitize-html`, or `isomorphic-dompurify`
- **Safe links**: Add `rel="noopener noreferrer"` and whitelist protocols (http, https, mailto)

```typescript
// PRODUCTION: Use DOMPurify for HTML output
import DOMPurify from 'dompurify';

const sanitizedHtml = DOMPurify.sanitize(markdownToHtml(userContent), {
  ALLOWED_TAGS: ['p', 'strong', 'em', 'ul', 'ol', 'li'],
  ALLOWED_ATTR: [],
});
```

### Checklist

- [ ] User content is escaped before insertion into Markdown (prevents structure injection)
- [ ] Markdown renderer is configured in safe mode OR output is sanitized (prevents XSS)
- [ ] Links use `rel="noopener noreferrer"` and only allow safe protocols
- [ ] For production, use a well-maintained sanitization library instead of custom regex

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

**Important**: These freshness guards also prevent state updates after the component unmounts, avoiding the React warning "Can't perform a React state update on an unmounted component" and keeping logs/tests clean.

### Checklist

- [ ] All async useEffect hooks have a freshness guard (requestId or AbortController)
- [ ] State updates are conditional on the request still being current and the component not unmounted

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

Calling `JSON.parse()` on localStorage values without error handling can throw exceptions when the stored data is corrupted. Additionally, even when parsing succeeds, the value may not be the expected type (e.g., stored as `"yes"` or `123` instead of a boolean).

### Bad Example

```typescript
// BAD: JSON.parse can throw on corrupted data AND may return wrong type
function getPreference(): boolean {
  const saved = localStorage.getItem('preference');
  return saved !== null ? JSON.parse(saved) : false; // Throws if corrupted, or returns "yes"/123/{}
}
```

### Good Example

```typescript
// GOOD: Wrap in try/catch, validate type, and clean up invalid values
function getPreference(): boolean {
  if (typeof window === 'undefined') return false;

  try {
    const saved = localStorage.getItem('preference');
    if (saved === null) return false;

    const parsed = JSON.parse(saved);

    // Validate the parsed value is actually a boolean
    if (typeof parsed !== 'boolean') {
      localStorage.removeItem('preference');
      return false;
    }

    return parsed;
  } catch {
    localStorage.removeItem('preference');
    return false;
  }
}
```

### Checklist

- [ ] All `JSON.parse(localStorage.getItem(...))` wrapped in try/catch
- [ ] Validate parsed value's type matches expected type (e.g., `typeof parsed === 'boolean'`)
- [ ] Return sensible default when type is invalid
- [ ] Remove corrupted or invalid keys from localStorage

---

## F7. Dates: Unvalidated Date Parsing

### Severity: Major (UI Crash / Bad UX)

### Problem

Calling `toLocaleDateString()` on an invalid Date object displays "Invalid Date" to users.

### Bad Example

```typescript
// BAD: No validation - shows "Invalid Date" if malformed
const momentDate = new Date(moment.momentDate);
const dateStr = momentDate.toLocaleDateString('es-MX', { ... });
```

### Good Example

```typescript
// GOOD: Validate date before formatting, use app locale constant
import { APP_LOCALE } from '@/lib/constants'; // e.g., 'es-MX'

function formatDate(dateValue: string): { dateStr: string; timeStr: string } {
  const date = new Date(dateValue);

  if (isNaN(date.getTime())) {
    return { dateStr: 'Fecha inválida', timeStr: '' };
  }

  const dateStr = date.toLocaleDateString(APP_LOCALE, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
  const timeStr = date.toLocaleTimeString(APP_LOCALE, {
    hour: '2-digit',
    minute: '2-digit',
  });

  return { dateStr, timeStr };
}
```

### Checklist

- [ ] All Date parsing validates with `isNaN(date.getTime())`
- [ ] Provide user-friendly fallback for invalid dates
- [ ] Use the app's locale constant (e.g., `APP_LOCALE`) instead of hardcoded locale strings
- [ ] Consider extracting to reusable utility function in `lib/date-utils.ts`

---

## Quick Reference Checklist

- [ ] **Markdown**: User content is escaped AND renderer is in safe mode or output is sanitized
- [ ] **Async**: useEffect with fetch uses requestId or AbortController (also prevents unmount warnings)
- [ ] **Pagination**: Exports handle truncation (warn or fetch all pages)
- [ ] **External Store**: Never copy `useSyncExternalStore` result into `useState`
- [ ] **Accessibility**: Clickable divs have `role="button"`, `tabIndex={0}`, keyboard handlers
- [ ] **Accessibility**: Prefer semantic `<button>` elements
- [ ] **localStorage**: Wrap `JSON.parse` in try/catch AND validate parsed type
- [ ] **Dates**: Validate dates with `isNaN(date.getTime())` and use app locale constant
