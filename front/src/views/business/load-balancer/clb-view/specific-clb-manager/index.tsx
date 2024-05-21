import { defineComponent, ref } from 'vue';
import './index.scss';
import ListenerList from './listener-list';
import SecurityGroup from './security-group';
import ClbDetail from './clb-detail';
import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
export enum TypeEnum {
  listener = 'listener',
  detail = 'detail',
  security = 'security',
}

export default defineComponent({
  setup() {
    const activeTab = ref(TypeEnum.listener);
    const tabList = [
      {
        name: TypeEnum.listener,
        label: '监听器',
        component: <ListenerList />,
      },
      {
        name: TypeEnum.detail,
        label: '基本信息',
        component: <ClbDetail />,
      },
      {
        name: TypeEnum.security,
        label: '安全组',
        component: <SecurityGroup />,
      },
    ];
    return () => (
      <Tab v-model:active={activeTab.value} type={'card-grid'}>
        {tabList.map((tab) => (
          <BkTabPanel key={tab.name} name={tab.name} label={tab.label} class={'clb-list-tab-content-container'}>
            <div>{tab.component}</div>
          </BkTabPanel>
        ))}
      </Tab>
    );
  },
});
