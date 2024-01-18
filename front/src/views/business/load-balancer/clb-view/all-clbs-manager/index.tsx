import { defineComponent } from 'vue';
import { useRouter } from 'vue-router';
import { Button } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import './index.scss';

export default defineComponent({
  name: 'AllClbsManager',
  setup() {
    const router = useRouter();
    const { columns, settings } = useColumns('clbs');
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        width: 120,
        render: () => (<span class='operate-text-btn'>删除</span>),
      },
    ];
    const searchData: any = [];
    const { CommonTable } = useTable({
      columns: tableColumns,
      searchData,
      type: 'clbs',
      tableExtraOptions: {
        settings: settings.value,
      },
    });

    const handleApply = () => {
      router.push({
        path: '/business/service/service-apply/clb',
      });
    };

    return () => (
      <div class='common-card-wrap has-selection'>
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={handleApply}>购买</Button>
                <Button>批量删除</Button>
              </>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
