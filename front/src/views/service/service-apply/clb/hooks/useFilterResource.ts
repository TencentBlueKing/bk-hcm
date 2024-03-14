import { reactive, ref, watch } from 'vue';
// import types
import { ApplyClbModel, ResourceOfCurrentRegionResp, SpecAvailability } from '@/api/load_balancers/apply-clb/types';
import { reqResourceListOfCurrentRegion } from '@/api/load_balancers/apply-clb';
import { Message } from 'bkui-vue';
import bus from '@/common/bus';

// 当云地域变更时, 获取用户在当前地域支持可用区列表和资源列表
export default (formModel: ApplyClbModel) => {
  // define data
  const currentResourceListMap = reactive({}); // 资源映射
  const ispList = ref([]); // 运营商类型
  const specAvailabilitySet = ref<Array<SpecAvailability>>([]); // 负载均衡规格类型

  // 前端改映射
  const ipVersionMap = {
    IPV4: 'ipv4',
    IPv6FullChain: 'ipv6',
    IPV6: 'ipv6_nat',
  };

  watch(
    () => formModel.region,
    (val) => {
      if (!val) return;
      // 当云地域变更时, 获取当前地域「可用区列表和资源列表的映射关系」
      reqResourceListOfCurrentRegion({ account_id: formModel.account_id, region: val }).then(
        ({ data }: ResourceOfCurrentRegionResp) => {
          const { ZoneResourceSet } = data;
          ZoneResourceSet.forEach(({ MasterZone, SlaveZone, IPVersion, ResourceSet }) => {
            // '主可用区|备可用区|IP版本' 为对象的 key, [{ Isp, TypeSet }, { Isp, TypeSet }...] 为对象的 value
            const key = `${MasterZone}|${SlaveZone}|${IPVersion}`.toLowerCase();

            ResourceSet?.forEach(({ Isp, TypeSet }) => {
              currentResourceListMap[key] = currentResourceListMap[key] || {};
              currentResourceListMap[key][Isp] = currentResourceListMap[key][Isp] || [];

              TypeSet?.forEach(({ SpecAvailabilitySet }) => {
                currentResourceListMap[key][Isp].push(...SpecAvailabilitySet);
              });
            });
          });
        },
      );
    },
  );

  watch(
    [() => formModel.zones, () => formModel.backup_zones, () => formModel.address_ip_version],
    () => {
      const { zones, backup_zones, address_ip_version, region } = formModel;
      // 拼接key, 用于定位对应的 isp 列表
      const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
      const key = `${zonesRule || null}|${backup_zones || null}|${ipVersionMap[address_ip_version]}`.toLowerCase();
      ispList.value = Object.keys(currentResourceListMap[key] || {}).filter((isp) => isp !== 'INTERNAL');
    },
    {
      deep: true,
    },
  );

  watch(
    () => formModel.vip_isp,
    () => {
      const { zones, backup_zones, address_ip_version, vip_isp, region } = formModel;
      if (!vip_isp) {
        specAvailabilitySet.value = [];
        formModel.sla_type = 'shared';
        return;
      }
      const zonesRule = address_ip_version !== 'IPV4' ? region : zones;
      const key = `${zonesRule || null}|${backup_zones || null}|${ipVersionMap[address_ip_version]}`.toLowerCase();
      // 公有云TypeSet数组暂时取第一个元素, 忽略 Type 的作用, 直接取 SpecAvailabilitySet 作为性能容量型的机型选择
      specAvailabilitySet.value = currentResourceListMap[key][vip_isp];
    },
  );

  watch(specAvailabilitySet, (val) => {
    if (val) {
      bus.$emit(
        'updateSpecAvailabilitySet',
        val.filter(({ SpecType }) => SpecType !== 'shared'),
      );
    } else {
      Message({ theme: 'warning', message: '当前地域下无可用规格, 请切换地域' });
    }
  });

  return {
    ispList,
  };
};
