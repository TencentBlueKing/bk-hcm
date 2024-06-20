import { defineComponent, ref } from 'vue';

import { Button } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import Search from '../components/search';
import CreateAdjustSideSlider from './create';
import Amount from '../components/amount';

import { useI18n } from 'vue-i18n';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

export default defineComponent({
  name: 'BillAdjust',
  setup() {
    const { t } = useI18n();
    const createAdjustSideSliderRef = ref();

    const { columns, settings } = useColumns('billsRootAccountSummary');
    const { CommonTable } = useTable({
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
        apiMethod: null,
      },
    });

    return () => (
      <div class='bill-adjust-module'>
        <Panel>
          <Search style={{ padding: 0, boxShadow: 'none' }} />
        </Panel>
        <Panel class='mt12'>
          <CommonTable>
            {{
              operation: () => (
                <>
                  <Button onClick={() => createAdjustSideSliderRef.value.triggerShow(true)}>
                    <Plus style={{ fontSize: '22px' }} />
                    {t('新增调账')}
                  </Button>
                  <Button>{t('导入')}</Button>
                  <Button>{t('批量删除')}</Button>
                </>
              ),
              operationBarEnd: () => <Amount isAdjust />,
            }}
          </CommonTable>
        </Panel>
        <CreateAdjustSideSlider ref={createAdjustSideSliderRef} />
      </div>
    );
  },
});
