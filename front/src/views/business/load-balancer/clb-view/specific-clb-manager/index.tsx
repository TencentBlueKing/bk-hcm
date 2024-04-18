import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
import { Message, Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import ListenerList from './listener-list';
import ClbDetail from './clb-detail';
import SecurityGroup from './security-group';
import { useBusinessStore, useLoadBalancerStore } from '@/store';
import useActiveTab from '@/hooks/useActiveTab';
import { debounce } from 'lodash';
import bus from '@/common/bus';
import './index.scss';

export enum TypeEnum {
  list = 'list',
  detail = 'detail',
  security = 'security',
}

export default defineComponent({
  setup() {
    // use stores
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();

    const { activeTab, handleActiveTabChange } = useActiveTab(TypeEnum.list);
    const tabList = [
      {
        name: TypeEnum.list,
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

    const detail: { [key: string]: any } = ref({});
    const getDetails = async (id: string) => {
      const res = await businessStore.getLbDetail(id);
      detail.value = res.data;
    };
    const updateLb = debounce(async (payload: Record<string, any>) => {
      await businessStore.updateLbDetail({
        id: detail.value.id,
        ...payload,
      });
      Message({
        message: '更新成功',
        theme: 'success',
      });
    }, 1000);

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

    onMounted(() => {
      bus.$on('changeSpecificClbActiveTab', handleActiveTabChange);
    });

    onUnmounted(() => {
      bus.$off('changeSpecificClbActiveTab');
    });

    return () => (
      <Tab v-model:active={activeTab.value} type={'card-grid'} onChange={handleActiveTabChange}>
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
