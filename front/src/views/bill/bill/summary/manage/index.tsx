import { defineComponent, ref } from 'vue';

import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import PrimaryAccount from '../primary';
import SubAccount from '../sub';
import OperationProduct from '../product';

import { useI18n } from 'vue-i18n';

export default defineComponent({
  name: 'BillSummaryManage',
  setup() {
    const { t } = useI18n();

    const activeTab = ref('primary');
    const tabs = ref([
      { name: 'primary', label: t('一级账号'), Component: PrimaryAccount },
      { name: 'sub', label: t('二级账号'), Component: SubAccount },
      { name: 'product', label: t('运营产品'), Component: OperationProduct },
    ]);

    return () => (
      <Tab v-model:active={activeTab.value} type='card-grid'>
        {tabs.value.map(({ name, label, Component }) => (
          <BkTabPanel key={name} name={name} label={label} renderDirective='if'>
            <Component />
          </BkTabPanel>
        ))}
      </Tab>
    );
  },
});
