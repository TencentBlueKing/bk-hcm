import { reactive, ref, nextTick } from 'vue';
// import components
import { Message } from 'bkui-vue';
// import stores
import { useBusinessStore } from '@/store';
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import useResolveListenerFormData from './useResolveListenerFormData';
// import utils
import bus from '@/common/bus';
// import types
import { IOriginPage, Protocol } from '@/typings';

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

  const getDefaultCertificate = () => ({
    ssl_mode: 'UNIDIRECTIONAL',
    ca_cloud_id: '',
    cert_cloud_ids: [] as any[],
  });
  const getDefaultFormData = () => ({
    id: '',
    account_id: loadBalancerStore.currentSelectedTreeNode.account_id,
    lb_id: loadBalancerStore.currentSelectedTreeNode.id,
    name: '',
    protocol: 'TCP' as Protocol,
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
    certificate: getDefaultCertificate(),
  });
  const listenerFormData = reactive(getDefaultFormData());

  // 清空表单数据
  const clearFormData = () => {
    Object.assign(listenerFormData, getDefaultFormData());
  };

  // 新增监听器
  const handleAddListener = () => {
    // 初始化
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
    // 获取监听器详情, 回填
    const { data } = await businessStore.detail('listeners', id);
    Object.assign(listenerFormData, {
      ...data,
      domain: data.default_domain,
      session_open: data.session_expire !== 0,
      // SNI开启时，证书在域名上；SNI关闭时，域名在监听器上
      certificate: (data.sni_switch ? data.certificate : data.extension.certificate) || getDefaultCertificate(),
    });

    isSniOpen.value = !!data.sni_switch;
    isSliderShow.value = true;
  };

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
          id: listenerFormData.id,
          account_id: listenerFormData.account_id,
          name: listenerFormData.name,
          sni_switch: listenerFormData.sni_switch,
          extension: { certificate: listenerFormData.protocol === 'HTTPS' ? listenerFormData.certificate : undefined },
        });
        // 如果启用了SNI, 需要调用规则更新接口来更新证书信息
        if (listenerFormData.sni_switch) {
          await businessStore.updateDomains(listenerFormData.id, {
            lbl_id: listenerFormData.id,
            domain: (listenerFormData as any).default_domain,
            certificate: listenerFormData.protocol === 'HTTPS' ? listenerFormData.certificate : undefined,
          });
        }
      } else {
        // 新增监听器
        const params = {
          ...listenerFormData,
          // 只有https协议才需要传证书
          certificate: listenerFormData.protocol === 'HTTPS' ? listenerFormData.certificate : undefined,
        };
        await businessStore.createListener(params);
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

  // 参数处理
  useResolveListenerFormData(listenerFormData);

  return {
    isSliderShow,
    isEdit,
    isAddOrUpdateListenerSubmit,
    isSniOpen,
    formRef,
    listenerFormData,
    handleAddListener,
    handleEditListener,
    handleAddOrUpdateListener,
    isLbLocked,
    lockedLbInfo,
  };
};
