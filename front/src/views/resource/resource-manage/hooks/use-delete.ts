/**
 * 删除相关事件和状态
 */
import {
  Ref,
  ref,
  h,
} from 'vue';

import DeleteResource from '../children/dialog/delete-resource/delete-resource';
import i18n from '@/language/i18n';
import {
  Message } from 'bkui-vue';

import {
  useResourceStore,
} from '@/store/resource';

export default (
  columns: any[],
  data: Ref<any[]>,
  type: string,
  title: string,
  isBatch?: boolean,
) => {
  const resourceStore = useResourceStore();

  const { t } = i18n.global;

  const isShow = ref(false);
  const isDeleting = ref(false);
  const deleteId = ref(0);

  // 展示删除弹框
  const handleShowDelete = (value: any) => {
    deleteId.value = value;
    isShow.value = true;
  };

  // 关闭删除弹框
  const handleClose = () => {
    isShow.value = false;
  };

  // 删除数据
  const handleDelete = () => {
    isDeleting.value = true;
    console.log('isBatch', isBatch);
    if (isBatch) {
      resourceStore
        .deleteBatch(type, { ids: deleteId.value })
        .then(() => {
          isShow.value = false;
          Message({
            theme: 'success',
            message: t('删除成功'),
          });
        })
        .catch((err: any) => {
          Message({
            theme: 'error',
            message: err.message || err,
          });
        })
        .finally(() => {
          isDeleting.value = false;
        });
    } else {
      resourceStore
        .delete(type, deleteId.value)
        .then(() => {
          isShow.value = false;
          Message({
            theme: 'success',
            message: t('删除成功'),
          });
        })
        .catch((err: any) => {
          Message({
            theme: 'error',
            message: err.message || err,
          });
        })
        .finally(() => {
          isDeleting.value = false;
        });
    }
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
          data: data.value,
          onConfirm: handleDelete,
          onClose: handleClose,
        },
        {
          default: slots.default ?? slots.default?.(),
        },
      );
    },
  };
};
