import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { defineComponent } from 'vue';
import { RouterView } from 'vue-router';
import './index.scss';

const ACCOUNT_DETAIL_TABS = [{ key: 'basic', label: '基本信息' }];

export default defineComponent({
  setup() {
    return () => (
      <div class={'page-container account-info-container'}>
        <Tab active={ACCOUNT_DETAIL_TABS[0].key} type='card-grid'>
          {ACCOUNT_DETAIL_TABS.map(({ key, label }) => (
            <BkTabPanel key={key} label={label} name={key as any}>
              <RouterView />
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
