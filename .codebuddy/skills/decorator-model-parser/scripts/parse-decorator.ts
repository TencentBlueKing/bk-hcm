/**
 * Decorator Model Parser Utilities
 * 
 * This script provides utility functions for parsing TypeScript decorator-based
 * model definitions into structured property arrays.
 * 
 * Usage: Import these functions or use as reference for manual parsing.
 */

import type { ModelPropertyType, ModelProperty, ModelPropertySearch } from '@/model/typings';

/**
 * Column decorator options interface
 */
interface ColumnOptions {
  name?: string;
  index?: number;
  option?: Record<string | number, any>;
  list?: Array<{ [key: string]: any }>;
  meta?: {
    search?: {
      op?: string;
      filterRules?: (value: any) => any;
      format?: (value: any) => any;
      enableEmpty?: boolean;
    };
    column?: {
      sort?: boolean;
      width?: number | string;
      fixed?: 'left' | 'right';
      render?: (args: any) => any;
    };
    display?: {
      appearance?: string;
      format?: (value: any) => any;
      render?: (value: any) => any;
    };
  };
  [key: string]: any;
}

/**
 * Parsed column result
 */
interface ParsedColumn {
  id: string;
  name: string;
  type: ModelPropertyType;
  [key: string]: any;
}

/**
 * Parse a single @Column decorator definition
 * 
 * @param propertyName - The property name in the class
 * @param type - The column type (first argument of @Column)
 * @param options - The column options (second argument of @Column)
 * @returns Parsed column object
 * 
 * @example
 * // Input:
 * // @Column('datetime', { name: '操作时间', index: 0 })
 * // created_at: string;
 * 
 * parseColumnDecorator('created_at', 'datetime', { name: '操作时间', index: 0 })
 * // Returns: { id: 'created_at', name: '操作时间', type: 'datetime', index: 0 }
 */
export function parseColumnDecorator(
  propertyName: string,
  type: ModelPropertyType,
  options: ColumnOptions = {}
): ParsedColumn {
  return {
    id: propertyName,
    name: options.name || propertyName,
    type,
    ...options,
  };
}

/**
 * Parse multiple column definitions into a property array
 * 
 * @param columns - Array of column definitions
 * @returns Sorted array of parsed columns
 * 
 * @example
 * parseColumns([
 *   { propertyName: 'created_at', type: 'datetime', options: { name: '操作时间', index: 0 } },
 *   { propertyName: 'operator', type: 'user', options: { name: '操作人', index: 2 } },
 * ])
 */
export function parseColumns(
  columns: Array<{
    propertyName: string;
    type: ModelPropertyType;
    options?: ColumnOptions;
  }>
): ParsedColumn[] {
  const parsed = columns.map(({ propertyName, type, options }) =>
    parseColumnDecorator(propertyName, type, options)
  );

  // Sort by index if present
  return parsed.sort((a, b) => {
    const indexA = a.index ?? Number.MAX_SAFE_INTEGER;
    const indexB = b.index ?? Number.MAX_SAFE_INTEGER;
    return indexA - indexB;
  });
}

/**
 * Generate a formatted property array string from parsed columns
 * 
 * @param columns - Parsed column array
 * @param format - Output format ('json' | 'typescript')
 * @returns Formatted string representation
 */
export function generatePropertyArrayString(
  columns: ParsedColumn[],
  format: 'json' | 'typescript' = 'typescript'
): string {
  if (format === 'json') {
    return JSON.stringify(columns, null, 2);
  }

  // TypeScript format with proper typing
  const items = columns.map((col) => {
    const entries = Object.entries(col)
      .map(([key, value]) => {
        if (typeof value === 'function') {
          return `${key}: ${value.toString()}`;
        }
        if (typeof value === 'object' && value !== null) {
          return `${key}: ${JSON.stringify(value, null, 2)}`;
        }
        if (typeof value === 'string') {
          return `${key}: '${value}'`;
        }
        return `${key}: ${value}`;
      })
      .join(',\n    ');
    return `  {\n    ${entries}\n  }`;
  });

  return `[\n${items.join(',\n')}\n]`;
}

/**
 * Extract column definitions from a class source code string
 * 
 * This is a simplified parser for demonstration. For production use,
 * consider using TypeScript's compiler API for accurate parsing.
 * 
 * @param sourceCode - The TypeScript source code string
 * @returns Array of extracted column definitions
 */
export function extractColumnsFromSource(sourceCode: string): Array<{
  propertyName: string;
  type: string;
  options: Record<string, any>;
}> {
  const columns: Array<{
    propertyName: string;
    type: string;
    options: Record<string, any>;
  }> = [];

  // Regex to match @Column decorator pattern
  const columnRegex = /@Column\(\s*['"](\w+)['"]\s*(?:,\s*(\{[\s\S]*?\}))?\s*\)\s*(\w+)\s*[;:]/g;

  let match;
  while ((match = columnRegex.exec(sourceCode)) !== null) {
    const [, type, optionsStr, propertyName] = match;
    
    let options: Record<string, any> = {};
    if (optionsStr) {
      // Note: This is a simplified parser, may not handle all cases
      // For complex options, manual parsing or AST analysis is recommended
      try {
        // Extract simple key-value pairs
        const nameMatch = optionsStr.match(/name:\s*['"]([^'"]+)['"]/);
        const indexMatch = optionsStr.match(/index:\s*(\d+)/);
        
        if (nameMatch) options.name = nameMatch[1];
        if (indexMatch) options.index = parseInt(indexMatch[1], 10);
        
        // Check for option reference
        const optionMatch = optionsStr.match(/option:\s*(\w+)/);
        if (optionMatch) options.optionRef = optionMatch[1];
      } catch (e) {
        console.warn(`Failed to parse options for ${propertyName}:`, e);
      }
    }

    columns.push({ propertyName, type, options });
  }

  return columns;
}

/**
 * Convert extracted columns to ModelPropertySearch array format
 * 
 * @param columns - Extracted column definitions
 * @param resolveOption - Function to resolve option references
 * @returns ModelPropertySearch compatible array
 */
export function toModelPropertySearchArray(
  columns: Array<{
    propertyName: string;
    type: string;
    options: Record<string, any>;
  }>,
  resolveOption?: (ref: string) => Record<string, any> | undefined
): ModelPropertySearch[] {
  return columns.map(({ propertyName, type, options }) => {
    const result: ModelPropertySearch = {
      id: propertyName,
      name: options.name || propertyName,
      type: type as ModelPropertyType,
    };

    if (options.index !== undefined) {
      result.index = options.index;
    }

    if (options.optionRef && resolveOption) {
      const resolved = resolveOption(options.optionRef);
      if (resolved) {
        result.option = resolved;
      }
    }

    if (options.meta) {
      result.meta = options.meta;
    }

    return result;
  });
}

// Example usage demonstration
const exampleSource = `
@Model('operation-log/search-condition')
export class SearchCondition {
  @Column('datetime', { name: '操作时间', index: 0 })
  created_at: string;

  @Column('enum', { name: '操作来源', option: OPERATION_LOG_SOURCE_NAME, index: 3 })
  source: OperationLogSource;

  @Column('user', { name: '操作人', index: 3 })
  operator: string;
}
`;

// Uncomment to test:
// const extracted = extractColumnsFromSource(exampleSource);
// console.log(generatePropertyArrayString(parseColumns(extracted)));
