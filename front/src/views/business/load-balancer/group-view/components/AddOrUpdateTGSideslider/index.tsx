import { defineComponent, onMounted, onUnmounted, ref, PropType, reactive, computed, nextTick } from 'vue';
// import components
import { Alert, Button, Form, Message } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
// import stores
import { useAccountStore, useBusinessStore, useLoadBalancerStore } from '@/store';
// import custom hooks
import useAddOrUpdateTGForm from './useAddOrUpdateTGForm';
import useChangeScene from './useChangeScene';
// import utils
import bus from '@/common/bus';
import { goAsyncTaskDetail } from '@/utils';
import { TG_OPERATION_SCENE_MAP } from '@/constants';

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
    const isSubmitLoading = ref(false);
    const isEdit = ref(false);
    const lastAsyncTaskInfo = reactive({ tgId: '', flowId: '', state: '' });
    const isSubmitDisabled = computed(() => {
      return !['add', 'edit', 'AddRs', 'port', 'weight', 'BatchDeleteRs'].includes(loadBalancerStore.currentScene);
    });
    let timer: any;
    const lbDetail = ref(null);

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
    const { formItemOptions, canUpdateRegionOrVpc, formRef, rules, deletedRsList, regionVpcSelectorRef } =
      useAddOrUpdateTGForm(formData, updateCount, isEdit, lbDetail);

    // click-handler - 新建目标组
    const handleAddTargetGroup = () => {
      clearFormData();
      loadBalancerStore.setCurrentScene('add');
      isShow.value = true;
      isEdit.value = false;
      nextTick(async () => {
        // 侧边栏显示后, 刷新 vpc 列表, 支持编辑的时候默认选中 vpc
        await regionVpcSelectorRef.value.handleRefresh();
        formRef.value.clearValidate();
      });
    };

    const getListenerDetail = async (targetGroup: any) => {
      // 请求绑定的监听器规则
      const rulesRes = await businessStore.list(
        {
          page: { limit: 1, start: 0, count: false },
          filter: { op: 'and', rules: [] },
        },
        `vendors/tcloud/target_groups/${targetGroup.id}/rules`,
      );
      const listenerItem = rulesRes.data.details[0];
      if (!listenerItem) return;
      // 请求监听器详情, 获取端口段信息
      const detailRes = await businessStore.detail('listeners', listenerItem.lbl_id);
      loadBalancerStore.setListenerDetailWithTargetGroup(detailRes.data);
    };
    // click-handler - 编辑目标组
    const handleEditTargetGroup = async (data: any) => {
      clearInterval(timer);
      // 初始化场景值
      loadBalancerStore.setUpdateCount(0);
      loadBalancerStore.setCurrentScene(null);
      // 初始化表单
      clearFormData();
      Object.assign(formData, data);
      isShow.value = true;
      isEdit.value = true;
      // 判断是否有异步任务在执行
      if (lastAsyncTaskInfo.tgId === data.id) {
        // 轮询查询异步任务状态
        loadBalancerStore.setUpdateCount(2);
        timer = setInterval(() => {
          reqAsyncTaskStatus(lastAsyncTaskInfo.tgId, lastAsyncTaskInfo.flowId);
        }, 2000);
      }
      nextTick(() => {
        // 侧边栏显示后, 刷新 vpc 列表, 支持编辑的时候默认选中 vpc
        regionVpcSelectorRef.value.handleRefresh();
      });
      // 请求关联的负载均衡detail, 获取跨域信息
      if (data.lb_id) {
        const res = await businessStore.getLbDetail(data.lb_id);
        lbDetail.value = res.data;
      }
      getListenerDetail(data);
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
            .map(({ cloud_id, port, weight, private_ipv4_addresses }) => {
              return {
                inst_type: lbDetail.value?.extension?.snat_pro ? 'ENI' : 'CVM',
                ip: lbDetail.value?.extension?.snat_pro ? private_ipv4_addresses[0] : undefined,
                cloud_inst_id: lbDetail.value?.extension?.snat_pro ? undefined : cloud_id,
                port,
                weight,
              };
            }),
        },
      ],
    });
    // 处理参数 - 批量修改端口/权重
    const resolveFormDataForBatchUpdate = (type: 'port' | 'weight') => ({
      target_ids: formData.rs_list.map(({ id }) => id),
      [`new_${type}`]: formData.rs_list[0][type],
    });
    // 处理参数 - 批量移除rs
    const resolveFormDataForBatchDeleteRs = () => ({
      account_id: formData.account_id,
      target_groups: [{ target_group_id: formData.id, target_ids: deletedRsList.value.map((item) => item.id) }],
    });

    // check-status - 查询异步任务执行状态
    const reqAsyncTaskStatus = (tgId: string, flowId: string) => {
      businessStore.getAsyncTaskDetail(flowId).then(({ data: { state } }) => {
        if (state === 'success') {
          // 如果异步任务状态为 success, 则重新拉取 detail 详情
          Object.assign(lastAsyncTaskInfo, { tgId: '', flowId: '', state });
          businessStore.getTargetGroupDetail(tgId).then((tgDetailRes: any) => {
            handleEditTargetGroup({ ...tgDetailRes.data, rs_list: tgDetailRes.data.target_list });
          });
        } else if (['canceled', 'failed'].includes(state)) {
          // 如果异步任务为非 success 的结束状态, 停止轮询, 并给用户错误提示
          clearInterval(timer);
          Object.assign(lastAsyncTaskInfo, { tgId, flowId, state });
        } else {
          // 如果异步任务状态为非结束状态, 则记录异步任务id, 当用户下一次点击该目标组详情时, 再查询一次异步任务状态
          Object.assign(lastAsyncTaskInfo, { tgId, flowId, state });
        }
      });
    };

    // submit - [新增/编辑目标组] 或 [批量添加rs] 或 [批量修改端口] 或 [批量修改权重]
    const handleAddOrUpdateTargetGroupSubmit = async () => {
      await formRef.value.validate();
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
          message = 'RS添加异步任务已提交';
          break;
        case 'port':
          promise = businessStore.batchUpdateRsPort(formData.id, resolveFormDataForBatchUpdate('port'));
          message = '批量修改端口异步任务已提交';
          break;
        case 'weight':
          promise = businessStore.batchUpdateRsWeight(formData.id, resolveFormDataForBatchUpdate('weight'));
          message = '批量修改权重异步任务已提交';
          break;
        case 'BatchDeleteRs':
          promise = businessStore.batchDeleteTargets(resolveFormDataForBatchDeleteRs());
          message = '批量移除RS异步任务已提交';
          break;
      }
      try {
        isSubmitLoading.value = true;
        const { data } = await promise;
        Message({ message, theme: 'success' });
        // 异步任务非结束状态, 记录异步任务flow_id以及当前操作目标组id
        if (data?.flow_id) {
          Object.assign(lastAsyncTaskInfo, { tgId: formData.id, flowId: data.flow_id, state: 'pending' });
          // 重置状态
          handleEditTargetGroup({ ...formData });
        }
        // 初始化场景值
        loadBalancerStore.setUpdateCount(0);
        // 关闭侧栏
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
        isSubmitLoading.value = false;
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
      // 清除定时器
      clearInterval(timer);
    });

    return () => (
      <CommonSideslider
        title={isEdit.value ? '编辑目标组' : '新建目标组'}
        width={960}
        v-model:isShow={isShow.value}
        isSubmitLoading={isSubmitLoading.value}
        isSubmitDisabled={isSubmitDisabled.value}
        onHandleSubmit={handleAddOrUpdateTargetGroupSubmit}
        handleClose={() => {
          if (['canceled', 'failed'].includes(lastAsyncTaskInfo.state)) {
            Object.assign(lastAsyncTaskInfo, { tgId: '', flowId: '', state: '' });
          }
        }}>
        <bk-container margin={0}>
          <Form formType='vertical' model={formData} ref={formRef} rules={rules}>
            {/* 异步任务提示 */}
            {(function () {
              const { state } = lastAsyncTaskInfo;
              if (state === 'success' || !state) return;
              if (['canceled', 'failed'].includes(state)) {
                return (
                  <Alert theme='danger' class='mb24'>
                    当前目标组有异步任务存在异常，
                    <Button text theme='primary' onClick={() => goAsyncTaskDetail(lastAsyncTaskInfo.flowId)}>
                      查看任务
                    </Button>
                    。
                  </Alert>
                );
              }
              return (
                <Alert theme='info' class='mb24'>
                  当前目标组有异步任务正在进行中，
                  <Button text theme='primary' onClick={() => goAsyncTaskDetail(lastAsyncTaskInfo.flowId)}>
                    查看任务
                  </Button>
                  。
                </Alert>
              );
            })()}
            {/* 操作类型提示 */}
            {loadBalancerStore.updateCount === 2 && loadBalancerStore.currentScene && (
              <Alert theme='info' class='mb24'>
                当前操作为；{TG_OPERATION_SCENE_MAP[loadBalancerStore.currentScene]}
              </Alert>
            )}
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
