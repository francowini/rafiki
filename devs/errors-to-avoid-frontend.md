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
// GOOD: Validate date before formatting, use consistent locale via default parameter
function formatDate(dateValue: string, locale = 'es-MX'): { dateStr: string; timeStr: string } {
  const date = new Date(dateValue);

  if (isNaN(date.getTime())) {
    return { dateStr: 'Fecha inválida', timeStr: '' };
  }

  const dateStr = date.toLocaleDateString(locale, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
  const timeStr = date.toLocaleTimeString(locale, {
    hour: '2-digit',
    minute: '2-digit',
  });

  return { dateStr, timeStr };
}
```

### Checklist

- [ ] All Date parsing validates with `isNaN(date.getTime())`
- [ ] Provide user-friendly fallback for invalid dates
- [ ] Use consistent locale across the app (default parameter or shared constant)
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

---

## F8. Async: Using Promise.all for Batch Operations with Partial Failure Risk

### Severity: Major (Partial State, Data Integrity)

### Problem

Using `Promise.all` for batch operations can leave the system in a partial state if some requests fail. The entire operation fails on the first rejection, potentially leaving some items processed and others not.

### Bad Example

```typescript
// BAD: Promise.all fails entirely on first rejection
const handleProceed = async () => {
  try {
    await Promise.all(
      activeLifeVisions.map((vision) => api.lifeVisions.archive(vision.id))
    );
    toast({ title: 'Success', description: 'All visions archived.' });

    // Archive the value - but some visions might have failed!
    await api.values.archive(valueToArchive.id);
  } catch (err) {
    toast({ variant: 'destructive', title: 'Error', description: err.message });
  }
};
```

### Good Example

```typescript
// GOOD: Promise.allSettled handles partial failures gracefully
const handleProceed = async () => {
  try {
    const results = await Promise.allSettled(
      activeLifeVisions.map((vision) => api.lifeVisions.archive(vision.id))
    );

    const succeeded = results.filter((r) => r.status === 'fulfilled').length;
    const failed = results.filter((r) => r.status === 'rejected').length;

    if (failed > 0) {
      console.error('Failed operations:', results.filter((r) => r.status === 'rejected'));
      toast({
        variant: 'destructive',
        title: 'Error parcial',
        description: `${succeeded} de ${total} completados. ${failed} fallaron.`,
      });
      return; // Don't proceed with dependent operation
    }

    // Only proceed if all operations succeeded
    await api.values.archive(valueToArchive.id);
    toast({ title: 'Success', description: 'All operations completed.' });
  } catch (err) {
    toast({ variant: 'destructive', title: 'Error', description: err.message });
  }
};
```

### Checklist

- [ ] Use `Promise.allSettled` instead of `Promise.all` for batch operations
- [ ] Count and report successes vs failures to the user
- [ ] Log failed operations for debugging
- [ ] Don't proceed with dependent operations unless all prerequisites succeeded

---

## F9. State Management: Not Resetting Dialog State on Close

### Severity: Medium (Stale Data, UX Bug)

### Problem

When a dialog closes, local state (like selections) may not be reset, causing stale values to appear when the dialog reopens.

### Bad Example

```typescript
// BAD: State persists after dialog closes
const [selectedValueId, setSelectedValueId] = useState<string>('');

const handleProceed = async () => {
  // ... process ...
  onComplete();
  onOpenChange(false); // Dialog closes but selectedValueId keeps its value
};
```

### Good Example

```typescript
// GOOD: Reset state before closing
const [selectedValueId, setSelectedValueId] = useState<string>('');

const handleProceed = async () => {
  // ... process ...

  // Reset local state before closing
  setSelectedValueId('');
  setArchiveVisions(false);
  onComplete();
  onOpenChange(false);
};

const handleCancel = () => {
  // Also reset on cancel
  setSelectedValueId('');
  setArchiveVisions(false);
  onOpenChange(false);
};
```

### Checklist

- [ ] Reset all local dialog state before calling `onOpenChange(false)`
- [ ] Reset state in both success and cancel paths
- [ ] Consider using a `resetState` helper function for consistency

---

## F10. Accessibility: Labels Not Associated with Form Controls

### Severity: Medium (Accessibility Violation)

### Problem

Labels without `htmlFor` attributes are not programmatically associated with their form controls, making the UI inaccessible to screen readers.

### Bad Example

```tsx
// BAD: Label not associated with Select
<div className="space-y-2">
  <label className="text-sm font-medium">Reasignar a:</label>
  <Select value={selectedValueId} onValueChange={setSelectedValueId}>
    <SelectTrigger>
      <SelectValue placeholder="Selecciona..." />
    </SelectTrigger>
  </Select>
</div>
```

### Good Example

```tsx
// GOOD: Label properly associated via htmlFor and id
<div className="space-y-2">
  <label htmlFor="reassign-value-select" className="text-sm font-medium">
    Reasignar a:
  </label>
  <Select value={selectedValueId} onValueChange={setSelectedValueId}>
    <SelectTrigger id="reassign-value-select">
      <SelectValue placeholder="Selecciona..." />
    </SelectTrigger>
  </Select>
</div>
```

### Checklist

- [ ] Add `htmlFor` to all `<label>` elements
- [ ] Add matching `id` to the associated form control
- [ ] Use unique IDs within the component scope

---

## F11. Error Handling: Missing Try-Catch for Follow-up Async Calls

### Severity: Medium (Unhandled Errors, Poor UX)

### Problem

When handling one error leads to additional async calls (e.g., fetching data for a dialog), those calls should also be wrapped in error handling.

### Bad Example

```typescript
// BAD: Follow-up calls not wrapped in try-catch
} catch (err) {
  if (err instanceof APIError && err.status === 409) {
    // These can throw but aren't caught!
    const visionsRes = await api.lifeVisions.getAll({ valueId: value.id });
    const valuesRes = await api.values.getAll();

    setActiveLifeVisions(visionsRes.items);
    setShowReassignmentDialog(true);
  }
}
```

### Good Example

```typescript
// GOOD: Nested try-catch for follow-up calls
} catch (err) {
  if (err instanceof APIError && err.status === 409) {
    try {
      const visionsRes = await api.lifeVisions.getAll({ valueId: value.id });
      const valuesRes = await api.values.getAll();

      setActiveLifeVisions(visionsRes.items);
      setShowReassignmentDialog(true);
    } catch (fetchErr) {
      console.error('Failed to fetch data:', fetchErr);
      toast({
        variant: 'destructive',
        title: 'Error al cargar datos',
        description: 'No se pudieron cargar los datos. Intenta de nuevo.',
      });
      // Clean up state
      setValueToArchive(null);
    }
  }
}
```

### Checklist

- [ ] Wrap all async calls in try-catch, including follow-up calls in catch blocks
- [ ] Log errors for debugging
- [ ] Show user-friendly error messages
- [ ] Clean up state on error (clear loading flags, reset selections)

---

## F12. React Query: Cache Key Not Including All Parameters

### Severity: Major (Cache Collision Bug)

### Problem

When React Query's queryKey doesn't include all parameters used in the queryFn, different requests can collide in the cache, returning stale data.

### Bad Example

```typescript
// BAD: queryKey only includes id, but queryFn uses params
export function useRecords(id: string, params?: { startDate?: string; endDate?: string }) {
  return useQuery({
    queryKey: ['records', id], // Missing params!
    queryFn: () => api.getRecords(id, params), // params affects response
    enabled: !!id,
  });
}
```

### Good Example

```typescript
// GOOD: queryKey includes all parameters that affect the response
export function useRecords(id: string, params?: { startDate?: string; endDate?: string }) {
  return useQuery({
    queryKey: ['records', id, params], // Includes params
    queryFn: () => api.getRecords(id, params),
    enabled: !!id,
  });
}
```

### Checklist

- [ ] queryKey includes all parameters that affect the response
- [ ] Use query key factories for consistent key generation
- [ ] Parameters in queryFn should be reflected in queryKey

---

## F13. Optimistic Updates: Weak Temporary ID Generation

### Severity: Medium (Race Condition)

### Problem

Using `Date.now()` for optimistic update IDs can collide when rapid submissions occur in the same millisecond.

### Bad Example

```typescript
// BAD: Can collide on rapid submissions
const optimisticRecord = {
  id: `temp-${Date.now()}`, // May produce duplicate IDs
  ...newData,
};
```

### Good Example

```typescript
// GOOD: Cryptographically unique ID
const optimisticRecord = {
  id: crypto.randomUUID(), // Guaranteed unique
  ...newData,
};
```

### Checklist

- [ ] Use `crypto.randomUUID()` for optimistic record IDs
- [ ] Ensure environment supports `crypto.randomUUID()` (modern browsers, Node 19+)

---

## F14. Route Parameters: Unvalidated Dynamic Segments

### Severity: Medium (Runtime Error / Security)

### Problem

Next.js dynamic route parameters (`params.id`) can be undefined, arrays, or unexpected types. Casting without validation leads to runtime errors.

### Bad Example

```typescript
// BAD: Assumes params.id is always a string
export default function DetailPage() {
  const params = useParams();
  const id = params.id as string; // Can be undefined or string[]

  return <Detail itemId={id} />; // May pass undefined
}
```

### Good Example

```typescript
// GOOD: Validate and handle edge cases
export default function DetailPage() {
  const params = useParams();

  // Handle undefined, array, or invalid values
  const rawId = params?.id;
  const id = Array.isArray(rawId) ? rawId[0] : rawId;

  if (!id || typeof id !== 'string') {
    notFound();
  }

  return <Detail itemId={id} />;
}
```

### Checklist

- [ ] Check if `params?.id` exists before using
- [ ] Handle both string and string[] types (use first element if array)
- [ ] Call `notFound()` for invalid or missing IDs

---

## F15. Rendering: State Updates During Render

### Severity: Medium (React Warning / Infinite Loop Risk)

### Problem

Calling `setState` during the render phase causes React warnings and can lead to infinite loops.

### Bad Example

```typescript
// BAD: setState during render
function Dialog({ open, objective }) {
  const [value, setValue] = useState('');
  const [initialized, setInitialized] = useState(false);

  if (open && objective && !initialized) {
    setValue(objective.value); // Called during render!
    setInitialized(true);
  }

  return <input value={value} />;
}
```

### Good Example

```typescript
// GOOD: setState in event handlers or effects
function Dialog({ open, objective, onOpenChange }) {
  const [value, setValue] = useState('');
  const [initialized, setInitialized] = useState(false);

  const handleOpenChange = useCallback(
    (newOpen: boolean) => {
      if (newOpen && objective && !initialized) {
        setValue(objective.value);
        setInitialized(true);
      }
      if (!newOpen) {
        setInitialized(false);
      }
      onOpenChange(newOpen);
    },
    [objective, onOpenChange, initialized],
  );

  return <Dialog open={open} onOpenChange={handleOpenChange}>...</Dialog>;
}
```

### Checklist

- [ ] Never call setState directly in the component body
- [ ] Use event handlers (onClick, onOpenChange) or useEffect for state updates
- [ ] Reset dialog state in the open/close handler, not during render

---

## F16. UI Controls: Not Disabling for Terminal States

### Severity: Medium (UX Bug)

### Problem

Interactive controls that modify state should be disabled when the entity is in a terminal (immutable) state.

### Bad Example

```typescript
// BAD: Status can be changed even for completed/abandoned items
<Select value={item.status} onValueChange={handleStatusChange}>
  {statuses.map((s) => (
    <SelectItem key={s} value={s}>
      {s}
    </SelectItem>
  ))}
