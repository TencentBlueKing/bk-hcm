import { Ref, computed, reactive, ref } from 'vue';
// import components
import { Input, Message, Select } from 'bkui-vue';
// import hooks
import { useI18n } from 'vue-i18n';
import { useBusinessStore } from '@/store';
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useRoute } from 'vue-router';
// import types
import { IOriginPage } from '@/typings';

const { Option } = Select;

export enum OpAction {
  ADD,
  UPDATE,
}

export const RuleModeList = [
  {
    id: 'WRR',
    name: '按权重轮询',
  },
  {
    id: 'LEAST_CONN',
    name: '最小连接数',
  },
  {
    id: 'IP_HASH',
    name: 'IP Hash',
  },
];

export default (getListData: () => void, originPage: IOriginPage) => {
  // use hooks
  const { t } = useI18n();
  const businessStore = useBusinessStore();
  const loadbalancer = useLoadBalancerStore();
  const route = useRoute();
  const oldDomain = ref('');

  const isShow = ref(false);
  const action = ref<number>(); // 0 - add, 1 - update
  const formData = reactive({
    domain: '',
    url: '',
    scheduler: '',
  });

  // 清空表单参数
  const clearParams = () => {
    Object.assign(formData, {
      domain: '',
      url: '',
      scheduler: '',
    });
  };

  /**
   * 显示 dialog
   * @param data 域名信息, 如果为 undefined, 表示新增
   */
  const handleShow = (data?: any) => {
    isShow.value = true;
    clearParams();
    if (data) {
      action.value = OpAction.UPDATE;
      Object.assign(formData, data);
      oldDomain.value = data.domain;
    } else {
      action.value = OpAction.ADD;
    }
  };

  const handleSubmit = async (formInstance: Ref<any>) => {
    await formInstance.value.validate();
    const lbl_id =
      originPage === 'listener'
        ? loadbalancer.currentSelectedTreeNode.id
        : loadbalancer.currentSelectedTreeNode.listener_id;
    const promise =
      action.value === OpAction.ADD
        ? businessStore.createRules({
            bk_biz_id: route.query.bizs,
            lbl_id,
            rules: [
              {
                url: formData.url,
                domains: [formData.domain],
                scheduler: formData.scheduler,
              },
            ],
          })
        : businessStore.updateDomains(lbl_id, {
            lbl_id,
            domain: oldDomain.value,
            new_domain: formData.domain,
          });
    await promise;
    isShow.value = false;
    Message({
      message: action.value === OpAction.ADD ? '新建成功' : '编辑成功',
      theme: 'success',
    });
    getListData();
  };

  const formItemOptions = computed(() => [
    {
      label: t('域名'),
      property: 'domains',
      required: true,
      content: () => <Input v-model={formData.domain} />,
    },
    {
      label: t('URL 路径'),
      property: 'url',
      required: true,
      hidden: action.value === OpAction.UPDATE,
      content: () => <Input v-model={formData.url} />,
    },
    {
      label: '均衡方式',
      property: 'scheduler',
      required: true,
      hidden: action.value === OpAction.UPDATE,
      content: () => (
        <Select v-model={formData.scheduler} placeholder={t('请选择模式')}>
          {RuleModeList.map(({ id, name }) => (
            <Option name={name} id={id} />
          ))}
        </Select>
      ),
    },
  ]);

  return {
    isShow,
    action,
    formItemOptions,
    handleShow,
    handleSubmit,
    formData,
  };
};
