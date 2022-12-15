import {
  onMounted,
  ref,
} from 'vue';

import {
  useResourceStore,
} from '@/store/resource';

type Field = {
  name: string;
  prop: string | number;
  link?: string;
  copy?: boolean;
  edit?: boolean;
};

export default (type: string, id: string, fields: Field[]) => {
  const loading = ref(false);
  const detail = ref([]);
  const resourceStore = useResourceStore();

  // 从接口获取数据，并拼装需要的信息
  const getDetail = () => {
    loading.value = true;
    resourceStore
      .detail(type, id)
      .then(({ data = {} }: { data: any }) => {
        const plainData = {
          ...data,
          ...data.spec,
          ...data.attachment,
          ...data.revision,
        };
        detail.value = fields.map((field) => {
          return {
            ...field,
            value: plainData[field.prop],
          };
        });
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
