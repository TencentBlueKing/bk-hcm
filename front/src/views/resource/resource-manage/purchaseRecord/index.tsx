import { useTable } from '@/hooks/useTable/useTable';
import { Button } from 'bkui-vue';
import { defineComponent } from 'vue';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
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
      {
        name: '自定义名称',
        id: 'name',
      },
    ];
    const {
      CommonTable,
    } = useTable({
      columns,
      searchData,
      searchUrl: `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/sub_accounts/list`,
    });
    return () => <div>购买记录
      <CommonTable/>
    </div>;
  },
});
