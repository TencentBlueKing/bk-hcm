import { VendorEnum } from '@/common/constant';
import { SECURITY_RULES_MAP } from '@/constants';
import { ref, watch } from 'vue';

export type SecurityVendorType = Exclude<
  `${VendorEnum}`,
  `${VendorEnum.GCP}` | `${VendorEnum.KAOPU}` | `${VendorEnum.ZENLAYER}`
>;

export const useProtocols = (vendor: SecurityVendorType) => {
  const protocols = ref([]);

  watch(
    () => vendor,
    () => {
      protocols.value = SECURITY_RULES_MAP[vendor]?.map(({ id, name }) => ({
        label: name,
        value: id,
      }));
    },
    {
      immediate: true,
    },
  );

  return {
    protocols,
  };
};
