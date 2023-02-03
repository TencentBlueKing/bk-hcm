import {
  onMounted,
  ref,
} from 'vue';

import {
  useResourceStore,
} from '@/store/resource';
import { CloudType } from '@/typings';

export default (type: string, id: string) => {
  const loading = ref(false);
  const detail = ref({});
  const resourceStore = useResourceStore();

  // 从接口获取数据，并拼装需要的信息
  const getDetail = async () => {
    loading.value = true;
    try {
      const { data } = await resourceStore.detail(type, id);
      data.vendorName = CloudType[data.vendor];
      data.bk_biz_id = data.bk_biz_id === -1 ? '全部' : data.bk_biz_id;
      detail.value = {
        ...data,
        ...data.spec,
        ...data.attachment,
        ...data.revision,
      };
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  onMounted(getDetail);

  return {
    loading,
    detail,
  };
};