</Select>
```

### Good Example

```typescript
// GOOD: Disable for terminal states with tooltip explanation
const isTerminal = ['completado', 'abandonado'].includes(item.status);

const handleStatusChange = (newStatus: string) => {
  if (isTerminal) return; // Guard in handler too
  updateStatus(newStatus);
};

<Select
  value={item.status}
  onValueChange={handleStatusChange}
  disabled={isTerminal}
>
  <SelectTrigger title={isTerminal ? 'Estado final - no se puede cambiar' : undefined}>
    <SelectValue />
  </SelectTrigger>
  {/* ... */}
</Select>
```

### Checklist

- [ ] Identify terminal/immutable states in your domain
- [ ] Disable controls that modify state when in terminal state
- [ ] Add tooltip or aria-label explaining why control is disabled
- [ ] Guard handler functions against terminal state changes

---

## F17. Functions: Silent Empty Returns for Unknown Values

### Severity: Low (Debugging Difficulty)

### Problem

Returning empty strings or null for unknown/invalid values makes debugging difficult and can cause confusing UI.

### Bad Example

```typescript
// BAD: Returns empty string for unknown type - confusing UI
function getLabel(type: string): string {
  if (type === 'a') return 'Type A';
  if (type === 'b') return 'Type B';
  return ''; // Unknown type produces blank UI
}
```

### Good Example

```typescript
// GOOD: Return fallback label and log warning
function getLabel(type: string, entityId?: string): string {
  if (type === 'a') return 'Type A';
  if (type === 'b') return 'Type B';

  console.warn(`Unknown type "${type}" for entity ${entityId}`);
  return 'Tipo desconocido';
}
```

### Checklist

- [ ] Return user-friendly fallback for unknown values
- [ ] Log warning with context (entity ID, actual value) for debugging
- [ ] Consider if unknown values should throw or be reported to error tracking
