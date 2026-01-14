import { ref, watch } from 'vue';
import apiService from '@/api/scrApi';
import { ZoneHook } from './use-zone-factory';

export const useZoneZiyan: ZoneHook = (params) => {
  const list = ref([]);
  const loading = ref(false);

  const getZoneData = async (resourceType: string, regionVal: string | string[]) => {
    const region = Array.isArray(regionVal) ? regionVal : [regionVal];

    try {
      if (['QCLOUDCVM', 'QCLOUDDVM'].includes(resourceType)) {
        const { info = [] } = await apiService.getZones({ vendor: 'qcloud', region }, {});
        list.value = info.map((item: any) => {
          return {
            id: item.zone,
            name: `${item.zone_cn}${item.cmdb_zone_name ? `(${item.cmdb_zone_name})` : ''}`,
          };
        });
      } else if (['IDCDVM', 'IDCPM'].includes(resourceType)) {
        const { info = [] } = await apiService.getZones({ vendor: 'idc', region }, {});
        list.value = info.map((item: any) => {
          return {
            id: item.cmdb_zone_name,
            name: item.cmdb_zone_name,
          };
        });
      }
    } catch {
      list.value = [];
    } finally {
      loading.value = false;
    }
  };

  watch(
    [() => params.resourceType, () => params.region],
    ([resourceType, region]) => {
      list.value = [];
      if (resourceType && region) {
        getZoneData(resourceType, region);
      }
    },
    { immediate: true },
  );

  return {
    list,
    loading,
  };
};
