import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import ListenerList from './listener-list';
import SecurityGroup from './security-group';
import ClbDetail from './clb-detail';
import { Message, Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { useBusinessStore, useLoadBalancerStore } from '@/store';
export enum TypeEnum {
  listener = 'listener',
  detail = 'detail',
  security = 'security',
}

export default defineComponent({
  setup() {
    const activeTab = ref(TypeEnum.listener);
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();
    const detail: { [key: string]: any } = ref({});
    const getDetails = async (id: string) => {
      const res = await businessStore.getLbDetail(id);
      detail.value = res.data;
    };
    const tabList = [
      {
        name: TypeEnum.listener,
        label: '监听器',
        component: ListenerList,
      },
      {
        name: TypeEnum.detail,
        label: '基本信息',
        component: ClbDetail,
      },
      {
        name: TypeEnum.security,
        label: '安全组',
        component: SecurityGroup,
      },
    ];

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      async (val) => {
        const { id, type } = val;
        if (type === 'lb' && id) await getDetails(id);
      },
      {
        immediate: true,
      },
    );

    const updateLb = async (payload: Record<string, any>) => {
      await businessStore.updateLbDetail({
        id: detail.value.id,
        ...payload,
      });
      Message({
        message: '更新成功',
        theme: 'success',
      });
    };
    return () => (
      <Tab v-model:active={activeTab.value} type={'card-grid'}>
        {tabList.map((tab) => (
          <BkTabPanel key={tab.name} name={tab.name} label={tab.label} class={'clb-list-tab-content-container'}>
            <div>
              <tab.component detail={detail.value} getDetails={getDetails} updateLb={updateLb}></tab.component>
            </div>
          </BkTabPanel>
        ))}
      </Tab>
    );
  },
});
