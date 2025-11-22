# Values Feature - Frontend Implementation Guide

**Project:** Rafiki Habits Tracker
**Feature:** Personal Values Tracking Frontend
**Date:** November 22, 2025
**Status:** Ready for Implementation

---

## Table of Contents

1. [TypeScript Types](#1-typescript-types)
2. [API Client](#2-api-client)
3. [Utility Functions](#3-utility-functions)
4. [Components](#4-components)
5. [Pages](#5-pages)
6. [Testing Checklist](#6-testing-checklist)

---

## 1. TypeScript Types

**File:** `frontend/lib/types.ts`

Add the following types to the existing `types.ts` file:

```typescript
// ============================================================================
// Value Types
// ============================================================================

export type Facet =
  | 'health'
  | 'relationships'
  | 'career'
  | 'personal_growth'
  | 'family'
  | 'creativity'
  | 'community'
  | 'spirituality'
  | 'leisure'
  | 'financial';

export interface Value {
  id: string;
  content: string;
  facet: Facet;
  displayOrder: number;
  dateCreated: string; // ISO 8601
  dateUpdated: string; // ISO 8601
}

export interface NewValue {
  content: string;
  facet: Facet;
  displayOrder: number;
}

export interface UpdateValue {
  content?: string;
  facet?: Facet;
  displayOrder?: number;
}

export interface ValueListResponse {
  items: Value[];
  total: number;
}
```

---

## 2. API Client

**File:** `frontend/lib/api.ts`

Add the following to the `api` export object:

```typescript
// Add to imports at the top
import {
  // ... existing imports
  Value,
  ValueListResponse,
  NewValue,
  UpdateValue,
  Facet,
} from './types';

// Add to the api object
export const api = {
  // ... existing endpoints (thinks, moments, health)

  values: {
    /**
     * Get all values for the current user, sorted by display_order
     */
    getAll: async (params?: { facet?: Facet }): Promise<ValueListResponse> => {
      const queryParams = new URLSearchParams();
      if (params?.facet) queryParams.set('facet', params.facet);

      const query = queryParams.toString();
      const endpoint = `/v1/values${query ? `?${query}` : ''}`;

      return fetchAPI<ValueListResponse>(endpoint);
    },

    /**
     * Get a single value by ID
     */
    getById: async (id: string): Promise<Value> => {
      return fetchAPI<Value>(`/v1/values/${id}`);
    },

    /**
     * Create a new value
     */
    create: async (data: NewValue): Promise<Value> => {
      return fetchAPI<Value>('/v1/values', {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    /**
     * Update an existing value
     */
    update: async (id: string, data: UpdateValue): Promise<Value> => {
      return fetchAPI<Value>(`/v1/values/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },

    /**
     * Delete a value
     */
    delete: async (id: string): Promise<void> => {
      return fetchAPI<void>(`/v1/values/${id}`, {
        method: 'DELETE',
      });
    },
  },
};
```

---

## 3. Utility Functions

**File:** `frontend/lib/value-utils.ts` (NEW FILE)

```typescript
import { Facet } from './types';

export interface FacetConfig {
  label: string;
  description: string;
  color: string;
  bgColor: string;
  borderColor: string;
  icon: string;
}

/**
 * Configuration for each life facet
 */
export const FACET_CONFIG: Record<Facet, FacetConfig> = {
  health: {
    label: 'Health',
    description: 'Physical and mental wellbeing',
    color: 'text-emerald-700',
    bgColor: 'bg-emerald-100',
    borderColor: 'border-emerald-300',
    icon: '🏃',
  },
  relationships: {
    label: 'Relationships',
    description: 'Family, friends, connections',
    color: 'text-blue-700',
    bgColor: 'bg-blue-100',
    borderColor: 'border-blue-300',
    icon: '🤝',
  },
  career: {
    label: 'Career',
    description: 'Professional development, work',
    color: 'text-amber-700',
    bgColor: 'bg-amber-100',
    borderColor: 'border-amber-300',
    icon: '💼',
  },
  personal_growth: {
    label: 'Personal Growth',
    description: 'Self-improvement, learning',
    color: 'text-purple-700',
    bgColor: 'bg-purple-100',
    borderColor: 'border-purple-300',
    icon: '🌱',
  },
  family: {
    label: 'Family',
    description: 'Family relationships',
    color: 'text-pink-700',
    bgColor: 'bg-pink-100',
    borderColor: 'border-pink-300',
    icon: '👨‍👩‍👧‍👦',
  },
  creativity: {
    label: 'Creativity',
    description: 'Creative expression, arts',
    color: 'text-orange-700',
    bgColor: 'bg-orange-100',
    borderColor: 'border-orange-300',
    icon: '🎨',
  },
  community: {
    label: 'Community',
    description: 'Service, contribution, helping others',
    color: 'text-green-700',
    bgColor: 'bg-green-100',
    borderColor: 'border-green-300',
    icon: '🤲',
  },
  spirituality: {
    label: 'Spirituality',
    description: 'Meaning, purpose, transcendence',
    color: 'text-indigo-700',
    bgColor: 'bg-indigo-100',
    borderColor: 'border-indigo-300',
    icon: '🙏',
  },
  leisure: {
    label: 'Leisure',
    description: 'Rest, recreation, enjoyment',
    color: 'text-cyan-700',
    bgColor: 'bg-cyan-100',
    borderColor: 'border-cyan-300',
    icon: '🎭',
  },
  financial: {
    label: 'Financial',
    description: 'Financial security, stability',
    color: 'text-teal-700',
    bgColor: 'bg-teal-100',
    borderColor: 'border-teal-300',
    icon: '💰',
  },
};

/**
 * Get facet configuration
 */
export function getFacetConfig(facet: Facet): FacetConfig {
  return FACET_CONFIG[facet];
}

/**
 * Get all facets as array
 */
export function getAllFacets(): Facet[] {
  return Object.keys(FACET_CONFIG) as Facet[];
}

/**
 * Format facet for display (e.g., "personal_growth" -> "Personal Growth")
 */
export function formatFacetLabel(facet: Facet): string {
  return FACET_CONFIG[facet].label;
}
```

---

## 4. Components

### 4.1 ValueCard Component

**File:** `frontend/components/features/ValueCard.tsx` (NEW FILE)

```typescript
'use client';

import { Value } from '@/lib/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Edit, Trash2 } from 'lucide-react';
import { getFacetConfig } from '@/lib/value-utils';

interface ValueCardProps {
  value: Value;
  rank: number; // 1-based ranking (1 = most important)
  onEdit?: () => void;
  onDelete?: () => void;
}

export function ValueCard({ value, rank, onEdit, onDelete }: ValueCardProps) {
  const facetConfig = getFacetConfig(value.facet);
  const isTopValue = rank === 1;

  return (
    <Card
      className={`transition-all hover:shadow-md ${
        isTopValue
          ? 'ring-2 ring-rose-500 bg-gradient-to-br from-rose-50 to-white'
          : 'hover:border-rose-200'
      }`}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 flex-wrap">
            {/* Priority Badge */}
            <Badge
              variant="outline"
              className={
                isTopValue
                  ? 'bg-rose-100 text-rose-800 border-rose-300 font-semibold'
                  : 'bg-gray-100 text-gray-700 border-gray-300'
              }
            >
              {isTopValue ? '#1 Core Value' : `#${rank}`}
            </Badge>

            {/* Facet Badge */}
            <Badge
              variant="outline"
              className={`${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
            >
              <span className="mr-1">{facetConfig.icon}</span>
              {facetConfig.label}
            </Badge>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Content */}
        <p className={`${isTopValue ? 'text-base font-medium' : 'text-sm'} text-foreground leading-relaxed`}>
          {value.content}
        </p>

        {/* Actions */}
        <div className="flex gap-2 pt-2">
          {onEdit && (
            <Button size="sm" variant="outline" onClick={onEdit} className="flex-1">
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

        {/* Metadata (small timestamp) */}
        <p className="text-xs text-muted-foreground">
          Updated {new Date(value.dateUpdated).toLocaleDateString()}
        </p>
      </CardContent>
    </Card>
  );
}
```

---

### 4.2 ValueForm Component

**File:** `frontend/components/features/ValueForm.tsx` (NEW FILE)

```typescript
'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/lib/api';
import { NewValue, Value, Facet } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getAllFacets, getFacetConfig } from '@/lib/value-utils';

const valueSchema = z.object({
  content: z
    .string()
    .min(3, 'Value must be at least 3 characters')
    .max(200, 'Value must be less than 200 characters'),
  facet: z.enum([
    'health',
    'relationships',
    'career',
    'personal_growth',
    'family',
    'creativity',
    'community',
    'spirituality',
    'leisure',
    'financial',
  ] as const),
  displayOrder: z.number().min(1).max(10),
});

type ValueFormData = z.infer<typeof valueSchema>;

interface ValueFormProps {
  value?: Value | null; // If provided, form is in edit mode
  existingValuesCount: number; // Number of existing values (for validation)
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function ValueForm({ value, existingValuesCount, onSuccess, onCancel }: ValueFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditMode = !!value;

  // Calculate next display order
  const nextDisplayOrder = isEditMode ? value.displayOrder : existingValuesCount + 1;

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
    reset,
  } = useForm<ValueFormData>({
    resolver: zodResolver(valueSchema),
    defaultValues: {
      facet: 'personal_growth',
      displayOrder: nextDisplayOrder,
    },
  });

  // Reset form when value changes (edit mode)
  useEffect(() => {
    if (value) {
      reset({
        content: value.content,
        facet: value.facet,
        displayOrder: value.displayOrder,
      });
    } else {
      reset({
        content: '',
        facet: 'personal_growth',
        displayOrder: nextDisplayOrder,
      });
    }
  }, [value, reset, nextDisplayOrder]);

  const selectedFacet = watch('facet');

  const onSubmit = async (data: ValueFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      // Check limit (only for create mode)
      if (!isEditMode && existingValuesCount >= 10) {
        setError('You can only have up to 10 values. Please delete one before adding a new value.');
        setIsSubmitting(false);
        return;
      }

      if (isEditMode && value) {
        // Update existing value
        await api.values.update(value.id, {
          content: data.content,
          facet: data.facet,
          displayOrder: data.displayOrder,
        });
      } else {
        // Create new value
        const newValue: NewValue = {
          content: data.content,
          facet: data.facet,
          displayOrder: data.displayOrder,
        };
        await api.values.create(newValue);
      }

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || `Error ${isEditMode ? 'updating' : 'creating'} value`);
    } finally {
      setIsSubmitting(false);
    }
  };

  const allFacets = getAllFacets();
  const selectedFacetConfig = selectedFacet ? getFacetConfig(selectedFacet) : null;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Privacy Message */}
      <Alert className="bg-rose-50 border-rose-200">
        <AlertDescription className="text-sm text-rose-900">
          Your values are private and only visible to you.
        </AlertDescription>
      </Alert>

      {/* Error Message */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Content Field */}
      <div className="space-y-2">
        <Label htmlFor="content" className="text-base font-medium">
          Your value statement
        </Label>
        <Textarea
          id="content"
          {...register('content')}
          placeholder="Example: Live with integrity and authenticity in all my relationships"
          className="min-h-24 text-base"
          maxLength={200}
        />
        <p className="text-sm text-muted-foreground">
          Describe what matters most to you (3-200 characters)
        </p>
        {errors.content && <p className="text-sm text-destructive">{errors.content.message}</p>}
      </div>

      {/* Facet Selection */}
      <div className="space-y-2">
        <Label htmlFor="facet" className="text-base font-medium">
          Life area
        </Label>
        <Select
          value={selectedFacet}
          onValueChange={(value) => setValue('facet', value as Facet)}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select a life area" />
          </SelectTrigger>
          <SelectContent>
            {allFacets.map((facet) => {
              const config = getFacetConfig(facet);
              return (
                <SelectItem key={facet} value={facet}>
                  <div className="flex items-center gap-2">
                    <span>{config.icon}</span>
                    <div>
                      <div className="font-medium">{config.label}</div>
                      <div className="text-xs text-muted-foreground">{config.description}</div>
                    </div>
                  </div>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
        {selectedFacetConfig && (
          <p className="text-sm text-muted-foreground">{selectedFacetConfig.description}</p>
        )}
        {errors.facet && <p className="text-sm text-destructive">{errors.facet.message}</p>}
      </div>

      {/* Display Order Field */}
      <div className="space-y-2">
        <Label htmlFor="displayOrder" className="text-base font-medium">
          Priority (1 = most important)
        </Label>
        <Select
          value={watch('displayOrder').toString()}
          onValueChange={(value) => setValue('displayOrder', parseInt(value))}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((num) => (
              <SelectItem key={num} value={num.toString()}>
                #{num} {num === 1 && '(Core Value)'}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-sm text-muted-foreground">
          Rank your values from most to least important
        </p>
        {errors.displayOrder && (
          <p className="text-sm text-destructive">{errors.displayOrder.message}</p>
        )}
      </div>

      {/* Action Buttons */}
      <div className="flex gap-3 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} className="flex-1">
            Cancel
          </Button>
        )}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="flex-1 bg-rose-600 hover:bg-rose-700"
        >
          {isSubmitting
            ? isEditMode
              ? 'Updating...'
              : 'Saving...'
            : isEditMode
              ? 'Update value'
              : 'Save value'}
        </Button>
      </div>
    </form>
  );
}
```

---

### 4.3 ValueList Component

**File:** `frontend/components/features/ValueList.tsx` (NEW FILE)

```typescript
'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Value } from '@/lib/types';
import { ValueCard } from './ValueCard';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2 } from 'lucide-react';

interface ValueListProps {
  refresh?: number; // Trigger refetch when this changes
  onValueEdit?: (value: Value) => void;
  onValueDelete?: (value: Value) => void;
}

export function ValueList({ refresh, onValueEdit, onValueDelete }: ValueListProps) {
  const [values, setValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchValues = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await api.values.getAll();
        // Sort by displayOrder (already sorted by backend, but double-check)
        const sortedValues = response.items.sort((a, b) => a.displayOrder - b.displayOrder);
        setValues(sortedValues);
      } catch (err: any) {
        setError(err.message || 'Failed to load values');
      } finally {
        setIsLoading(false);
      }
    };

    fetchValues();
  }, [refresh]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-rose-600" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (values.length === 0) {
    return (
      <Alert className="bg-rose-50 border-rose-200">
        <AlertDescription className="text-rose-900">
          You haven&apos;t created any values yet. Start by defining what matters most to you.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {values.map((value, index) => (
        <ValueCard
          key={value.id}
          value={value}
          rank={index + 1}
          onEdit={() => onValueEdit?.(value)}
          onDelete={() => onValueDelete?.(value)}
        />
      ))}
    </div>
  );
}
```

---

### 4.4 ValuesPreview Component (Optional Dashboard Widget)

**File:** `frontend/components/dashboard/ValuesPreview.tsx` (NEW FILE)

```typescript
'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Value } from '@/lib/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { getFacetConfig } from '@/lib/value-utils';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

export function ValuesPreview() {
  const [values, setValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchValues = async () => {
      try {
        const response = await api.values.getAll();
        // Show top 3 values
        setValues(response.items.slice(0, 3));
      } catch (err) {
        console.error('Failed to load values preview:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchValues();
  }, []);

  if (isLoading || values.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Your Core Values</CardTitle>
            <CardDescription>What matters most to you</CardDescription>
          </div>
          <Link
            href="/values"
            className="text-rose-600 hover:text-rose-700 flex items-center gap-1 text-sm font-medium"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {values.map((value, index) => {
          const facetConfig = getFacetConfig(value.facet);
          return (
            <div key={value.id} className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <Badge
                variant="outline"
                className={
                  index === 0
                    ? 'bg-rose-100 text-rose-800 border-rose-300 font-semibold shrink-0'
                    : 'bg-gray-100 text-gray-700 border-gray-300 shrink-0'
                }
              >
                #{index + 1}
              </Badge>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium line-clamp-2">{value.content}</p>
                <Badge
                  variant="outline"
                  className={`mt-1 text-xs ${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
                >
                  {facetConfig.icon} {facetConfig.label}
                </Badge>
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
```

---

## 5. Pages

### 5.1 Values Page

**File:** `frontend/app/(dashboard)/values/page.tsx` (NEW FILE)

```typescript
'use client';

import { useState } from 'react';
import { ValueForm } from '@/components/features/ValueForm';
import { ValueList } from '@/components/features/ValueList';
import { Value } from '@/lib/types';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, Info } from 'lucide-react';
import { api } from '@/lib/api';

export default function ValuesPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [valueToEdit, setValueToEdit] = useState<Value | null>(null);
  const [valueToDelete, setValueToDelete] = useState<Value | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [valuesCount, setValuesCount] = useState(0);

  const handleCreateSuccess = () => {
    setIsFormOpen(false);
    setRefresh((prev) => prev + 1);
  };

  const handleEditSuccess = () => {
    setIsEditFormOpen(false);
    setValueToEdit(null);
    setRefresh((prev) => prev + 1);
  };

  const handleEdit = (value: Value) => {
    setValueToEdit(value);
    setIsEditFormOpen(true);
  };

  const handleDelete = async () => {
    if (!valueToDelete) return;

    try {
      await api.values.delete(valueToDelete.id);
      setValueToDelete(null);
      setRefresh((prev) => prev + 1);
    } catch (err) {
      console.error('Error deleting value:', err);
    }
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-rose-900">Values</h1>
            <p className="text-muted-foreground mt-1">Define what matters most in your life</p>
          </div>

          <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
            <SheetTrigger asChild>
              <Button size="lg" className="bg-rose-600 hover:bg-rose-700">
                <Plus className="h-5 w-5 mr-2" />
                New value
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
              <SheetHeader>
                <SheetTitle className="text-xl">Define a value</SheetTitle>
                <SheetDescription>
                  What do you care about? What guides your decisions and actions?
                </SheetDescription>
              </SheetHeader>
              <div className="mt-6">
                <ValueForm
                  existingValuesCount={valuesCount}
                  onSuccess={handleCreateSuccess}
                  onCancel={() => setIsFormOpen(false)}
                />
              </div>
            </SheetContent>
          </Sheet>
        </div>

        {/* Info Alert */}
        <Alert className="bg-rose-50 border-rose-200">
          <Info className="h-4 w-4 text-rose-600" />
          <AlertDescription className="text-sm text-rose-900">
            You can define up to <strong>10 core values</strong>. Values are ranked by priority,
            with #1 being your most important value. Choose values that truly guide your life
            decisions.
          </AlertDescription>
        </Alert>
      </div>

      {/* Value List */}
      <ValueList
        refresh={refresh}
        onValueEdit={handleEdit}
        onValueDelete={setValueToDelete}
      />

      {/* Edit Form Sheet */}
      <Sheet open={isEditFormOpen} onOpenChange={setIsEditFormOpen}>
        <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl">Edit value</SheetTitle>
            <SheetDescription>Update your value statement or priority.</SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <ValueForm
              value={valueToEdit}
              existingValuesCount={valuesCount}
              onSuccess={handleEditSuccess}
              onCancel={() => {
                setIsEditFormOpen(false);
                setValueToEdit(null);
              }}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={!!valueToDelete}
        onOpenChange={(open) => !open && setValueToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete this value. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
```

---

### 5.2 Dashboard Integration (Optional)

**File:** `frontend/app/(dashboard)/page.tsx` (MODIFY EXISTING)

Add the values preview to the dashboard:

```typescript
// Add to imports
import { ValuesPreview } from '@/components/dashboard/ValuesPreview';

// Inside the dashboard page component, add to the grid:
export default function DashboardPage() {
  return (
    <div className="container mx-auto py-8 px-4">
      {/* ... existing dashboard content ... */}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
        {/* ... existing cards ... */}

        {/* Values Preview Card */}
        <ValuesPreview />
      </div>
    </div>
  );
}
```

---

## 6. Testing Checklist

### 6.1 Form Validation

- [ ] Content field enforces 3-200 character limit
- [ ] Form shows error for empty content
- [ ] Form shows error for content < 3 characters
- [ ] Form shows error for content > 200 characters
- [ ] Facet selection is required
- [ ] Display order selection works (1-10)

### 6.2 CRUD Operations

- [ ] Create: Can create a new value
- [ ] Create: Shows error when trying to create 11th value
- [ ] Read: Values load on page mount
- [ ] Read: Values are sorted by display_order
- [ ] Update: Can edit existing value
- [ ] Update: Changes persist after refresh
- [ ] Delete: Can delete a value
- [ ] Delete: Confirmation dialog appears before delete

### 6.3 UI/UX

- [ ] #1 value has special styling (ring border, gradient background)
- [ ] Facet badges show correct colors and icons
- [ ] Forms open in Sheet (side panel)
- [ ] Sheet scrolls properly on mobile
- [ ] Cards are responsive (1/2/3 columns)
- [ ] Loading state shows spinner
- [ ] Error states show error messages
- [ ] Empty state shows helpful message

### 6.4 Responsive Design

- [ ] Mobile (< 640px): Single column layout
- [ ] Tablet (640-1024px): 2 column layout
- [ ] Desktop (> 1024px): 3 column layout
- [ ] Sheet forms work on mobile
- [ ] Touch targets are minimum 44px

### 6.5 Accessibility

- [ ] All form fields have labels
- [ ] Keyboard navigation works (Tab order)
- [ ] Focus indicators are visible
- [ ] ARIA labels on interactive elements
- [ ] Screen reader friendly

### 6.6 Linting & Formatting

```bash
cd frontend

# Run all checks
npm run check

# Individual checks
npm run lint
npm run format:check
npm run typecheck

# Auto-fix
npm run lint:fix
npm run format
```

---

## Implementation Order

Follow this order for smooth implementation:

1. **Step 1: Types & Utilities**
   - Add types to `lib/types.ts`
   - Create `lib/value-utils.ts`

2. **Step 2: API Client**
   - Add values endpoints to `lib/api.ts`

3. **Step 3: Components (in order)**
   - Create `ValueCard.tsx`
   - Create `ValueForm.tsx`
   - Create `ValueList.tsx`
   - Create `ValuesPreview.tsx` (optional)

4. **Step 4: Pages**
   - Create `app/(dashboard)/values/page.tsx`
   - Update `app/(dashboard)/page.tsx` (optional)

5. **Step 5: Testing**
   - Test locally with `npm run dev`
   - Run linting: `npm run check`
   - Test all CRUD operations
   - Test responsive breakpoints

6. **Step 6: Deployment**
   - Commit and push changes
   - Deploy with `vercel --prod`
   - Verify in production

---

## Notes

- **Color Scheme**: Rose/crimson (#E11D48) is the primary feature color
- **Max Values**: Hard limit of 10 values enforced in both frontend and backend
- **Encryption**: Content is encrypted by backend (transparent to frontend)
- **No Pagination**: Max 10 values means no pagination needed
- **Simple Refetch**: Use `useState` refresh counter for re-fetching (no React Query)
- **Mobile-First**: Forms use Sheet component for mobile-friendly UX
- **No Emoji Support**: V1 doesn't include emoji support (future feature)

---

**Document Version:** 1.0
**Last Updated:** November 22, 2025
**Status:** Ready for Implementation
