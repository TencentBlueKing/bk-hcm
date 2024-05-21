import { defineComponent, watch } from 'vue';
// import hooks
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
// import stores
import { useBusinessStore, useLoadBalancerStore } from '@/store';
import { APPLICATION_LAYER_LIST } from '@/constants';
import './index.scss';

export default defineComponent({
  name: 'ListenerList',
  setup() {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const businessStore = useBusinessStore();
    const { columns, settings } = useColumns('targetGroupListener');
    // const searchData = [
    //   {
    //     name: '关联的URL',
    //     id: 'url',
    //   },
    // ];

    /**
     * 异步请求端口健康信息
     */
    const asyncGetTargetsHealth = async (dataList: any) => {
      const cloud_lb_ids = dataList.map(({ cloud_lb_id }: any) => cloud_lb_id);
      if (cloud_lb_ids.length === 0) return;
      // 查询指定的目标组绑定的负载均衡下的端口健康信息
      const res = await businessStore.asyncGetTargetsHealth(loadBalancerStore.targetGroupId, {
        cloud_lb_ids,
      });
      /*
        构建映射关系:
        1. protocol 如果为 "HTTP"/"HTTPS" 用 cloud_lb_id+cloud_rule_id 作为 key, cloud_rule_id 同级的 health_check 作为 value
        2. protocol 为其他值, 用 cloud_lb_id+cloud_lbl_id 作为key, cloud_lbl_id 同级的 health_check 作为 value
      */
      const healthCheckMap = {};
      res.data.details.forEach(({ cloud_lb_id, listeners }: any) => {
        listeners.forEach((listener: any) => {
          const { protocol, cloud_lbl_id, health_check } = listener;
          if (APPLICATION_LAYER_LIST.includes(protocol)) {
            // 七层
            const { rules } = listener;
            // 如果rules为null, 则表明监听器没有绑定rs, 没有端口数据
            rules?.forEach(({ cloud_rule_id, health_check }: any) => {
              healthCheckMap[`${cloud_lb_id}|${cloud_rule_id}`] = health_check;
            });
          } else {
            // 四层
            healthCheckMap[`${cloud_lb_id}|${cloud_lbl_id}`] = health_check;
          }
        });
      });
      // 根据映射关系进行匹配, 将 healthCheck 添加到 dataList 中并返回
      return dataList.map((data: any) => {
        const { cloud_lb_id, cloud_id } = data;
        const healthCheck = healthCheckMap[`${cloud_lb_id}|${cloud_id}`];
        if (healthCheck) {
          return { ...data, healthCheck };
        }
        return { ...data, healthCheck: null };
      });
    };

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
        resolveDataListCb(dataList: any[]) {
          return asyncGetTargetsHealth(dataList);
        },
      },
    });

    watch(
      () => loadBalancerStore.targetGroupId,
      (val) => {
        if (!val) return;
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
