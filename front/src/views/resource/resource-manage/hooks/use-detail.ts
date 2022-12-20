import {
  onMounted,
  ref,
} from 'vue';

import {
  useResourceStore,
} from '@/store/resource';

export default (type: string, id: string) => {
  const loading = ref(false);
  const detail = ref({});
  const resourceStore = useResourceStore();

  // 从接口获取数据，并拼装需要的信息
  const getDetail = () => {
    loading.value = true;
    resourceStore
      .detail(type, id)
      .then(({ data = {} }: { data: any }) => {
        detail.value = {
          ...data,
          ...data.spec,
          ...data.attachment,
          ...data.revision,
        };
      })
      .finally(() => {
        loading.value = false;
      });
  };

  onMounted(getDetail);

  return {
    loading,
    detail,
  };
};
