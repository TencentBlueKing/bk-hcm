import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import { Loading, Table } from 'bkui-vue';
import Empty from '@/components/empty';
import useRemoteTable from '@/hooks/useRemoteTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { QueryRuleOPEnum } from '@/typings';

export default defineComponent({
  name: 'SecurityCvmTable',
  setup() {
    const route = useRoute();
    const { getBusinessApiPath } = useWhereAmI();

    const { isLoading, dataList, pagination, handlePageLimitChange, handlePageValueChange, handleSort } =
      useRemoteTable(() => `/api/v1/cloud/${getBusinessApiPath()}security_group/${route.query.id}/cvm/list`, {
        rules: [{ field: 'security_group_id', op: QueryRuleOPEnum.EQ, value: route.query.id }],
        extApi: {
          url: () => `/api/v1/cloud/${getBusinessApiPath()}cvms/list`,
          rules: (dataList) => [{ field: 'id', value: dataList.map((item) => item.cvm_id), op: QueryRuleOPEnum.IN }],
        },
        immediate: true,
      });
    const { columns, settings } = useColumns('securityCvm');

    return () => (
      <Loading loading={isLoading.value} class='security-cvm-table has-selection'>
        <Table
          data={dataList.value}
          rowKey='id'
          columns={columns}
          settings={settings.value}
          pagination={pagination}
          remotePagination
          showOverflowTooltip
          onPageLimitChange={handlePageLimitChange}
          onPageValueChange={handlePageValueChange}
          onColumnSort={handleSort}>
          {{
            empty: () => {
              if (isLoading.value) return null;
              return <Empty />;
            },
          }}
        </Table>
      </Loading>
    );
  },
});
