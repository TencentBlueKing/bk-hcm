import { defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { ArrowsLeft } from 'bkui-vue/lib/icon';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import Panel from '@/components/panel';

import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsSyncRecordList } from '@/api/bill';

export default defineComponent({
  name: 'BillSummaryOperationRecord',
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const { columns } = useColumns('billsSummaryOperationRecord');

    const actionTypes = [
      { label: 'sync', text: t('同步') },
      { label: 'confirm', text: t('确认') },
      { label: 'import', text: t('导入') },
    ];
    const activeActionType = ref('sync');

    const { CommonTable } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns,
      },
      requestOption: { apiMethod: reqBillsSyncRecordList },
    });

    return () => (
      <>
        <section class={cssModule.back} onClick={() => router.back()}>
          <ArrowsLeft class={cssModule['back-icon']} />
          <span class={cssModule['back-text']}>{t('返回上一级')}</span>
        </section>
        <Panel class={cssModule.table}>
          <CommonTable>
            {{
              operation: () => (
                <>
                  <span class={cssModule.title}>{t('操作记录')}</span>
                  <BkRadioGroup v-model={activeActionType.value} class={cssModule['action-type']}>
                    {actionTypes.map(({ label, text }) => (
                      <BkRadioButton label={label}>{text}</BkRadioButton>
                    ))}
                  </BkRadioGroup>
                </>
              ),
            }}
          </CommonTable>
        </Panel>
      </>
    );
  },
});
