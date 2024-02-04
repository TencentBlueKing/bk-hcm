import {
  // onMounted,
  ref,
} from 'vue';

import { useResourceStore } from '@/store/resource';

import { Message } from 'bkui-vue';
import i18n from '@/language/i18n';

// import { useI18n } from 'vue-i18n';

export default (type: string, data: any, id?: number) => {
  const { t } = i18n.global;
  const loading = ref(false);
  const resourceStore = useResourceStore();

  // 新增
  const addData = async () => {
    loading.value = true;
    try {
      await resourceStore.add(type, data);
      Message({
        message: t('添加成功'),
        theme: 'success',
      });
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }
  };

  // 更新
  const updateData = async () => {
    loading.value = true;
    try {
      await resourceStore.update(type, data, id);
      Message({
        message: t('编辑成功'),
        theme: 'success',
      });
    } catch (error) {
      console.log(error);
    } finally {
      loading.value = false;
    }

    // resourceStore
    //   .update(type, data, id)
    //   .then(() => {

    //   })
    //   .finally(() => {
    //     loading.value = false;
    //   });
  };

  // onMounted(addData);

  return {
    loading,
    addData,
    updateData,
  };
};
