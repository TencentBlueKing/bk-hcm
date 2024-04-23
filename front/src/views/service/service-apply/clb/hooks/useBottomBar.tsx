import { computed, defineComponent, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { Button, Message, Popover, Table } from 'bkui-vue';
// import stores
import { useResourceStore } from '@/store';
// import types
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
// import utils
import { useI18n } from 'vue-i18n';

const { Column } = Table;

// apply-clb, 底栏
export default (formModel: ApplyClbModel, formRef: any) => {
  // use hooks
  const router = useRouter();
  const route = useRoute();
  const { whereAmI } = useWhereAmI();
  const { t } = useI18n();
  // use stores
  const resourceStore = useResourceStore();
  // define data
  const applyLoading = ref(false);
  const priceTableData = [
    {
      billingItem: '实例费用',
      billingMode: '包年包月',
      price: '114.00 元',
    },
    {
      billingItem: '网络费用',
      billingMode: '包月',
      price: '12.02 元',
    },
  ];

  const isOpen = computed(() => formModel.load_balancer_type === 'OPEN');
  const isIpv4 = computed(() => formModel.address_ip_version === 'IPV4');
  const hasZonesConfig = computed(() => (isOpen.value && isIpv4.value) || formModel.load_balancer_type === 'INTERNAL');
  const hasBackupZonesConfig = computed(() => isOpen.value && isIpv4.value);
  const hasInternetChargeTypeConfig = computed(() => isOpen.value && formModel.account_type !== 'LEGACY');

  const handleParams = () => {
    return {
      ...formModel,
      // 只有公网下可以配置
      address_ip_version: isOpen.value ? formModel.address_ip_version : undefined,
      vip_isp: isOpen.value ? formModel.vip_isp : undefined,
      // eslint-disable-next-line no-nested-ternary
      sla_type: isOpen.value ? (formModel.sla_type === 'shared' ? '' : formModel.sla_type) : undefined,
      // 只有公网下的标准账号可以配置
      internet_charge_type: hasInternetChargeTypeConfig.value ? formModel.internet_charge_type : undefined,
      internet_max_bandwidth_out: hasInternetChargeTypeConfig.value ? formModel.internet_charge_type : undefined,
      // 只有公网下ipv4以及内网下可以配置
      zones: hasZonesConfig.value ? [formModel.zones] : undefined,
      // 只有公网下ipv4可以配置
      // eslint-disable-next-line no-nested-ternary
      backup_zones: hasBackupZonesConfig.value ? (formModel.backup_zones ? [formModel.backup_zones] : []) : undefined,
      // 只有内网下可以配置
      cloud_subnet_id: !isOpen.value ? formModel.cloud_subnet_id : undefined,
      cloud_eip_id: !isOpen.value ? formModel.cloud_eip_id : undefined,
      // 后端无用字段
      account_type: undefined,
      zoneType: undefined,
    };
  };

  // define handler function
  const handleApplyClb = async () => {
    try {
      await formRef.value.validate();
      // 整理参数
      applyLoading.value = true;
      await resourceStore.create('load_balancers', handleParams());
      Message({ theme: 'success', message: '购买成功' });
      goBack();
    } finally {
      applyLoading.value = false;
    }
  };

  // 返回上一级, 并且不保留历史记录
  const goBack = () => {
    router.replace({
      path:
        whereAmI.value === Senarios.business
          ? `/business/loadbalancer/clb-view?bizs=${route.query.bizs}`
          : '/resource/resource?type=clb',
      query: { ...route.query },
    });
  };

  // define component
  const ApplyClbBottomBar = defineComponent({
    setup() {
      return () => (
        <div class='apply-clb-bottom-bar'>
          <div class='info-wrap'>
            <span class='label'>{t('IP资源费用')}</span>:
            <span class='value'>
              <span class='number'>0.01</span>
              <span class='unit'>{t('元/小时')}</span>
            </span>
          </div>
          <div class='info-wrap'>
            <Popover theme='light' trigger='click' width={362} placement='top' offset={12}>
              {{
                default: () => <span class='label has-tips'>{t('配置费用')}</span>,
                content: () => (
                  <Table data={priceTableData}>
                    <Column field='billingItem' label={t('计费项')}></Column>
                    <Column field='billingMode' label={t('计费模式')}></Column>
                    <Column field='price' label={t('价格')} align='right'></Column>
                  </Table>
                ),
              }}
            </Popover>
            :
            <span class='value'>
              <span class='unit'>￥</span>
              <span class='number'>126.02</span>
            </span>
          </div>
          <div class='operation-btn-wrap'>
            <Button theme='primary' onClick={handleApplyClb} loading={applyLoading.value}>
              {t('立即购买')}
            </Button>
            <Button loading={applyLoading.value} onClick={goBack}>
              {t('取消')}
            </Button>
          </div>
        </div>
      );
    },
  });

  return {
    ApplyClbBottomBar,
  };
};
