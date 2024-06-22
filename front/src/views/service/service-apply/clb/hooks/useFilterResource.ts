import { computed, reactive, ref, shallowRef, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useBusinessStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import bus from '@/common/bus';
// import types
import { ApplyClbModel, SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { reqResourceListOfCurrentRegion } from '@/api/load_balancers/apply-clb';
import { ClbQuota, LbPrice } from '@/typings';
import { debounce } from 'lodash';

// 当云地域变更时, 获取用户在当前地域支持可用区列表和资源列表
export default (formModel: ApplyClbModel) => {
  const { isBusinessPage } = useWhereAmI();
  const businessStore = useBusinessStore();
  // define data
  const isResourceListLoading = ref(false); // 是否正在获取资源列表
  const currentResourceListMap = reactive({}); // 资源映射
  const ispList = ref([]); // 运营商类型
  const specAvailabilitySet = ref<Array<SpecAvailability>>([]); // 负载均衡规格类型
  const quotas = ref<ClbQuota[]>([]); // 配额
  const isInquiryPricesLoading = ref(false); // 是否正在询价
  const isInquiryPrices = computed(() => {
    // 内网下, account_id, region, zones, cloud_vpc_id, cloud_subnet_id, require_count, name 不为空时才询价一次
    if (formModel.load_balancer_type === 'INTERNAL') {
      return Boolean(
        formModel.account_id &&
          formModel.region &&
          formModel.zones &&
          formModel.cloud_vpc_id &&
          formModel.cloud_subnet_id &&
          formModel.require_count !== 0 &&
          formModel.name &&
          /^[a-zA-Z0-9]([-a-zA-Z0-9]{0,58})[a-zA-Z0-9]$/.test(formModel.name),
      );
    }
    // 公网下, 如果账号类型为传统类型, 则 account_id, region, address_ip_version, zones, cloud_vpc_id, vip_isp, sla_type, require_count, name 不为空时才询价一次
    if (formModel.account_type === 'LEGACY') {
      return Boolean(
        formModel.account_id &&
          formModel.region &&
          formModel.address_ip_version &&
          formModel.zones &&
          formModel.cloud_vpc_id &&
          formModel.vip_isp &&
          formModel.sla_type &&
          formModel.require_count !== 0 &&
          formModel.name &&
          /^[a-zA-Z0-9]([-a-zA-Z0-9]{0,58})[a-zA-Z0-9]$/.test(formModel.name),
      );
    }
    // 公网下, 如果账号类型为标准类型, 则 account_id, region, address_ip_version, zones, cloud_vpc_id, vip_isp, sla_type, internet_charge_type, internet_max_bandwidth_out, require_count, name 不为空时才询价一次
    return Boolean(
      formModel.account_id &&
        formModel.region &&
        formModel.address_ip_version &&
        formModel.zones &&
        formModel.cloud_vpc_id &&
        formModel.vip_isp &&
        formModel.sla_type &&
        formModel.internet_charge_type &&
        formModel.internet_max_bandwidth_out &&
        formModel.require_count !== 0 &&
        formModel.name &&
        /^[a-zA-Z0-9]([-a-zA-Z0-9]{0,58})[a-zA-Z0-9]$/.test(formModel.name),
    );
  }); // 是否询价
  const prices = shallowRef<LbPrice>(); // 价格信息

  // 前端改映射
  const ipVersionMap = {
    IPV4: 'ipv4',
    IPv6FullChain: 'ipv6',
    IPV6: 'ipv6_nat',
  };

  /**
   * 获取当前地域「可用区列表和资源列表的映射关系」
   * @param region 地域
   */
  const getResourceListOfCurrentRegion = async (region: string) => {
    isResourceListLoading.value = true;
    const { data } = await reqResourceListOfCurrentRegion({ account_id: formModel.account_id, region });
    const { ZoneResourceSet } = data;
    ZoneResourceSet.forEach(({ MasterZone, IPVersion, ResourceSet }) => {
      // '主可用区|IP版本' 为对象的 key, [{ Isp, TypeSet }, { Isp, TypeSet }...] 为对象的 value
      const key = `${MasterZone}|${IPVersion}`.toLowerCase();

      ResourceSet?.forEach(({ Isp, TypeSet }) => {
        currentResourceListMap[key] = currentResourceListMap[key] || {};
        currentResourceListMap[key][Isp] = currentResourceListMap[key][Isp] || [];

        TypeSet?.forEach(({ SpecAvailabilitySet }) => {
          currentResourceListMap[key][Isp].push(...SpecAvailabilitySet);
        });
      });
    });
    isResourceListLoading.value = false;
  };

  /**
   * 获取当前地域「腾讯云账号负载均衡的配额」
   * @param region 地域
   */
  const getLbQuotas = async (region: string) => {
    const { data } = await businessStore.getClbQuotas({ account_id: formModel.account_id, region });
    // 根据 load_balancer_type 获取对应的配额名称
    // 设置当前地域 CLB 的配额信息
    quotas.value = data;
  };

  /**
   * 询价
   */
  const inquiryPrices = async () => {
    isInquiryPricesLoading.value = true;
    try {
      const { data } = await businessStore.lbPricesInquiry({
        ...formModel,
        bk_biz_id: isBusinessPage ? formModel.bk_biz_id : undefined,
        zones: [formModel.zones],
        backup_zones: formModel.backup_zones ? [formModel.backup_zones] : undefined,
        bandwidthpkg_sub_type: formModel.vip_isp === 'BGP' ? 'BGP' : 'SINGLE_ISP',
        bandwidth_package_id: undefined,
      });
      prices.value = data;
    } finally {
      isInquiryPricesLoading.value = false;
    }
  };

  watch(
    () => formModel.region,
    (val) => {
      if (!val) return;
      // 当云地域变更时, 获取当前地域「可用区列表和资源列表的映射关系」
      getResourceListOfCurrentRegion(val);
      // 当云地域变更时, 获取当前地域「腾讯云账号负载均衡的配额」
      getLbQuotas(val);
    },
  );

  watch(
    [() => formModel.zones, () => formModel.address_ip_version],
    () => {
      const { zones, address_ip_version, region } = formModel;
      // 拼接key, 用于定位对应的 isp 列表
      const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
      const key = `${zonesRule || null}|${ipVersionMap[address_ip_version]}`.toLowerCase();
      ispList.value = Object.keys(currentResourceListMap[key] || {}).filter(
        (isp) =>
          // 内网下的 isp 选项不显示
          isp !== 'INTERNAL' &&
          // 如果机型可用性全部为Unavailable, 则该 isp 选项不显示
          currentResourceListMap[key][isp].some(({ Availability }: any) => Availability !== 'Unavailable'),
      );
    },
    {
      deep: true,
    },
  );

  watch(
    ispList,
    () => {
      formModel.vip_isp = ispList.value.length ? 'BGP' : '';
      formModel.slaType = '0';
      formModel.sla_type = 'shared';
    },
    { deep: true },
  );

  watch(
    () => formModel.vip_isp,
    () => {
      const { zones, address_ip_version, vip_isp, region } = formModel;
      specAvailabilitySet.value = [];
      if (!vip_isp) {
        formModel.sla_type = 'shared';
        return;
      }
      const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
      const key = `${zonesRule || null}|${ipVersionMap[address_ip_version]}`.toLowerCase();
      // 公有云TypeSet数组暂时取第一个元素, 忽略 Type 的作用, 直接取 SpecAvailabilitySet 作为性能容量型的机型选择
      specAvailabilitySet.value = currentResourceListMap[key][vip_isp]
        // 机型可用性全部为Unavailable时, 才能在性能容量型选项中选择该机型
        .filter(({ Availability }: any) => Availability !== 'Unavailable');
    },
  );

  watch(
    specAvailabilitySet,
    (val) => {
      // 重置 sla_type
      formModel.slaType = '0';
      formModel.sla_type = 'shared';
      if (val) {
        bus.$emit(
          'updateSpecAvailabilitySet',
          val.filter(({ SpecType }) => SpecType !== 'shared'),
        );
      } else {
        Message({ theme: 'warning', message: '当前地域下无可用规格, 请切换地域' });
      }
    },
    {
      deep: true,
    },
  );

  watch(
    formModel,
    debounce(() => {
      // 询价
      if (isInquiryPrices.value) inquiryPrices();
      else prices.value = { bandwidth_price: null, instance_price: null, lcu_price: null };
    }, 500),
    { deep: true },
  );

  watch(
    prices,
    (val) => {
      bus.$emit('changeLbPrice', val);
    },
    { deep: true },
  );

  return {
    ispList,
    isResourceListLoading,
    quotas,
    isInquiryPricesLoading,
  };
};
