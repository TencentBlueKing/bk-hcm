import { defineComponent, ref } from 'vue';

import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { getTabs } from './load-tabs.plugin';

export default defineComponent({
  name: 'BillSummaryManage',
  setup() {
    const activeTab = ref('primary');
    const tabs = getTabs();

    return () => (
      <Tab v-model:active={activeTab.value} type='card-grid'>
        {tabs.map(({ name, label, Component }) => (
          <BkTabPanel key={name} name={name} label={label} renderDirective='if'>
            <Component />
          </BkTabPanel>
        ))}
      </Tab>
    );
  },
});
