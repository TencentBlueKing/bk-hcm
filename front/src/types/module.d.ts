declare module '*.vue' {
  import { defineComponent } from 'vue';
  const Component: ReturnType<typeof defineComponent>;
  export default Component;
}
declare module '*.svg';
declare module '*.png';
declare module '*.module.scss';
declare module 'vue-virtual-scroller' {
  export const RecycleScroller: any;
}
declare module 'hot-formula-parser' {
  export class Parser {
    constructor();
    parse(expression: string): { error: string | null; result: any };
    setVariable(name: string, value: any): Parser;
    getVariable(name: string): any;
    setFunction(name: string, fn: (params: any[]) => any): Parser;
    getFunction(name: string): ((...args: any[]) => any) | undefined;
    on(
      event: 'callCellValue',
      callback: (
        cellCoord: {
          label: string;
          row: { index: number; label: string; isAbsolute: boolean };
          column: { index: number; label: string; isAbsolute: boolean };
        },
        done: (value: any) => void,
      ) => void,
    ): void;
    on(
      event: 'callRangeValue',
      callback: (
        startCellCoord: { label: string; row: { index: number }; column: { index: number } },
        endCellCoord: { label: string; row: { index: number }; column: { index: number } },
        done: (value: any[]) => void,
      ) => void,
    ): void;
    on(event: 'callVariable', callback: (name: string, done: (value: any) => void) => void): void;
    on(event: 'callFunction', callback: (name: string, params: any[], done: (value: any) => void) => void): void;
  }
  export const SUPPORTED_FORMULAS: string[];
  export function columnIndexToLabel(index: number): string;
  export function columnLabelToIndex(label: string): number;
  export function rowIndexToLabel(index: number): string;
  export function rowLabelToIndex(label: string): number;
  export function extractLabel(label: string): any[];
  export function toLabel(row: any, column: any): string;
  export const ERROR: string;
  export const ERROR_DIV_ZERO: string;
  export const ERROR_NAME: string;
  export const ERROR_NOT_AVAILABLE: string;
  export const ERROR_NULL: string;
  export const ERROR_NUM: string;
  export const ERROR_REF: string;
  export const ERROR_VALUE: string;
  export function error(type: string): string | null;
}
