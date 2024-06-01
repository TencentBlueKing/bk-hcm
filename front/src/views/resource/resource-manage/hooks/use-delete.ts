/**
 * 删除相关事件和状态
 */
import { Ref, ref, h, watch } from 'vue';

import DeleteResource from '../children/dialog/delete-resource/delete-resource';
import i18n from '@/language/i18n';
import { Message } from 'bkui-vue';

import { useResourceStore } from '@/store/resource';

export default (
  columns: any[],
  data: Ref<any[]>,
  type: string,
  title: string,
  isBatch?: boolean,
  operationType: 'delete' | 'recycle' = 'delete',
  onFinishedCallback?: () => void,
) => {
  const resourceStore = useResourceStore();

  const { t } = i18n.global;

  const isShow = ref(false);
  const isDeleting = ref(false);
  const deleteIds = ref(Array<number>);

  // 展示删除弹框
  const handleShowDelete = (value: any) => {
    deleteIds.value = value;
    isShow.value = true;
  };

  // 关闭删除弹框
  const handleClose = () => {
    isShow.value = false;
  };

  watch(
    () => isShow.value,
    () => {
      data.value = data.value.filter((selection: { id: number }) =>
        (deleteIds.value as unknown as Array<number>).includes(selection.id),
      );
    },
  );

  // 删除\回收数据
  const handleDelete = () => {
    isDeleting.value = true;
    let promise;
    switch (operationType) {
      case 'recycle':
        promise = resourceStore.recycled(type, {
          infos: (deleteIds.value as unknown as Array<number>).map((id) => ({ id })),
        });
        break;
      case 'delete':
      default:
        promise = isBatch
          ? resourceStore.deleteBatch(type, { ids: deleteIds.value })
          : resourceStore.delete(type, deleteIds.value as unknown as number | string);
    }
    promise
      .then(() => {
        isShow.value = false;
        Message({
          theme: 'success',
          message: t('操作成功'),
        });
        onFinishedCallback?.(); // 删除数据回调列表接口
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
