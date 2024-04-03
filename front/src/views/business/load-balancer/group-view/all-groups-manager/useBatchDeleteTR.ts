import { Ref, ref } from 'vue';
// import stores
import { useAccountStore, useBusinessStore, useLoadBalancerStore } from '@/store';
// import types
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { Message } from 'bkui-vue';

export default (searchData: ISearchItem[], selections: Ref<any[]>, getListData: (...args: any) => any) => {
  // use stores
  const accountStore = useAccountStore();
  const businessStore = useBusinessStore();
  const loadBalancerStore = useLoadBalancerStore();

  const isSubmitLoading = ref(false);
  const isBatchDeleteTargetGroupShow = ref(false);
  const canDeleteTargetGroup = ref(false);
  const batchDeleteTargetGroupTableProps = {
    data: selections.value,
    columns: [
      {
        label: '目标组名称',
        field: 'name',
      },
      {
        label: '协议',
        field: 'protocol',
        filter: true,
        render({ cell }: any) {
          return cell.trim() || '--';
        },
      },
      {
        label: '端口',
        field: 'port',
        filter: true,
      },
      {
        label: '关联的负载均衡',
        field: 'lb_name',
        render({ cell }: any) {
          return cell.trim() || '--';
        },
      },
      {
        label: '绑定监听器数量',
        field: 'listener_num',
        sort: true,
        align: 'right',
      },
    ],
    searchData,
  };

  // submit-handler
  const batchDeleteTargetGroup = async () => {
    try {
      isSubmitLoading.value = true;
      await businessStore.deleteTargetGroups({
        bk_biz_id: accountStore.bizs,
        ids: selections.value.map(({ id }) => id),
      });
      Message({
        message: '批量删除成功',
        theme: 'success',
      });
      isBatchDeleteTargetGroupShow.value = false;
      loadBalancerStore.getTargetGroupList();
      getListData();
    } finally {
      isSubmitLoading.value = false;
    }
  };

  return {
    isSubmitLoading,
    isBatchDeleteTargetGroupShow,
    canDeleteTargetGroup,
    batchDeleteTargetGroupTableProps,
    batchDeleteTargetGroup,
  };
};
