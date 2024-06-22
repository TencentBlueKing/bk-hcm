import { defineComponent, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
// import components
import { Form, Message, Tag } from 'bkui-vue';
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
    const account_id = ref(''); // 当前选中目标组同属的账号id
    const vpc_id = ref(''); // 当前选中目标组同属的vpc_id
    const selectedTargetGroups = ref([]); // 当前选中目标组的信息

    const getDefaultFormData = () => ({
      targets: [] as any[],
      rs_list: [] as any[],
    });
    const clearFormData = () => {
      Object.assign(formData, getDefaultFormData());
    };
    const formData = reactive(getDefaultFormData());

    const handleShow = ({
      accountId,
      vpcId,
      selectedRsList,
    }: {
      accountId: string;
      vpcId: string;
      selectedRsList: any[];
    }) => {
      isShow.value = true;
      // 更新account_id
      account_id.value = accountId;
      // 更新vpc_id
      vpc_id.value = vpcId;
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
      // 同步rs_list, 用于form校验
      Object.assign(formData.rs_list, formData.targets);
    };

    // 处理参数
    const resolveFormData = () => {
      return {
        account_id: account_id.value,
        target_groups: selectedTargetGroups.value.map((item) => ({
          target_group_id: item.id,
          targets: formData.targets.map(({ cloud_id, port, weight }) => ({
            inst_type: 'CVM',
            cloud_inst_id: cloud_id,
            port,
            weight,
          })),
        })),
      };
    };
    // submit-handler
    const handleSubmit = async () => {
      try {
        isSubmitDisabled.value = true;
        await businessStore.batchAddTargets(resolveFormData());
        Message({ theme: 'success', message: `批量添加rs成功` });
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
        isSubmitLoading={isSubmitDisabled.value}>
        <div class='rs-sideslider-content'>
          <div class='selected-target-groups'>
            <span class='label'>已选择目标组</span>
            <div class='tags'>
              {selectedTargetGroups.value.map(({ id, name }) => (
                <Tag key={id}>{name}</Tag>
              ))}
            </div>
          </div>
          <Form formType='vertical' model={formData}>
            <RsConfigTable
              v-model:rsList={formData.targets}
              accountId={account_id.value}
              vpcId={vpc_id.value}
              noDisabled={true}
            />
          </Form>
        </div>
      </CommonSideslider>
    );
  },
});
