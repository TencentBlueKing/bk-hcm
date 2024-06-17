import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import RemoteTable from '@/components/RemoteTable';
import { QueryRuleOPEnum } from '@/typings';

export default defineComponent({
  name: 'SecurityLbTable',
  setup() {
    const route = useRoute();
    const { getBusinessApiPath } = useWhereAmI();

    return () => (
      <RemoteTable
        columnName='lb'
        noSort={true}
        apis={[
          {
            url: () => `/api/v1/cloud/${getBusinessApiPath()}security_group/${route.query.id}/common/list`,
            payload: { fields: ['res_id'] },
            rules: [
              { field: 'security_group_id', op: QueryRuleOPEnum.EQ, value: route.query.id },
              { field: 'res_type', op: QueryRuleOPEnum.EQ, value: 'load_balancer' },
            ],
          },
          {
            url: () => `/api/v1/cloud/${getBusinessApiPath()}load_balancers/list`,
            payload: { fields: ['res_id'] },
            rules: (dataList) => [{ field: 'id', value: dataList.map((item) => item.res_id), op: QueryRuleOPEnum.IN }],
            reject: (dataList) => dataList.length === 0,
          },
        ]}
      />
    );
  },
});
