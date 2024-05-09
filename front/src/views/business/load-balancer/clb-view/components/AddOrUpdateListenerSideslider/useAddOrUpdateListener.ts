import { reactive, ref, nextTick, computed } from 'vue';
// import components
import { Message } from 'bkui-vue';
// import stores
import { useBusinessStore, useResourceStore } from '@/store';
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import useSelectOptionListWithScroll from '@/hooks/useSelectOptionListWithScroll';
import useResolveListenerFormData from './useResolveListenerFormData';
// import utils
import bus from '@/common/bus';
// import types
import { IOriginPage, QueryRuleOPEnum } from '@/typings';

export default (getListData: (...args: any) => any, originPage: IOriginPage) => {
  // use stores
  const businessStore = useBusinessStore();
  const resourceStore = useResourceStore();
  const loadBalancerStore = useLoadBalancerStore();

  const isSliderShow = ref(false);
  const isAddOrUpdateListenerSubmit = ref(false);
  const isEdit = ref(false); // 标识当前是否为「编辑」操作
  const isSniOpen = ref(false); // 用于「编辑」操作. 记录SNI是否开启, 如果开启, 编辑的时候不可关闭
  const isLbLocked = ref(false); // 当前负载均衡是否被锁定
  // 表单相关
  const formRef = ref();
  const rules = {
    name: [
      {
        validator: (value: string) => /^[\u4e00-\u9fa5A-Za-z0-9\-._:]{1,60}$/.test(value),
        message: '不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号',
        trigger: 'change',
      },
    ],
    domain: [
      {
        validator: (value: string) => /^(?:(?:[a-zA-Z0-9]+-?)+(?:\.[a-zA-Z0-9-]+)+)$/.test(value),
        message: '域名不符合规范',
        trigger: 'change',
      },
    ],
  };

  const getDefaultFormData = () => ({
    account_id: loadBalancerStore.currentSelectedTreeNode.account_id,
    lb_id: loadBalancerStore.currentSelectedTreeNode.id,
    name: '',
    protocol: 'TCP',
    port: '',
    scheduler: '',
    session_open: false,
    session_type: 'NORMAL',
    session_expire: 0,
    target_group_id: '',
    domain: '',
    url: '/',
    sni_switch: 0,
    certificate: {
      ssl_mode: 'UNIDIRECTIONAL',
      ca_cloud_id: '',
      cert_cloud_ids: [] as any[],
    },
  });
  const listenerFormData = reactive(getDefaultFormData());

  // 清空表单数据
  const clearFormData = () => {
    Object.assign(listenerFormData, getDefaultFormData());
  };

  // 初始化select-option列表
  const initOptionState = (protocol?: string, isSniOpen?: boolean) => {
    // init state
    initTargetGroupOptionState();
    initSVRCertOptionState();
    initCACertOptionState();
    // get list
    if (!isEdit.value || (protocol === 'HTTPS' && !isSniOpen)) {
      getSVRCertList();
      getCACertList();
    }
    if (!isEdit.value) {
      getTargetGroupList();
    }
  };

  // 新增监听器
  const handleAddListener = () => {
    initOptionState();
    isEdit.value = false;
    isSliderShow.value = true;
    clearFormData();
    nextTick(() => {
      formRef.value.clearValidate();
    });
  };

  // 编辑监听器
  const handleEditListener = (id: string) => {
    clearFormData();
    // 获取监听器详情, 回填
    resourceStore.detail('listeners', id).then(({ data }: any) => {
      Object.assign(listenerFormData, data, {
        domain: data.default_domain,
        session_open: data.session_expire !== 0,
        certificate: data.extension.certificate || {
          ssl_mode: 'UNIDIRECTIONAL',
          ca_cloud_id: '',
          cert_cloud_ids: [],
        },
      });
      isSniOpen.value = !!data.sni_switch;
      isEdit.value = true;
      initOptionState(data.protocol, isSniOpen.value);
      isSliderShow.value = true;
    });
  };

  const computedProtocol = computed(() => listenerFormData.protocol);

  // 查询负载均衡是否处于锁定状态
  const checkLbIsLocked = async () => {
    const res = await businessStore.getLBLockStatus(loadBalancerStore.currentSelectedTreeNode.id);
    const status = res?.data?.status;
    if (status !== 'success') {
      isLbLocked.value = true;
      return Promise.reject();
    }
    isLbLocked.value = false;
    return res;
  };

  // submit handler
  const handleAddOrUpdateListener = async () => {
    try {
      await formRef.value.validate();
      isAddOrUpdateListenerSubmit.value = true;
      await checkLbIsLocked();
      if (isEdit.value) {
        // 编辑监听器
        await businessStore.updateListener({
          ...listenerFormData,
          extension: { certificate: listenerFormData.certificate },
        });
      } else {
        // 新增监听器
        await businessStore.createListener(listenerFormData);
      }
      Message({ theme: 'success', message: isEdit.value ? '更新成功' : '新增成功' });
      isSliderShow.value = false;
      typeof getListData === 'function' && getListData();
      // 如果是在监听器页面更新监听器详情, 则更新成功后刷新监听器详情
      originPage === 'listener' && bus.$emit('refreshListenerDetail');
    } finally {
      isAddOrUpdateListenerSubmit.value = false;
    }
  };

  // 目标组 options
  const {
    isScrollLoading: isTargetGroupListLoading,
    optionList: targetGroupList,
    initState: initTargetGroupOptionState,
    getOptionList: getTargetGroupList,
    handleOptionListScrollEnd: handleTargetGroupListScrollEnd,
    isFlashLoading: isTargetGroupListFlashLoading,
    handleRefreshOptionList: handleTargetGroupListRefreshOptionList,
  } = useSelectOptionListWithScroll(
    'target_groups',
    [
      {
        field: 'account_id',
        op: QueryRuleOPEnum.EQ,
        value: loadBalancerStore.currentSelectedTreeNode.account_id,
      },
      {
        field: 'cloud_vpc_id',
        op: QueryRuleOPEnum.EQ,
        value:
          loadBalancerStore.currentSelectedTreeNode.cloud_vpc_id ||
          loadBalancerStore.currentSelectedTreeNode.lb.cloud_vpc_id,
      },
      {
        field: 'region',
        op: QueryRuleOPEnum.EQ,
        value: loadBalancerStore.currentSelectedTreeNode.region || loadBalancerStore.currentSelectedTreeNode.lb.region,
      },
    ],
    false,
    computedProtocol,
  );

  // 服务器证书 options
  const {
    isScrollLoading: isSVRCertListLoading,
    optionList: SVRCertList,
    initState: initSVRCertOptionState,
    getOptionList: getSVRCertList,
    handleOptionListScrollEnd: handleSVRCertListScrollEnd,
  } = useSelectOptionListWithScroll(
    'certs',
    [
      { field: 'cert_type', op: QueryRuleOPEnum.EQ, value: 'SVR' },
      {
        field: 'account_id',
        op: QueryRuleOPEnum.EQ,
        value: loadBalancerStore.currentSelectedTreeNode.account_id,
      },
    ],
    false,
  );

  // 客户端证书 options
  const {
    isScrollLoading: isCACertListLoading,
    optionList: CACertList,
    initState: initCACertOptionState,
    getOptionList: getCACertList,
    handleOptionListScrollEnd: handleCACertListScrollEnd,
  } = useSelectOptionListWithScroll(
    'certs',
    [
      { field: 'cert_type', op: QueryRuleOPEnum.EQ, value: 'CA' },
      {
        field: 'account_id',
        op: QueryRuleOPEnum.EQ,
        value: loadBalancerStore.currentSelectedTreeNode.account_id,
      },
    ],
    false,
  );

  // 参数处理
  useResolveListenerFormData(listenerFormData);

  return {
    isSliderShow,
    isEdit,
    isAddOrUpdateListenerSubmit,
    isSniOpen,
    formRef,
    rules,
    listenerFormData,
    handleAddListener,
    handleEditListener,
    handleAddOrUpdateListener,
    isTargetGroupListLoading,
    targetGroupList,
    handleTargetGroupListScrollEnd,
    isTargetGroupListFlashLoading,
    handleTargetGroupListRefreshOptionList,
    isSVRCertListLoading,
    SVRCertList,
    handleSVRCertListScrollEnd,
    isCACertListLoading,
    CACertList,
    handleCACertListScrollEnd,
    isLbLocked,
  };
};
