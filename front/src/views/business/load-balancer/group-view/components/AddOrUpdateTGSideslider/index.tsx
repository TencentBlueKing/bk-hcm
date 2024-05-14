import { defineComponent, onMounted, onUnmounted, ref, PropType, reactive } from 'vue';
// import components
import { Form, Message } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
// import stores
import { useAccountStore, useBusinessStore, useLoadBalancerStore } from '@/store';
// import custom hooks
import useAddOrUpdateTGForm from './useAddOrUpdateTGForm';
import useChangeScene from './useChangeScene';
// import utils
import bus from '@/common/bus';

const { FormItem } = Form;

export default defineComponent({
  name: 'AddOrUpdateTGSideslider',
  props: {
    origin: String as PropType<'list' | 'info'>,
    getListData: Function as PropType<(...args: any) => any>,
    getTargetGroupDetail: Function as PropType<(...args: any) => any>,
  },
  setup(props) {
    // use stores
    const accountStore = useAccountStore();
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();

    const isShow = ref(false);
    const isSubmitDisabled = ref(false);
    const isEdit = ref(false);

    // 表单相关
    const getDefaultFormData = () => ({
      id: '',
      bk_biz_id: accountStore.bizs,
      account_id: '',
      name: '',
      protocol: '',
      port: '',
      region: '',
      cloud_vpc_id: '',
      rs_list: [] as any[],
    });
    const clearFormData = () => {
      Object.assign(formData, getDefaultFormData());
    };
    const formData = reactive(getDefaultFormData());
    const { updateCount } = useChangeScene(isShow, formData);
    const { formItemOptions, canUpdateRegionOrVpc } = useAddOrUpdateTGForm(formData, updateCount, isEdit);

    // click-handler - 新建目标组
    const handleAddTargetGroup = () => {
      clearFormData();
      loadBalancerStore.setCurrentScene('add');
      isShow.value = true;
      isEdit.value = false;
    };

    // click-handler - 编辑目标组
    const handleEditTargetGroup = (data: any) => {
      clearFormData();
      Object.assign(formData, data);
      // 初始化场景值
      loadBalancerStore.setCurrentScene(null);
      isShow.value = true;
      isEdit.value = true;
    };

    // 处理参数 - add
    const resolveFormDataForAdd = () => ({
      bk_biz_id: formData.bk_biz_id,
      account_id: formData.account_id,
      name: formData.name,
      protocol: formData.protocol,
      port: +formData.port,
      region: formData.region,
      cloud_vpc_id: formData.cloud_vpc_id,
      rs_list:
        formData.rs_list.length > 0
          ? formData.rs_list.map(({ cloud_id, port, weight }) => ({
              inst_type: 'CVM',
              cloud_inst_id: cloud_id,
              port,
              weight,
            }))
          : undefined,
    });
    // 处理参数 - edit
    const resolveFormDataForEdit = () => ({
      id: formData.id,
      bk_biz_id: formData.bk_biz_id,
      account_id: formData.account_id,
      name: formData.name,
      protocol: formData.protocol,
      port: +formData.port,
      region: canUpdateRegionOrVpc.value ? formData.region : undefined,
      cloud_vpc_id: canUpdateRegionOrVpc.value ? formData.cloud_vpc_id : undefined,
    });
    // 处理参数 - 添加rs
    const resolveFormDataForAddRs = () => ({
      account_id: formData.account_id,
      target_groups: [
        {
          target_group_id: formData.id,
          targets: formData.rs_list
            // 只提交新增的rs
            .filter(({ isNew }) => isNew)
            .map(({ cloud_id, port, weight }) => ({
              inst_type: 'CVM',
              cloud_inst_id: cloud_id,
              port,
              weight,
            })),
        },
      ],
    });
    // 处理参数 - 批量修改端口/权重
    const resolveFormDataForBatchUpdate = (type: 'port' | 'weight') => ({
      target_ids: formData.rs_list.map(({ id }) => id),
      [`new_${type}`]: formData.rs_list[0][type],
    });

    // submit - [新增/编辑目标组] 或 [批量添加rs] 或 [批量修改端口] 或 [批量修改权重]
    const handleAddOrUpdateTargetGroupSubmit = async () => {
      let promise;
      let message;
      switch (loadBalancerStore.currentScene) {
        case 'add':
          promise = businessStore.createTargetGroups(resolveFormDataForAdd());
          message = '新建成功';
          break;
        case 'edit':
          promise = businessStore.editTargetGroups(resolveFormDataForEdit());
          message = '编辑成功';
          break;
        case 'AddRs':
          promise = businessStore.batchAddTargets(resolveFormDataForAddRs());
          message = 'RS添加成功';
          break;
        case 'port':
          promise = businessStore.batchUpdateRsPort(formData.id, resolveFormDataForBatchUpdate('port'));
          message = '批量修改端口成功';
          break;
        case 'weight':
          promise = businessStore.batchUpdateRsWeight(formData.id, resolveFormDataForBatchUpdate('weight'));
          message = '批量修改权重成功';
          break;
      }
      try {
        isSubmitDisabled.value = true;
        await promise;
        Message({ message, theme: 'success' });
        isShow.value = false;

        // 如果组件用于list页面, 则重新请求list接口; 如果组件用于info页面, 则重新请求detail接口
        if (props.origin === 'list') {
          // 表格目标组list
          props.getListData();
        } else {
          props.getTargetGroupDetail(formData.id);
        }
        // 刷新左侧目标组list
        bus.$emit('refreshTargetGroupList');
      } finally {
        isSubmitDisabled.value = false;
      }
    };

    // 更新rsConfigTable中显示的rsList
    const handleUpdateSelectedRsList = (selectedRsList: any[]) => {
      formData.rs_list = [
        ...formData.rs_list,
        ...selectedRsList.reduce((prev, curr) => {
          if (!formData.rs_list.find((item) => item.inst_id === curr.id || item.id === curr.id)) {
            prev.push(curr);
          }
          return prev;
        }, []),
      ];
    };

    onMounted(() => {
      bus.$on('addTargetGroup', handleAddTargetGroup);
      bus.$on('editTargetGroup', handleEditTargetGroup);
      bus.$on('updateSelectedRsList', handleUpdateSelectedRsList);
    });

    onUnmounted(() => {
      bus.$off('addTargetGroup');
      bus.$off('editTargetGroup');
      bus.$off('updateSelectedRsList');
    });

    return () => (
      <CommonSideslider
        title={isEdit.value ? '编辑目标组' : '新建目标组'}
        width={960}
        v-model:isShow={isShow.value}
        isSubmitLoading={isSubmitDisabled.value}
        onHandleSubmit={handleAddOrUpdateTargetGroupSubmit}>
        <bk-container margin={0}>
          <Form formType='vertical' model={formData}>
            {formItemOptions.value.map((item) => (
              <bk-row>
                {Array.isArray(item) ? (
                  item.map((subItem) => (
                    <bk-col span={subItem.span}>
                      <FormItem label={subItem.label} property={subItem.property} required={subItem.required}>
                        {subItem.content()}
                      </FormItem>
                    </bk-col>
                  ))
                ) : (
                  <bk-col span={item.span}>
                    <FormItem label={item.label} property={item.property} required={item.required}>
                      {item.content()}
                    </FormItem>
                  </bk-col>
                )}
              </bk-row>
            ))}
          </Form>
        </bk-container>
      </CommonSideslider>
    );
  },
});
