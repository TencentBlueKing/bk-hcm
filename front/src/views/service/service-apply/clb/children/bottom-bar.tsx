import { computed, defineComponent, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { AUTH_BIZ_CREATE_CLB, AUTH_CREATE_CLB } from '@/constants/auth-symbols';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { VendorEnum } from '@/common/constant';
import { BGP_VIP_ISP_TYPES } from '@/constants';

export default defineComponent({
  props: {
    list: Array as PropType<ApplyClbModel[]>,
    loading: Boolean as PropType<boolean>,
    onConfirm: Function as PropType<(params: ApplyClbModel[]) => void>,
    onCancel: Function as PropType<() => void>,
  },
  setup(props, { emit }) {
    const { t } = useI18n();
    const { whereAmI, isBusinessPage, getBizsId } = useWhereAmI();

    // 权限校验
    const computedBizId = computed(() => (whereAmI.value === Senarios.business ? getBizsId() : undefined));
    const createClbAuthSymbol = computed(() => {
      return whereAmI.value === Senarios.business ? AUTH_BIZ_CREATE_CLB : AUTH_CREATE_CLB;
    });

    const isOpen = (loadbalancerType: 'OPEN' | 'INTERNAL') => loadbalancerType === 'OPEN';
    const isIpv4 = (addressIP: 'IPV4' | 'IPv6FullChain' | 'IPV6') => addressIP === 'IPV4';
    const hasZonesConfig = (data: ApplyClbModel) => {
      const { load_balancer_type, address_ip_version } = data;
      const openVal = isOpen(load_balancer_type);
      const IpV4Val = isIpv4(address_ip_version);
      return (openVal && IpV4Val) || !openVal;
    };
    const hasBackupZonesConfig = (data: ApplyClbModel) => {
      const { load_balancer_type, address_ip_version } = data;
      const openVal = isOpen(load_balancer_type);
      const IpV4Val = isIpv4(address_ip_version);
      return openVal && IpV4Val;
    };
    const hasInternetChargeTypeConfig = (data: ApplyClbModel) => {
      const { load_balancer_type, account_type } = data;
      const openVal = isOpen(load_balancer_type);
      return openVal && account_type !== 'LEGACY';
    };
    // 提交
    const handleParams = (formModel: ApplyClbModel) => {
      const { load_balancer_type } = formModel;
      const isOpenVal = isOpen(load_balancer_type);
      // eslint-disable-next-line
      let zones = hasZonesConfig(formModel) ? (formModel.zones ? [formModel.zones] : []) : undefined;
      let vipIsp = isOpenVal ? formModel.vip_isp : undefined;
      let tgwGroupName = formModel.tgw_group_name;

      if (formModel.vendor === VendorEnum.ZIYAN) {
        if (!isOpenVal) {
          // 自研云内网下支持多可用区
          zones = formModel.zones as string[];
          tgwGroupName = undefined;
          vipIsp = undefined;
        } else {
          // VipIsp不传就是BGP/限速CAP
          if (BGP_VIP_ISP_TYPES.includes(vipIsp)) {
            tgwGroupName = vipIsp;
            vipIsp = undefined;
          }
        }
      } else {
        tgwGroupName = undefined;
      }

      return {
        ...formModel,
        bk_biz_id: isBusinessPage ? formModel.bk_biz_id : undefined,
        sla_type: formModel.sla_type === 'shared' ? '' : formModel.sla_type,
        // 只有公网下可以配置
        address_ip_version: isOpenVal ? formModel.address_ip_version : undefined,
        vip_isp: vipIsp,
        // 只有公网下的标准账号可以配置（内网支持配置带宽上限）
        internet_charge_type: hasInternetChargeTypeConfig(formModel) ? formModel.internet_charge_type : undefined,
        internet_max_bandwidth_out:
          hasInternetChargeTypeConfig(formModel) || !isOpenVal ? formModel.internet_max_bandwidth_out : undefined,
        // 只有公网下ipv4以及内网下可以配置
        zones,
        // 只有公网下ipv4可以配置
        // eslint-disable-next-line no-nested-ternary
        backup_zones: hasBackupZonesConfig(formModel)
          ? formModel.backup_zones
            ? [formModel.backup_zones]
            : []
          : undefined,
        // 内网/公网IPv6需要选择子网
        cloud_subnet_id:
          !isOpenVal || (isOpenVal && formModel.address_ip_version === 'IPv6FullChain')
            ? formModel.cloud_subnet_id
            : undefined,
        // 内网下支持EIP
        cloud_eip_id: !isOpenVal ? formModel.cloud_eip_id ?? undefined : undefined,
        // 自研云字段
        tgw_group_name: tgwGroupName,
        zhi_tong: formModel.vendor === VendorEnum.ZIYAN && isOpenVal ? formModel.zhi_tong : undefined,
        // 带宽包ID只在公网且计费模式为带宽包时才传递
        bandwidth_package_id:
          isOpenVal && formModel.internet_charge_type === 'BANDWIDTH_PACKAGE'
            ? formModel.bandwidth_package_id
            : undefined,
        // 后端无用字段
        account_type: undefined as undefined,
        zoneType: undefined as undefined,
        slaType: undefined as undefined,
      };
    };
    const handleConfirm = async () => {
      const { list } = props;
      if (!list.length) return;
      const params = list.map((item: ApplyClbModel) => handleParams(item));
      emit('confirm', params);
    };
    const handleCancel = () => {
      emit('cancel');
    };

    return () => (
      <div class={['save-footer', 'apply-clb-bottom-bar', { 'business-apply-clb-bottom-bar': !isBusinessPage }]}>
        <hcm-auth sign={{ type: createClbAuthSymbol.value, relation: [computedBizId.value] }} class='mr10'>
          {{
            default: ({ noPerm }: { noPerm: boolean }) => (
              <bk-button
                theme='primary'
                onClick={handleConfirm}
                loading={props.loading}
                disabled={!props.list.length || props.list.length > 5 || noPerm}>
                {t('提交')}
              </bk-button>
            ),
          }}
        </hcm-auth>
        <bk-button onClick={handleCancel}>{t('取消')}</bk-button>
      </div>
    );
  },
});
