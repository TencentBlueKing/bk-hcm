import { computed, ref, shallowRef, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useBusinessStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import bus from '@/common/bus';
// import types
import { ApplyClbModel, SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { reqResourceListOfCurrentRegion } from '@/api/load_balancers/apply-clb';
import { ClbQuota, LbPrice } from '@/typings';
import { cloneDeep, debounce, uniqBy } from 'lodash';

// 当云地域变更时, 获取用户在当前地域支持可用区列表和资源列表
export default (formModel: ApplyClbModel) => {
  const { isBusinessPage } = useWhereAmI();
  const businessStore = useBusinessStore();
  // define data
  const isResourceListLoading = ref(false); // 是否正在获取资源列表
  const currentResourceListMap = ref<{ [key: string]: any }>(); // 资源映射
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
    IPV4: 'IPv4',
    IPv6FullChain: 'IPv6',
    IPV6: 'IPv6_Nat',
  };

  /**
   * 获取当前地域「可用区列表和资源列表的映射关系」
   * @param region 地域
   */
  const getResourceListOfCurrentRegion = async (params: any) => {
    isResourceListLoading.value = true;
    try {
      const { data } = await reqResourceListOfCurrentRegion(params);
      const { ZoneResourceSet } = data;

      // 重置资源映射
      currentResourceListMap.value = {};

      // 构建资源列表映射
      ZoneResourceSet.forEach(({ MasterZone, SlaveZone, ResourceSet }) => {
        const key = `${MasterZone}|${SlaveZone || ''}`.toLowerCase();
        const resource = currentResourceListMap.value[key] || { ispList: [] };

        ResourceSet?.forEach(({ Isp, TypeSet, AvailabilitySet }) => {
          const ispEntries = AvailabilitySet
            ? AvailabilitySet.map((curr) => ({ Isp, ...curr })) // 如果有AvailabilitySet, 则Isp可用性取决于它
            : [{ Isp, Availability: 'Available' }]; // 如果没有AvailabilitySet, 则默认Isp可用

          resource.ispList.push(...ispEntries);

          // 存储性能容量型资源
          resource[Isp] = resource[Isp] || [];
          TypeSet?.forEach(({ SpecAvailabilitySet }) => {
            resource[Isp].push(...SpecAvailabilitySet);
          });
        });

        currentResourceListMap.value[key] = resource;
      });

      // 处理 ispList 去重并聚合 TypeSet
      Object.keys(currentResourceListMap.value).forEach((key) => {
        const { ispList } = currentResourceListMap.value[key];
        const newIspList = aggregateIspList(ispList);
        currentResourceListMap.value[key].ispList = newIspList;
      });

      // 构建 noZone 对象
      const noZone = constructNoZone(currentResourceListMap.value);

      currentResourceListMap.value.noZone = noZone;
    } finally {
      isResourceListLoading.value = false;
    }
  };

  // 聚合 ispList 并去重
  const aggregateIspList = (ispList: any[]) => {
    const newIspList = ispList.reduce((acc: any[], { Isp, Type = Isp, Availability = 'Available' }: any) => {
      const typeObj = { Type, Availability };
      const existing = acc.find((item) => item.Isp === Isp);

      if (existing) {
        existing.TypeSet.push(typeObj);
      } else {
        acc.push({ Isp, TypeSet: [typeObj] });
      }

      return acc;
    }, []);

    // TypeSet 去重
    newIspList.forEach((isp) => (isp.TypeSet = uniqBy(isp.TypeSet, JSON.stringify)));

    return newIspList;
  };

  // 构建 noZone 对象
  const constructNoZone = (resourceMap: any) => {
    const noZone: Record<string, any> = {};

    Object.keys(resourceMap).forEach((key) => {
      const target = cloneDeep(resourceMap[key]);

      Object.keys(target).forEach((subKey) => {
        noZone[subKey] = noZone[subKey] || [];
        noZone[subKey].push(...target[subKey]);
      });
    });

    Object.keys(noZone).forEach((key) => {
      if (key === 'ispList') {
        noZone[key] = uniqBy(noZone[key], 'Isp').map((isp: any) => {
          isp.TypeSet = isp.TypeSet.map((type: any) => ({
            ...type,
            Availability: 'Available',
          }));
          return isp;
        });
      } else {
        noZone[key] = uniqBy(noZone[key], 'SpecType').map((specType) => ({
          ...specType,
          Availability: 'Available',
        }));
      }
    });

    return noZone;
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
      // eslint-disable-next-line prefer-const
      let zones = formModel.zones ? [formModel.zones] : [];

      const { data } = await businessStore.lbPricesInquiry({
        ...formModel,
        bk_biz_id: isBusinessPage ? formModel.bk_biz_id : undefined,
        zones,
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
      // 当云地域变更时, 获取当前地域「腾讯云账号负载均衡的配额」
      getLbQuotas(val);
    },
  );

  watch(
    [() => formModel.region, () => formModel.zones, () => formModel.address_ip_version],
    ([region, zones, address_ip_version]) => {
      // 内网下不需要选择运营商类型
      if (!region || formModel.load_balancer_type === 'INTERNAL') return;

      let master_zone;
      if (zones) master_zone = Array.isArray(zones) && zones.length > 0 ? zones : [zones];
      const params = {
        account_id: formModel.account_id,
        region,
        master_zone,
        ip_version: [ipVersionMap[address_ip_version]],
      };

      // 获取当前地域「可用区列表和资源列表的映射关系」
      getResourceListOfCurrentRegion(params);
    },
    { deep: true },
  );

  watch(
    currentResourceListMap,
    (val) => {
      const { zones, backup_zones, address_ip_version, region } = formModel;
      if (!zones && !backup_zones) {
        // 只选择地域
        ispList.value = val.noZone?.ispList?.filter(({ Isp }: { Isp: string }) => Isp !== 'INTERNAL') || [];
      } else {
        // 拼接key, 用于定位对应的 isp 列表
        const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
        const key = `${zonesRule || ''}|${backup_zones || ''}`.toLowerCase();
        // 内网下的 isp 选项不显示
        ispList.value = val[key]?.ispList?.filter(({ Isp }: { Isp: string }) => Isp !== 'INTERNAL') || [];
      }
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
      const { zones, backup_zones, address_ip_version, vip_isp, region } = formModel;
      specAvailabilitySet.value = [];
      if (!vip_isp) {
        formModel.sla_type = 'shared';
        return;
      }
      if (!zones && !backup_zones) {
        // 只选择地域
        specAvailabilitySet.value = currentResourceListMap.value.noZone[vip_isp];
      } else {
        const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
        const key = `${zonesRule || ''}|${backup_zones || ''}`.toLowerCase();
        // 公有云TypeSet数组暂时取第一个元素, 忽略 Type 的作用, 直接取 SpecAvailabilitySet 作为性能容量型的机型选择
        specAvailabilitySet.value = currentResourceListMap.value[key][vip_isp];
      }
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
