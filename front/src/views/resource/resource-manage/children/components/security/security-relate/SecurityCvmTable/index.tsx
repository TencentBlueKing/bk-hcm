import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { QueryRuleOPEnum } from '@/typings';
import RemoteTable from '@/components/RemoteTable';

export default defineComponent({
  name: 'SecurityCvmTable',
  setup() {
    const route = useRoute();
    const { getBusinessApiPath } = useWhereAmI();

    return () => (
      <RemoteTable
        columnName='cvms'
        noSort={true}
        apis={[
          {
            url: () => `/api/v1/cloud/${getBusinessApiPath()}security_group/${route.query.id}/cvm/list`,
            payload: {},
            rules: [{ field: 'security_group_id', op: QueryRuleOPEnum.EQ, value: route.query.id }],
          },
          {
            url: () => `/api/v1/cloud/${getBusinessApiPath()}cvms/list`,
            payload: {},
            rules: (dataList) => [{ field: 'id', value: dataList.map((item) => item.cvm_id), op: QueryRuleOPEnum.IN }],
            reject: (dataList) => dataList.length === 0,
          },
        ]}
      />
    );
  },
});
