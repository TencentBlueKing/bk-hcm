import { reactive, ref, nextTick, computed } from 'vue';
// import components
import { Message } from 'bkui-vue';
// import stores
import { useBusinessStore } from '@/store';
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import useSelectOptionListWithScroll from '@/hooks/useSelectOptionListWithScroll';
import useResolveListenerFormData from './useResolveListenerFormData';
// import utils
import bus from '@/common/bus';
import { cloneDeep } from 'lodash';
// import types
import { IOriginPage, QueryRuleOPEnum } from '@/typings';

export default (getListData: (...args: any) => any, originPage: IOriginPage) => {
  // use stores
  const businessStore = useBusinessStore();
  const loadBalancerStore = useLoadBalancerStore();

  const isSliderShow = ref(false);
  const isAddOrUpdateListenerSubmit = ref(false);
  const isEdit = ref(false); // 标识当前是否为「编辑」操作
  const isSniOpen = ref(false); // 用于「编辑」操作. 记录SNI是否开启, 如果开启, 编辑的时候不可关闭
  const isLbLocked = ref(false); // 当前负载均衡是否被锁定
  const lockedLbInfo = ref(null);
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
    port: [
      {
        validator: (value: number) => value >= 1 && value <= 65535,
        message: '端口号不符合规范',
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
    url: [
      {
        validator: (value: string) => /^\/[\w\-/]*$/.test(value),
        message: 'URL路径不符合规范',
        trigger: 'change',
      },
    ],
    'certificate.cert_cloud_ids': [
      {
        validator: (value: string[]) => value.length <= 2,
        message: '最多选择 2 个证书',
        trigger: 'change',
      },
      {
        validator: (value: string[]) => {
          // 判断证书类型是否重复
          const [cert1, cert2] = SVRCertList.value.filter((cert) => value.includes(cert.cloud_id));
          return cert1?.encrypt_algorithm !== cert2?.encrypt_algorithm;
        },
        message: '不能选择加密算法相同的证书',
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
    target_group_name: '',
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
  const initOptionState = () => {
    // init state
    initTargetGroupOptionState();
    initSVRCertOptionState();
    initCACertOptionState();
  };

  // 新增监听器
  const handleAddListener = () => {
    // 初始化
    initOptionState();
    getSVRCertList();
    getCACertList();
    getTargetGroupList();
    isEdit.value = false;
    isSniOpen.value = false;
    isSliderShow.value = true;
    clearFormData();
    nextTick(() => {
      formRef.value.clearValidate();
    });
  };

  // 编辑监听器
  const handleEditListener = async (id: string) => {
    // 初始化
    isEdit.value = true;
    clearFormData();
    initOptionState();
    // 获取监听器详情, 回填
    const { data } = await businessStore.detail('listeners', id);
    const certificate = cloneDeep(data.certificate);
    Object.assign(listenerFormData, { ...data, domain: data.default_domain, session_open: data.session_expire !== 0 });

    // get list
    if (!isEdit.value || data.protocol === 'HTTPS') {
      await getSVRCertList();
      await getCACertList();
      Object.assign(listenerFormData.certificate, certificate);
    }

    isSniOpen.value = !!data.sni_switch;
    isSliderShow.value = true;
  };

  const computedProtocol = computed(() => listenerFormData.protocol);

  // 查询负载均衡是否处于锁定状态
  const checkLbIsLocked = async () => {
    const lbId = loadBalancerStore.currentSelectedTreeNode.lb?.id || loadBalancerStore.currentSelectedTreeNode.id;
    const res = await businessStore.getLBLockStatus(lbId);
    const status = res?.data?.status;
    if (status !== 'success') {
      isLbLocked.value = true;
      lockedLbInfo.value = res.data;
      return Promise.reject();
    }
    isLbLocked.value = false;
    lockedLbInfo.value = null;
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
    lockedLbInfo,
  };
};
