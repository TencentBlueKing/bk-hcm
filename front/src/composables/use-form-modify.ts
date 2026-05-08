import { ref, watch, computed, type Ref } from 'vue';
import { isEqual, cloneDeep } from 'lodash';

/**
 * 表单修改检测 composable
 * 检测表单字段变化，仅提交被修改的字段
 * @returns { formData, isModified, changedFields, initForm, markClean }
 */
export function useFormModify<T extends Record<string, any>>() {
  const formData = ref<T | null>(null);
  const initialSnapshot = ref<T | null>(null);
  const isModified = ref(false);

  // 初始化表单数据和基准值
  const initForm = (data: T) => {
    formData.value = cloneDeep(data);
    initialSnapshot.value = cloneDeep(data);
    isModified.value = false;
  };

  // 获取被修改的字段
  const getChangedFields = (): Partial<T> => {
    if (!initialSnapshot.value || !formData.value) return {};

    const changed: Partial<T> = {};
    const initial = initialSnapshot.value;
    const current = formData.value;

    for (const key of Object.keys(current)) {
      if (!isEqual((initial as any)[key], (current as any)[key])) {
        (changed as any)[key] = (current as any)[key];
      }
    }

    return changed;
  };

  // 计算属性：被修改的字段
  const changedFields = computed<Partial<T>>(() => getChangedFields());

  // 监听表单数据变化
  watch(
    formData,
    () => {
      if (!initialSnapshot.value || !formData.value) return;
      isModified.value = !isEqual(formData.value, initialSnapshot.value);
    },
    { deep: true },
  );

  // 将当前状态标记为干净（保存成功后调用）
  const markClean = () => {
    if (formData.value) {
      initialSnapshot.value = cloneDeep(formData.value);
      isModified.value = false;
    }
  };

  return {
    formData: formData as Ref<T>,
    isModified,
    changedFields,
    initForm,
    markClean,
  };
}
