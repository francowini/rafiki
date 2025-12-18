'use client';

import { useState, useMemo } from 'react';
import { Value } from '@/lib/types';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { FACET_CONFIG } from '@/lib/value-utils';

interface ValueSelectorProps {
  values: Value[];
  selectedValueIds: string[];
  onChange: (valueIds: string[]) => void;
  disabled?: boolean;
}

// Group values by facet
function groupBy<T>(arr: T[], key: keyof T): Record<string, T[]> {
  return arr.reduce(
    (acc, item) => {
      const groupKey = String(item[key]);
      if (!acc[groupKey]) {
        acc[groupKey] = [];
      }
      acc[groupKey].push(item);
      return acc;
    },
    {} as Record<string, T[]>,
  );
}

export function ValueSelector({
  values,
  selectedValueIds,
  onChange,
  disabled,
}: ValueSelectorProps) {
  const [searchQuery, setSearchQuery] = useState('');

  // Group values by facet
  const groupedValues = useMemo(() => {
    const filtered = values.filter((v) =>
      v.content.toLowerCase().includes(searchQuery.toLowerCase()),
    );
    return groupBy(filtered, 'facet');
  }, [values, searchQuery]);

  const toggleValue = (valueId: string) => {
    const newIds = selectedValueIds.includes(valueId)
      ? selectedValueIds.filter((id) => id !== valueId)
      : [...selectedValueIds, valueId];
    onChange(newIds);
  };

  const selectAll = () => onChange(values.map((v) => v.id));
  const clearAll = () => onChange([]);

  return (
    <div className="space-y-4">
      {/* Search */}
      <Input
        placeholder="Buscar valores..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        disabled={disabled}
      />

      {/* Bulk actions */}
      <div className="flex gap-2 items-center">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={selectAll}
          disabled={disabled}
        >
          Seleccionar todos
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={clearAll}
          disabled={disabled}
        >
          Limpiar
        </Button>
        <Badge variant="secondary">{selectedValueIds.length} seleccionados</Badge>
      </div>

      {/* Grouped checkboxes with proper accessibility (F10) */}
      <div className="max-h-64 overflow-y-auto space-y-4 border rounded-md p-4">
        {Object.entries(groupedValues).length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            {searchQuery ? 'No se encontraron valores.' : 'No hay valores disponibles.'}
          </p>
        ) : (
          Object.entries(groupedValues).map(([facet, facetValues]) => {
            const facetConfig = FACET_CONFIG[facet as keyof typeof FACET_CONFIG];
            if (!facetConfig) return null;

            return (
              <div key={facet}>
                <h4 className="font-medium text-sm mb-2">
                  {facetConfig.icon} {facetConfig.label}
                </h4>
                {facetValues.map((value) => (
                  <label
                    key={value.id}
                    htmlFor={`value-checkbox-${value.id}`}
                    className={`flex items-center gap-2 p-2 hover:bg-gray-50 rounded cursor-pointer ${
                      disabled ? 'opacity-50 cursor-not-allowed' : ''
                    }`}
                  >
                    <input
                      id={`value-checkbox-${value.id}`}
                      type="checkbox"
                      checked={selectedValueIds.includes(value.id)}
                      onChange={() => toggleValue(value.id)}
                      disabled={disabled}
                      className="rounded border-gray-300 text-purple-600 focus:ring-purple-500"
                    />
                    <span className="text-sm">{value.content}</span>
                  </label>
                ))}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
