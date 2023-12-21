import { defineComponent } from 'vue';
import { Button } from 'bkui-vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { Plus } from 'bkui-vue/lib/icon';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'DomainList',
  setup() {
    const { columns, settings } = useColumns('domain');
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        render() {
          return (
            <div class='operate-groups'>
              <span>编辑</span>
              <span>删除</span>
            </div>
          );
        },
      },
    ];
    const searchData: any = [];
    const searchUrl = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`;
    const { CommonTable } = useTable({
      columns: tableColumns,
      settings: settings.value,
      searchUrl,
      searchData,
    });

    return () => (
      <CommonTable>
        {{
          operation: () => (
            <>
              <Button theme='primary'>
                <Plus class='f20' />
                新增域名
              </Button>
              <Button>批量删除</Button>
            </>
          ),
        }}
      </CommonTable>
    );
  },
});
