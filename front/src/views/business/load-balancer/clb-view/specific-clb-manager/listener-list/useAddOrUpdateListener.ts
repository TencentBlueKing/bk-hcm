import { reactive, ref, watch, nextTick } from 'vue';
// import components
import { Message } from 'bkui-vue';
// import stores
import { useBusinessStore, useResourceStore } from '@/store';
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import useSelectOptionList from '@/hooks/useSelectOptionListWithScroll';
// import types
import { QueryRuleOPEnum } from '@/typings';

export default (getListData: any) => {
  // use stores
  const businessStore = useBusinessStore();
  const resourceStore = useResourceStore();
  const loadBalancerStore = useLoadBalancerStore();

  const isSliderShow = ref(false);
  const isAddOrUpdateListenerSubmit = ref(false);
  const isEdit = ref(false); // 标识当前是否为「编辑」操作
  const isSniOpen = ref(false); // 用于「编辑」操作. 记录SNI是否开启, 如果开启, 编辑的时候不可关闭
  // 表单相关
  const formRef = ref();
  const rules = {
    name: [
      {
        validator: (value: string) => /^[\u4e00-\u9fa5A-Za-z0-9\-._:]{1,60}$/.test(value),
        message: '姓名格式不符合要求, 请重新输入!',
        trigger: 'change',
      },
    ],
  };
  const listenerFormData = reactive({
    account_id: loadBalancerStore.currentSelectedTreeNode.account_id,
    lb_id: loadBalancerStore.currentSelectedTreeNode.id,
    name: '',
    protocol: 'TCP',
    port: undefined,
    scheduler: '',
    session_open: true,
    session_type: 'NORMAL',
    session_expire: 30,
    target_group_id: '',
    domain: '',
    url: '/',
    sni_switch: 0,
    certificate: {
      ssl_mode: 'UNIDIRECTIONAL',
      ca_cloud_id: '',
      cert_cloud_ids: [],
    },
  });

  // 清空表单数据
  const clearFormData = () => {
    Object.assign(listenerFormData, {
      account_id: loadBalancerStore.currentSelectedTreeNode.account_id,
      lb_id: loadBalancerStore.currentSelectedTreeNode.id,
      name: '',
      protocol: 'TCP',
      port: undefined,
      scheduler: '',
      session_open: true,
      session_type: 'NORMAL',
      session_expire: 30,
      target_group_id: '',
      domain: '',
      url: '/',
      sni_switch: 0,
      certificate: {
        ssl_mode: 'UNIDIRECTIONAL',
        ca_cloud_id: '',
        cert_cloud_ids: [],
      },
    });
  };

  // 获取select-option列表
  const getOptionList = () => {
    getTargetGroupList();
    getSVRCertList();
    getCACertList();
  };

  // 新增监听器
  const handleAddListener = () => {
    getOptionList();
    isEdit.value = false;
    isSliderShow.value = true;
    clearFormData();
    nextTick(() => {
      formRef.value.clearValidate();
    });
  };

  // 编辑监听器
  const handleEditListener = (id: string) => {
    getOptionList();
    clearFormData();
    // 获取监听器详情, 回填
    resourceStore.detail('listeners', id).then(({ data }: any) => {
      Object.assign(listenerFormData, data, {
        domain: data.default_domain,
        session_open: data.session_expire !== 0,
        certificate: data.certificate || {
          ssl_mode: 'UNIDIRECTIONAL',
          ca_cloud_id: '',
          cert_cloud_ids: [],
        },
      });
      isSniOpen.value = !!data.sni_switch;
      isEdit.value = true;
      isSliderShow.value = true;
    });
  };

  // submit handler
  const handleAddOrUpdateListener = async () => {
    try {
      await formRef.value.validate();
      isAddOrUpdateListenerSubmit.value = true;
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
      getListData();
    } finally {
      isAddOrUpdateListenerSubmit.value = false;
    }
  };

  // 目标组 options
  const [isTargetGroupListLoading, targetGroupList, getTargetGroupList, handleTargetGroupListScrollEnd] =
    useSelectOptionList(
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
          value: loadBalancerStore.currentSelectedTreeNode.cloud_vpc_id,
        },
        {
          field: 'region',
          op: QueryRuleOPEnum.EQ,
          value: loadBalancerStore.currentSelectedTreeNode.region,
        },
      ],
      false,
    );

  // 服务器证书 options
  const [isSVRCertListLoading, SVRCertList, getSVRCertList, handleSVRCertListScrollEnd] = useSelectOptionList(
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
  const [isCACertListLoading, CACertList, getCACertList, handleCACertListScrollEnd] = useSelectOptionList(
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

  watch(
    () => listenerFormData.session_open,
    (val) => {
      // session_expire传0即为关闭会话保持
      val ? (listenerFormData.session_expire = 30) : (listenerFormData.session_expire = 0);
    },
  );

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
    isSVRCertListLoading,
    SVRCertList,
    handleSVRCertListScrollEnd,
    isCACertListLoading,
    CACertList,
    handleCACertListScrollEnd,
  };
};
