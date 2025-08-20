import { ref, watch } from 'vue';
import { useBusinessStore } from '@/store';
import { cloneDeep, isEqual, uniqBy } from 'lodash';
import { VendorEnum } from '@/common/constant';
import { reqResourceListOfCurrentRegion } from '@/api/load_balancers/apply-clb';
import type { ApplyClbModel, SpecAvailability, ZoneResource } from '@/api/load_balancers/apply-clb/types';
import type { ClbQuota } from '@/typings';
import { BGP_VIP_ISP_TYPES } from '@/constants';

// 类型别名，提高代码可读性
type ResourceMap = Record<string, ResourceMapItem>;
type IspEntry = {
  Isp: string;
  Type?: string;
  Availability?: string;
  TypeSet?: Array<{ Type: string; Availability: string }>;
};
type SpecAvailabilityList = SpecAvailability[];

interface ResourceMapItem {
  ispList: IspEntry[];
  masterZone: string;
  slaveZone: string | null;
  slaveZoneOptions: string[];
  withoutIsp: SpecAvailabilityList;
  [isp: string]: any; // 动态属性
}

// 当云地域变更时, 获取用户在当前地域支持可用区列表和资源列表
export default (formModel: ApplyClbModel) => {
  const businessStore = useBusinessStore();

  // 定义响应式数据
  const isResourceListLoading = ref(false);
  const currentResourceListMap = ref<ResourceMap>({});
  const ispList = ref<IspEntry[]>([]);
  const specAvailabilitySet = ref<SpecAvailabilityList>([]);
  const quotas = ref<ClbQuota[]>([]);

  // 前端IP版本映射
  const ipVersionMap = {
    IPV4: 'IPv4',
    IPv6FullChain: 'IPv6',
    IPV6: 'IPv6_Nat',
  };

  // 地域变更时，获取配额信息
  watch(
    () => formModel.region,
    (val) => {
      if (!val) return;
      // 当云地域变更时, 获取当前地域「腾讯云账号负载均衡的配额」
      getLbQuotas(val);
    },
  );
  const getLbQuotas = async (region: string) => {
    try {
      const { data } = await businessStore.getClbQuotas({
        account_id: formModel.account_id,
        region,
      });
      quotas.value = data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  // 监听地域、可用区和IP版本变更时，获取当前地域「可用区列表和资源列表的映射关系」
  watch(
    [() => formModel.region, () => formModel.zones, () => formModel.address_ip_version],
    ([newRegion, newZones, newAddressIpVersion], [oldRegion, oldZones, oldAddressIpVersion]) => {
      if (
        isEqual(newRegion, oldRegion) &&
        isEqual(newZones, oldZones) &&
        isEqual(newAddressIpVersion, oldAddressIpVersion)
      ) {
        return;
      }

      if (!newRegion) return;

      let master_zone;
      if (newZones) master_zone = Array.isArray(newZones) ? newZones : [newZones];
      const params = {
        account_id: formModel.account_id,
        region: newRegion,
        master_zone,
        ip_version: newAddressIpVersion ? [ipVersionMap[newAddressIpVersion]] : undefined,
      };

      // 获取当前地域「可用区列表和资源列表的映射关系」
      getResourceListOfCurrentRegion(formModel.vendor, params);
    },
  );
  const getResourceListOfCurrentRegion = async (vendor: VendorEnum, params: any) => {
    currentResourceListMap.value = {}; // 重置资源映射
    isResourceListLoading.value = true;
    try {
      const { data } = await reqResourceListOfCurrentRegion(vendor, params);
      const { ZoneResourceSet = [] } = data;

      // 构建主备可用区映射
      const zoneMapping = buildZoneMapping(ZoneResourceSet);

      // 处理每个资源项
      ZoneResourceSet.forEach((item) => processResourceSet(item, zoneMapping));

      // 聚合ISP列表
      Object.keys(currentResourceListMap.value).forEach((key) => {
        const { ispList } = currentResourceListMap.value[key];
        currentResourceListMap.value[key].ispList = resolveIspList(ispList);
      });

      // 构建无可用区资源
      currentResourceListMap.value.withoutZone = buildResourceMapItemWithoutZone(currentResourceListMap.value);
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isResourceListLoading.value = false;
    }
  };
  const buildZoneMapping = (ZoneResourceSet: ZoneResource[]) => {
    const zoneMapping: Record<string, string[]> = {};
    ZoneResourceSet.forEach(({ MasterZone, SlaveZone }) => {
      if (!zoneMapping[MasterZone]) {
        zoneMapping[MasterZone] = [];
      }
      if (SlaveZone) {
        zoneMapping[MasterZone].push(SlaveZone);
      }
    });
    return zoneMapping;
  };
  const processResourceSet = (item: ZoneResource, zoneMapping: Record<string, string[]>) => {
    const { MasterZone, SlaveZone, ResourceSet } = item;
    const key = `${MasterZone}|${SlaveZone || ''}`.toLowerCase();

    const resource: ResourceMapItem = currentResourceListMap.value[key] || {
      ispList: [],
      masterZone: MasterZone,
      slaveZone: SlaveZone || null,
      slaveZoneOptions: zoneMapping[MasterZone] || [],
      withoutIsp: [],
    };

    ResourceSet?.forEach(({ Isp, Type, TypeSet, AvailabilitySet }) => {
      const ispEntries = Type.map((type) => {
        const availability = AvailabilitySet?.find((curr) => curr.Type === type);
        return { Isp, Type: type, Availability: availability?.Availability || 'Available' };
      });

      resource.ispList.push(...ispEntries);

      // 初始化ISP资源
      if (!resource[Isp]) resource[Isp] = [];

      // 处理性能容量型资源
      TypeSet?.forEach(({ SpecAvailabilitySet }) => {
        const filteredSpecs = SpecAvailabilitySet.filter((spec) => spec.SpecType !== 'clb.c1.small');
        resource[Isp].push(...filteredSpecs);
        // 内网下无ISP资源，这里单独存储一份
        resource.withoutIsp.push(...filteredSpecs);
      });
    });

    resource.withoutIsp = resolveSpecListBySpecType(resource.withoutIsp); // 去重
    currentResourceListMap.value[key] = resource;
  };
  const resolveIspList = (ispList: IspEntry[]) => {
    return ispList
      .reduce((acc, { Isp, Type, Availability }) => {
        const typeObj = { Type, Availability };
        const existing = acc.find((item) => item.Isp === Isp);

        if (existing) {
          existing.TypeSet.push(typeObj);
        } else {
          acc.push({ Isp, TypeSet: [typeObj] });
        }
        return acc;
      }, [])
      .map((isp) => ({
        ...isp,
        TypeSet: uniqBy(isp.TypeSet, JSON.stringify),
      }));
  };
  const buildResourceMapItemWithoutZone = (resourceMap: ResourceMap) => {
    const resourceMapItem: ResourceMapItem = {
      ispList: [],
      masterZone: '',
      slaveZone: '',
      slaveZoneOptions: [],
      withoutIsp: [],
    };

    Object.keys(resourceMap).forEach((key) => {
      const target = cloneDeep(resourceMap[key]);
      Object.keys(target).forEach((subKey) => {
        resourceMapItem[subKey] = resourceMapItem[subKey] || [];
        resourceMapItem[subKey].push(...(target[subKey] || []));
      });
    });

    // 特殊处理ispList
    if (resourceMapItem.ispList) {
      resourceMapItem.ispList = uniqBy(resourceMapItem.ispList, 'Isp').map((isp) => ({
        ...isp,
        TypeSet: isp.TypeSet.map((type) => ({ ...type, Availability: 'Available' })),
      }));
    }

    // 处理其他属性
    Object.keys(resourceMapItem)
      .filter((key) => key !== 'ispList')
      .forEach((key) => {
        resourceMapItem[key] = uniqBy(resourceMapItem[key], 'SpecType').map((specType) => ({
          ...specType,
          Availability: 'Available',
        }));
      });

    return resourceMapItem;
  };
  const resolveSpecListBySpecType = (list: SpecAvailability[]): SpecAvailability[] => {
    const specTypeMap = new Map<string, SpecAvailability>();

    list.forEach((item) => {
      const existing = specTypeMap.get(item.SpecType);
      if (!existing || item.Availability === 'Available') {
        specTypeMap.set(item.SpecType, item);
      }
    });

    return Array.from(specTypeMap.values());
  };

  // 监听资源映射和备可用区变化，更新Isp列表
  watch(
    [currentResourceListMap, () => formModel.backup_zones],
    ([map, backupZones]) => {
      const { zones, address_ip_version, region } = formModel;
      if (!zones && !backupZones) {
        // 只选择地域
        ispList.value = map.withoutZone?.ispList?.filter(({ Isp }) => Isp !== 'INTERNAL') || [];
      } else {
        // 拼接key, 用于定位对应的 isp 列表
        const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
        const key = `${zonesRule || ''}|${backupZones || ''}`.toLowerCase();
        // 内网下的 isp 选项不显示
        ispList.value = map[key]?.ispList?.filter(({ Isp }) => Isp !== 'INTERNAL') || [];
      }
    },
    { deep: true },
  );

  // 监听ISP列表变化，设置默认ISP，负载均衡规格类型
  watch(
    ispList,
    (val) => {
      // 优先选择BGP，否则选择第一个可用的ISP
      const availableBGP = val.some(
        ({ Isp, TypeSet }) => Isp === 'BGP' && TypeSet?.some(({ Availability }) => Availability === 'Available'),
      );

      const firstAvailableISP = val.find(({ TypeSet }) =>
        TypeSet?.some(({ Availability }) => Availability === 'Available'),
      )?.Isp;

      formModel.vip_isp = val.length && availableBGP ? 'BGP' : firstAvailableISP ?? '';
      formModel.slaType = '0';
      formModel.sla_type = 'shared';
    },
    { deep: true },
  );

  // 监听vip_isp变化，更新负载均衡规格类型列表，计费模式
  watch([() => formModel.vip_isp, () => formModel.load_balancer_type], ([vipIsp, loadBalancerType]) => {
    if (!vipIsp) {
      formModel.sla_type = 'shared';
      return;
    }

    buildSpecAvailabilitySet(vipIsp, loadBalancerType);

    // 设置计费类型：clb运营商选三网（电信、移动、联通）时，只能选共享带宽包
    if (loadBalancerType === 'OPEN') {
      formModel.internet_charge_type = BGP_VIP_ISP_TYPES.includes(vipIsp)
        ? 'TRAFFIC_POSTPAID_BY_HOUR'
        : 'BANDWIDTH_PACKAGE';
    } else {
      formModel.internet_charge_type = undefined;
    }
  });
  const buildSpecAvailabilitySet = (isp: string, loadBalancerType: ApplyClbModel['load_balancer_type']) => {
    specAvailabilitySet.value = [];
    const { zones, backup_zones, address_ip_version, region } = formModel;

    // 内网下不区分ISP
    const specTypeKeyPath = loadBalancerType === 'OPEN' ? isp : 'withoutIsp';
    if (!zones && !backup_zones) {
      // 只选择地域
      specAvailabilitySet.value = currentResourceListMap.value.withoutZone[specTypeKeyPath] || [];
    } else {
      const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
      const key = `${zonesRule || ''}|${backup_zones || ''}`.toLowerCase();
      // 公有云TypeSet数组暂时取第一个元素, 忽略 Type 的作用, 直接取 SpecAvailabilitySet 作为性能容量型的机型选择
      specAvailabilitySet.value = currentResourceListMap.value[key]?.[specTypeKeyPath] || [];
    }
  };

  // 监听规格可用性变化
  watch(
    specAvailabilitySet,
    () => {
      formModel.slaType = '0';
      formModel.sla_type = 'shared';
    },
    { deep: true },
  );

  return {
    ispList,
    isResourceListLoading,
    quotas,
    currentResourceListMap,
    specAvailabilitySet,
  };
};
