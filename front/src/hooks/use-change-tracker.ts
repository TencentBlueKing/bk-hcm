import { ref, watch, computed, toValue, onScopeDispose, type Ref, type MaybeRefOrGetter } from 'vue';
import { isEqual, cloneDeep } from 'lodash';

type FieldComparators<T extends Record<string, any>> = Partial<{
  [K in keyof T]: (current: T[K], initial: T[K]) => boolean;
}>;

type FieldNormalizers<T extends Record<string, any>> = Partial<{
  [K in keyof T]: (value: T[K]) => unknown;
}>;

interface IUseChangeTrackerOptions<T extends Record<string, any>> {
  comparators?: FieldComparators<T>;
  normalizers?: FieldNormalizers<T>;
}

/**
 * 通用对象变更追踪。
 * 自动同步数据源到 currentData，并保留 baselineSnapshot 作为对比基线。
 */
export function useChangeTracker<T extends Record<string, any>>(
  source: MaybeRefOrGetter<T | null | undefined>,
  options: IUseChangeTrackerOptions<T> = {},
) {
  const currentData = ref<T | null>(null);
  const baselineSnapshot = ref<T | null>(null);
  const comparators: FieldComparators<T> = options.comparators ?? {};
  const normalizers: FieldNormalizers<T> = options.normalizers ?? {};

  const changeSet = computed<Partial<T>>(() => {
    if (!baselineSnapshot.value || !currentData.value) return {};

    const changed: Partial<T> = {};
    const initial = baselineSnapshot.value;
    const current = currentData.value;

    for (const key of Object.keys(current) as Array<keyof T>) {
      const initialValue = initial[key];
      const currentValue = current[key];

      const comparator = comparators[key];
      const normalizer = normalizers[key];

      const isSame = comparator
        ? comparator(currentValue, initialValue)
        : isEqual(
            normalizer ? normalizer(currentValue) : currentValue,
            normalizer ? normalizer(initialValue) : initialValue,
          );

      if (!isSame) {
        changed[key] = currentValue;
      }
    }

    return changed;
  });

  const stopWatch = watch(
    () => toValue(source),
    (data) => {
      if (!data) {
        currentData.value = null;
        baselineSnapshot.value = null;
        return;
      }

      currentData.value = cloneDeep(data);
      baselineSnapshot.value = cloneDeep(data);
    },
    {
      immediate: true,
      deep: true,
    },
  );

  onScopeDispose(() => {
    stopWatch();
  });

  const hasChanges = computed(() => {
    if (!currentData.value || !baselineSnapshot.value) return false;
    return Object.keys(changeSet.value).length > 0;
  });

  return {
    currentData: currentData as Ref<T | null>,
    hasChanges,
    changeSet,
    stopWatch,
  };
}
