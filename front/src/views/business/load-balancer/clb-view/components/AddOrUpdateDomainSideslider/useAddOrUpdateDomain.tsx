import { Ref, computed, reactive, ref } from 'vue';
// import components
import { Input, Message, Select, Tag } from 'bkui-vue';
// import hooks
import { useI18n } from 'vue-i18n';
import { useBusinessStore } from '@/store';
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useRoute } from 'vue-router';
// import types
import { IOriginPage, QueryRuleOPEnum } from '@/typings';
import useSelectOptionListWithScroll from '@/hooks/useSelectOptionListWithScroll';
import BkRadio, { BkRadioGroup } from 'bkui-vue/lib/radio';
import CertSelector from '@/components/cert-selector';

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

export default (getListData: () => void, originPage: IOriginPage, isHttpsAndSniOn: boolean) => {
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
    target_group_id: '',
    certificate: {
      cert_cloud_ids: [],
      ca_cloud_id: '',
      ssl_mode: 'UNIDIRECTIONAL',
    },
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

  const { isScrollLoading, optionList, handleOptionListScrollEnd } = useSelectOptionListWithScroll(
    'target_groups',
    [
      {
        field: 'account_id',
        op: QueryRuleOPEnum.EQ,
        value: loadbalancer.currentSelectedTreeNode.lb.account_id,
      },
      {
        field: 'cloud_vpc_id',
        op: QueryRuleOPEnum.EQ,
        value: loadbalancer.currentSelectedTreeNode.lb.cloud_vpc_id,
      },
      {
        field: 'region',
        op: QueryRuleOPEnum.EQ,
        value: loadbalancer.currentSelectedTreeNode.lb.region,
      },
    ],
    true,
  );

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
            target_group_id: formData.target_group_id,
            lbl_id,
            url: formData.url,
            domains: [formData.domain],
            scheduler: formData.scheduler,
            certificate: isHttpsAndSniOn ? formData.certificate : undefined,
          })
        : businessStore.updateDomains(lbl_id, {
            lbl_id,
            domain: oldDomain.value,
            new_domain: formData.domain,
            certificate: isHttpsAndSniOn ? formData.certificate : undefined,
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
    {
      label: '目标组',
      property: 'target_group_id',
      required: true,
      hidden: action.value === OpAction.UPDATE,
      content: () => (
        <Select
          v-model={formData.target_group_id}
          scrollLoading={isScrollLoading.value}
          onScroll-end={handleOptionListScrollEnd}>
          {optionList.value.map(({ id, name, listener_num }) => (
            <Option key={id} id={id} name={name} disabled={listener_num > 0} />
          ))}
        </Select>
      ),
    },
    {
      label: 'SSL解析方式',
      required: true,
      hidden: !isHttpsAndSniOn,
      content: () => (
        <BkRadioGroup v-model={formData.certificate.ssl_mode}>
          <BkRadio label='UNIDIRECTIONAL'>
            {t('单向认证')}
            <Tag theme='info' class='recommend-tag ml4'>
              {t('推荐')}
            </Tag>
          </BkRadio>
          <BkRadio label='MUTUAL' class='ml24 ml4'>
            {t('双向认证')}
          </BkRadio>
        </BkRadioGroup>
      ),
    },
    {
      label: '服务器证书',
      required: true,
      hidden: !isHttpsAndSniOn,
      content: () => (
        <CertSelector
          accountId={loadbalancer.currentSelectedTreeNode.account_id}
          type='SVR'
          v-model={formData.certificate.cert_cloud_ids}
        />
      ),
    },
    {
      label: 'CA证书',
      required: true,
      hidden: formData.certificate.ssl_mode === 'UNIDIRECTIONAL' && !isHttpsAndSniOn,
      content: () => (
        <CertSelector
          accountId={loadbalancer.currentSelectedTreeNode.account_id}
          type='CA'
          v-model={formData.certificate.ca_cloud_id}
        />
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
