import { defineComponent } from 'vue';
import { Button } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import './index.scss';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'AllClbsManager',
  setup() {
    const { columns, settings } = useColumns('clbs');
    const searchData: any = [];
    const searchUrl = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`;
    const { CommonTable } = useTable({
      columns,
      settings: settings.value,
      searchData,
      searchUrl,
    });

    return () => (
      <CommonTable>
        {{
          operation: () => (
            <>
              <Button theme='primary'>购买</Button>
              <Button>批量删除</Button>
            </>
          ),
        }}
      </CommonTable>
    );
  },
});