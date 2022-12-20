/**
 * 删除相关事件和状态
 */
import {
  ref,
  h,
} from 'vue';

import DeleteResource from '../children/dialog/delete-resource/delete-resource';

import {
  useResourceStore,
} from '@/store/resource';

export default (
  columns: any[],
  data: any[],
  type: string,
  title: string,
) => {
  const resourceStore = useResourceStore();

  const isShow = ref(false);
  const isDeleting = ref(false);

  // 展示删除弹框
  const handleShowDelete = () => {
    isShow.value = true;
  };

  // 关闭删除弹框
  const handleClose = () => {
    isShow.value = false;
  };

  // 删除数据
  const handleDelete = () => {
    isDeleting.value = true;
    resourceStore
      .delete(type, 123)
      .then(() => {
        isShow.value = false;
      })
      .finally(() => {
        isDeleting.value = false;
      });
  };

  return {
    handleShowDelete,
    DeleteDialog: (_: any, { slots }: any) => {
      return h(
        DeleteResource,
        {
          isShow: isShow.value,
          isDeleting: isDeleting.value,
          title,
          columns,
          data,
          onConfirm: handleDelete,
          onClose: handleClose,
        },
        {
          default: slots.default ?? slots.default(),
        },
      );
    },
  };
};
