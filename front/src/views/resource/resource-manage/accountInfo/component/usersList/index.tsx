import http from '@/http';
import { QueryRuleOPEnum } from '@/typings';
import { Button, SearchSelect, Table } from 'bkui-vue';
import { defineComponent, ref, watch } from 'vue';
import './index.scss';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const searchVal = ref('');
    const usersList = ref([]);
    const pagination = ref({
      current: 1,
      limit: 20,
      count: 0,
    });
    const columns = [
      {
        label: '账号 ID',
        field: 'id',
      },
      {
        label: '所属业务',
        field: 'bk_biz_ids',
        render: ({ data }: any) => data?.bk_biz_ids?.join(',') || '-',
      },
      {
        label: '备注',
        field: 'memo',
        render: ({ cell }: any) => cell || '-',
      },
      {
        label: '负责人',
        field: 'managers',
        render: ({ data }: any) => data?.managers?.join(',') || '-',
      },
      {
        label: '创建人',
        field: 'creator',
      },
      {
        label: '创建时间',
        field: 'created_at',
      },
      {
        label: '操作',
        field: 'operation',
        render: () => (
          <Button text theme='primary'>
            编辑
          </Button>
        ),
      },
    ];
    const searchData = [
      {
        name: '账号 ID',
        id: 'id',
      },
    ];
    watch(
      () => searchVal.value,
      async (val) => {
        const id = val?.[0]?.values?.[0]?.id;
        const res = await http.post(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/sub_accounts/list`,
          {
            page: {
              limit: 10,
              start: 0,
              count: false,
            },
            filter: {
              op: QueryRuleOPEnum.AND,
              rules: !!id
                ? [
                  {
                    field: 'id',
                    op: QueryRuleOPEnum.EQ,
                    value: id,
                  },
                ]
                : [],
            },
          },
        );
        usersList.value = res?.data?.details;
      },
      {
        immediate: true,
      },
    );
    return () => (
      <div>
        <SearchSelect
          class='w500 users-list-search-selector'
          v-model={searchVal.value}
          data={searchData}
        />
        <Table
          data={usersList.value}
          pagination={pagination.value}
          remotePagination
          columns={columns}>
        </Table>
      </div>
    );
  },
});
