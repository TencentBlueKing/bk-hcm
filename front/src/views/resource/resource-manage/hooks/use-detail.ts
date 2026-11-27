import { onBeforeMount, ref } from 'vue';
import type { Ref } from 'vue';
import { VendorEnum, VendorMap } from '@/common/constant';
import type { ResourceDetailLabels, ResourceDetailMap } from '@/typings/resource-detail';
import { useResourceStore } from '@/store/resource';
import { useBusinessMapStore } from '@/store/useBusinessMap';

interface DetailEndpoints extends ResourceDetailMap {
  'vendors/gcp/firewalls/rules': any;
}

type Detail<T extends keyof DetailEndpoints> = DetailEndpoints[T] & ResourceDetailLabels;

interface ResourceMetadata {
  vendor?: VendorEnum;
  bk_biz_id?: number;
}

export const useResourceDetail = <T extends keyof DetailEndpoints>(
  type: T,
  id: string,
  onLoaded?: (detail: Detail<T>) => void,
  vendor?: string,
) => {
  const loading = ref(false);
  const detail = ref({}) as Ref<Detail<T>>;
  const resourceStore = useResourceStore();
  const { getNameFromBusinessMap } = useBusinessMapStore();

  // 从接口获取数据，并拼装需要的信息
  const getDetail = async () => {
    loading.value = true;
    const request: Promise<{ data?: ResourceMetadata }> = resourceStore.detail(type, id, vendor);
    request
      .then(async ({ data = {} }) => {
        const { vendor: resourceVendor, bk_biz_id: bizId } = data;
        const labels: ResourceDetailLabels = {
          vendorName: VendorMap[resourceVendor],
          bk_biz_id_label: bizId === -1 ? '未分配' : bizId,
          bk_biz_id_name: '',
        };
        detail.value = { ...data, ...labels } as Detail<T>;
        detail.value.bk_biz_id_name = await getNameFromBusinessMap(bizId);
        onLoaded?.(detail.value);
        resourceStore.setVendorOfCurrentResource(resourceVendor);
      })
      .finally(() => {
        loading.value = false;
      });
  };

  onBeforeMount(getDetail);

  return {
    loading,
    detail,
    getDetail,
  };
};
