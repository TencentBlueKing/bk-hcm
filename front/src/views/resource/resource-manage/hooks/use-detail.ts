import {
  onBeforeMount,
  ref,
} from 'vue';

import {
  useResourceStore,
} from '@/store/resource';
import { CloudType } from '@/typings';

export default (type: string, id: string, cb?: Function, vendor?: string) => {
  const loading = ref(false);
  const detail = ref<any>({});
  const resourceStore = useResourceStore();

  // 从接口获取数据，并拼装需要的信息
  const getDetail = async () => {
    loading.value = true;
    resourceStore
      .detail(type, id, vendor)
      .then(({ data = {} }: { data: any }) => {
        detail.value = {
          ...data,
          ...data.extension,
          vendorName: CloudType[data.vendor],
          bk_biz_id: data.bk_biz_id === -1 ? '未分配' : data.bk_biz_id,
        };
        cb?.(detail.value);
      })
      .finally(() => {
        loading.value = false;
      });
  };

  onBeforeMount(getDetail);

  return {
    loading,
    detail,
  };
};
