import { ref, watch } from 'vue';
import apiService from '@/api/scrApi';

import { type FactoryType } from './data-common';
import type { IZoneTagSelectorProps } from './index.vue';

const useList = (props: IZoneTagSelectorProps) => {
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

      if (props.separateCampus && region.length) {
        list.value.push({
          id: 'cvm_separate_campus',
          name: '分Campus',
        });
      }
    } catch {
      list.value = [];
    } finally {
      loading.value = false;
    }
  };

  watch(
    [() => props.resourceType, () => props.region],
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

const dataZiyan: FactoryType = {
  useList,
};

export default dataZiyan;
