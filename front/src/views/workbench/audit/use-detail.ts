import { useAuditStore } from '@/store/audit';
import { ref, watchEffect } from 'vue';

export default (props: { id: number; bizId: number }) => {
  const auditStore = useAuditStore();

  const isLoading = ref(false);
  const details = ref({});

  watchEffect(
    void (async () => {
      isLoading.value = true;

      const result = await auditStore.detail(props.id, props.bizId);
      details.value = result?.data;

      isLoading.value = false;
    })(),
  );

  return {
    details,
    isLoading,
  };
};
