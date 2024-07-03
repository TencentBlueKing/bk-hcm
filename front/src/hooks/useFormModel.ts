import { reactive, UnwrapRef } from 'vue';

function useFormModel<T extends object>(initialState: T) {
  const formModel = reactive({ ...initialState }) as UnwrapRef<T>;

  function resetForm() {
    Object.assign(formModel, initialState);
  }

  function setFormValues(values: Partial<T>) {
    Object.assign(formModel, values);
  }

  return {
    formModel,
    resetForm,
    setFormValues,
  };
}

export default useFormModel;
