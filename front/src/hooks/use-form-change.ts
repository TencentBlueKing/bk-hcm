import { type MaybeRefOrGetter } from 'vue';
import { useChangeTracker } from './use-change-tracker';

type FieldComparators<T extends Record<string, any>> = Partial<{
  [K in keyof T]: (current: T[K], initial: T[K]) => boolean;
}>;

type FieldNormalizers<T extends Record<string, any>> = Partial<{
  [K in keyof T]: (value: T[K]) => unknown;
}>;

interface IUseFormChangeOptions<T extends Record<string, any>> {
  comparators?: FieldComparators<T>;
  normalizers?: FieldNormalizers<T>;
}

/**
 * 面向表单场景的变更追踪封装。
 * 基于 useChangeTracker 提供更直观的表单语义字段命名。
 */
export function useFormChange<T extends Record<string, any>>(
  source: MaybeRefOrGetter<T | null | undefined>,
  options: IUseFormChangeOptions<T> = {},
) {
  const { currentData, hasChanges, changeSet, stopWatch } = useChangeTracker(source, options);

  return {
    formData: currentData,
    isChanged: hasChanges,
    changedFormData: changeSet,
    stopWatch,
  };
}
