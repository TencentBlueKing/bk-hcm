import { defineComponent, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
// import components
import { Tag } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
import RsConfigTable from '../../components/RsConfigTable';
// import stores
import { useAccountStore, useLoadBalancerStore } from '@/store';
// import utils
import bus from '@/common/bus';
import './index.scss';

export default defineComponent({
  name: 'BatchAddRsSideslider',
  setup() {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const accountStore = useAccountStore();

    const isShow = ref(false);
    const accountId = ref(''); // 当前选中目标组同属的账号id
    const selectedTargetGroups = ref([]); // 当前选中目标组的信息

    const getDefaultFormData = () => ({
      bk_biz_id: accountStore.bizs,
      target_group_id: '',
      targets: [] as any[],
    });
    const clearFormData = () => {
      Object.assign(formData, getDefaultFormData());
    };
    const formData = reactive(getDefaultFormData());

    const handleShow = (account_id: string) => {
      isShow.value = true;
      accountId.value = account_id;
    };
    // submit-handler
    const handleSubmit = () => {};

    watch(isShow, (val) => {
      if (!val) clearFormData();
    });

    watch(
      () => loadBalancerStore.selectedRsList,
      (val) => {
        formData.targets = val;
      },
      {
        deep: true,
      },
    );

    onMounted(() => {
      bus.$on('showBatchAddRsSideslider', handleShow);
      bus.$on('setTargetGroups', (list: any[]) => (selectedTargetGroups.value = list));
    });

    onUnmounted(() => {
      bus.$off('showBatchAddRsSideslider');
      bus.$off('setTargetGroups');
    });

    return () => (
      <CommonSideslider title='批量添加 RS' width={960} v-model:isShow={isShow.value} onHandleSubmit={handleSubmit}>
        <div class='rs-sideslider-content'>
          <div class='selected-target-groups'>
            <span class='label'>已选择目标组</span>
            <div>
              {selectedTargetGroups.value.map(({ id, name }) => (
                <Tag key={id}>{name}</Tag>
              ))}
            </div>
          </div>
          <RsConfigTable v-model:rsList={formData.targets} accountId={accountId.value} noDisabled={true} />
        </div>
      </CommonSideslider>
    );
  },
});
