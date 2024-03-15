import { defineComponent, watch } from 'vue';
// import hooks
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import './index.scss';

export default defineComponent({
  name: 'ListenerList',
  setup() {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const { columns, settings } = useColumns('targetGroupListener');
    // const searchData = [
    //   {
    //     name: '关联的URL',
    //     id: 'url',
    //   },
    // ];

    const { CommonTable, getListData } = useTable({
      searchOptions: {
        disabled: true,
      },
      tableOptions: {
        columns,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: `vendors/tcloud/target_groups/${loadBalancerStore.targetGroupId}/rules`,
      },
    });

    watch(
      () => loadBalancerStore.targetGroupId,
      (val) => {
        getListData([], `vendors/tcloud/target_groups/${val}/rules`);
      },
    );

    return () => (
      <div class='listener-list-page'>
        <CommonTable></CommonTable>
      </div>
    );
  },
});
