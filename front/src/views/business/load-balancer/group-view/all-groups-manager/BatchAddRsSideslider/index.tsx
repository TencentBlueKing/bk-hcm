import { defineComponent, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
// import components
import { Message, Tag } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
import RsConfigTable from '../../components/RsConfigTable';
// import stores
import { useBusinessStore } from '@/store';
// import utils
import bus from '@/common/bus';
import './index.scss';

export default defineComponent({
  name: 'BatchAddRsSideslider',
  setup() {
    // use stores
    const businessStore = useBusinessStore();

    const isShow = ref(false);
    const isSubmitDisabled = ref(false);
    const accountId = ref(''); // 当前选中目标组同属的账号id
    const selectedTargetGroups = ref([]); // 当前选中目标组的信息

    const getDefaultFormData = () => ({
      targets: [] as any[],
    });
    const clearFormData = () => {
      Object.assign(formData, getDefaultFormData());
    };
    const formData = reactive(getDefaultFormData());

    const handleShow = ({ account_id, selectedRsList }: { account_id: string; selectedRsList: any[] }) => {
      isShow.value = true;
      // 更新account_id
      accountId.value = account_id;
      // 更新 targets
      formData.targets = [
        ...formData.targets,
        ...selectedRsList.reduce((prev, curr) => {
          // 已添加过的rs ip, 不允许重复添加
          if (!formData.targets.find((item) => item.inst_id === curr.id || item.id === curr.id)) {
            prev.push(curr);
          }
          return prev;
        }, []),
      ];
    };

    // 处理参数
    const resolveFormData = () => ({
      targets: formData.targets.map(({ cloud_id, port, weight }) => ({
        inst_type: 'CVM',
        cloud_inst_id: cloud_id,
        port,
        weight,
      })),
    });
    // submit-handler
    const handleSubmit = async () => {
      const data = resolveFormData();
      // 遍历当前选中的目标组, 为每个目标组添加rs
      try {
        isSubmitDisabled.value = true;
        const requestList = selectedTargetGroups.value.map(({ id }) => businessStore.addRsToTargetGroup(id, data));
        const results = await Promise.allSettled(requestList);
        results.forEach(({ status }, index) => {
          if (status === 'fulfilled') {
            Message({ theme: 'success', message: `目标组【${selectedTargetGroups.value[index].name}】批量添加rs成功` });
          }
        });
        isShow.value = false;
      } finally {
        isSubmitDisabled.value = false;
      }
    };

    watch(isShow, (val) => {
      if (!val) clearFormData();
    });

    onMounted(() => {
      bus.$on('showBatchAddRsSideslider', handleShow);
      bus.$on('setTargetGroups', (list: any[]) => (selectedTargetGroups.value = list));
    });

    onUnmounted(() => {
      bus.$off('showBatchAddRsSideslider');
      bus.$off('setTargetGroups');
    });

    return () => (
      <CommonSideslider
        title='批量添加 RS'
        width={960}
        v-model:isShow={isShow.value}
        onHandleSubmit={handleSubmit}
        isSubmitDisabled={isSubmitDisabled.value}>
        <div class='rs-sideslider-content'>
          <div class='selected-target-groups'>
            <span class='label'>已选择目标组</span>
            <div class='tags'>
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
