import { onBeforeMount, ref } from 'vue';
import { VendorMap } from '@/common/constant';
import { useResourceStore } from '@/store/resource';
import { useBusinessMapStore } from '@/store/useBusinessMap';

export default (type: string, id: string, cb?: Function, vendor?: string) => {
  const loading = ref(false);
  const detail = ref<any>({});
  const resourceStore = useResourceStore();
  const { getNameFromBusinessMap } = useBusinessMapStore();

  // 从接口获取数据，并拼装需要的信息
  const getDetail = async () => {
    loading.value = true;
    resourceStore
      .detail(type, id, vendor)
      .then(async ({ data = {} }: { data: any }) => {
        detail.value = {
          ...data,
          ...data.extension,
          vendorName: VendorMap[data.vendor],
          bk_biz_id_label: data.bk_biz_id === -1 ? '未分配' : data.bk_biz_id,
          bk_biz_id_name: '',
        };
        detail.value.bk_biz_id_name = await getNameFromBusinessMap(data.bk_biz_id);
        cb?.(detail.value);
        resourceStore.setVendorOfCurrentResource(data.vendor);
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
