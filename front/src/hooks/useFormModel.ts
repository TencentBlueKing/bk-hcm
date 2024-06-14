import { reactive, UnwrapRef } from 'vue';

function useFormModel<T extends object>(initialState: T) {
  const formModel = reactive({ ...initialState }) as UnwrapRef<T>;

  function resetForm() {
    Object.assign(formModel, initialState);
  }

  return {
    formModel,
    resetForm,
  };
}

export default useFormModel;
