import { defineComponent, PropType, reactive, ref, watch } from 'vue';
import { Divider, Select } from 'bkui-vue';
import { Plus, RightTurnLine, Spinner } from 'bkui-vue/lib/icon';
import http from '@/http';
import {
  BANDWIDTH_PACKAGE_CHARGE_TYPE_MAP,
  BANDWIDTH_PACKAGE_NETWORK_TYPE_MAP,
  BANDWIDTH_PACKAGE_STATUS,
} from '@/constants';
import { IQueryResData } from '@/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { Option } = Select;

interface IBandwidthPackage {
  id: string;
  name: string;
  network_type: string;
  charge_type: string;
  status: string;
  bandwidth: number;
  egress: string;
  create_time: string;
  deadline: string;
  resource_set: {
    resource_type: string;
    resource_id: string;
    address_ip: string;
  }[];
}

/**
 * 共享带宽包选择器
 * @todo 由于调用云上接口, 与本地接口协议不同, 所以不使用 useSingleList 获取 options 数据, 后续看是否可以在 useSingleList 中进行优化
 */
export default defineComponent({
  name: 'BandwidthPackageSelector',
  props: {
    accountId: String,
    region: String,
    // 带宽包的运营商（network_type字段）要和clb的运营商一致，其中如果是SINGLEISP，代表 电信/联通/移动都可以（先确定clb的运营商、过滤带宽包）
    networkTypes: Object as PropType<string[]>,
    // 带宽包是同region下的zone可用，上海的南昌一区特殊处理，通过带宽包的egress字段，只能选择值前缀为edge的带宽包；反过来说，上海的可用区也不能选择这个特殊的带宽包。
    zones: String,
    resourceType: Object as PropType<ResourceTypeEnum>,
  },
  setup(props) {
    const getDefaultPage = () => ({ offset: 0, limit: 50 });
    const { getBusinessApiPath } = useWhereAmI();

    const bandwidthPackageList = ref<IBandwidthPackage[]>([]);
    const totalCount = ref(0);
    const page = reactive(getDefaultPage());
    const isDataLoad = ref(false);
    const isDataRefresh = ref(false);

    const getBandwidthPackageList = async () => {
      const { accountId, region, networkTypes } = props;
      if (!accountId || !region) return;

      isDataLoad.value = true;
      try {
        const res: IQueryResData<{ packages: IBandwidthPackage[]; total_count: number }> = await http.post(
          `/api/v1/cloud/${getBusinessApiPath()}bandwidth_packages/query`,
          {
            account_id: accountId,
            region,
            network_types: networkTypes,
            page,
          },
        );

        bandwidthPackageList.value = res.data.packages;
        totalCount.value = res.data.total_count;
      } finally {
        isDataLoad.value = false;
      }
    };

    const handleScrollEnd = () => {
      if (bandwidthPackageList.value.length >= totalCount.value) return;
      page.offset += page.limit;
      getBandwidthPackageList();
    };

    const handleReset = () => {
      bandwidthPackageList.value = [];
      Object.assign(page, getDefaultPage());
    };

    const handleRefresh = async () => {
      handleReset();
      isDataRefresh.value = true;
      await getBandwidthPackageList();
      isDataRefresh.value = false;
    };

    // 检查带宽包可用性
    const checkBandwidthPackageAvailable = (egress: string) => {
      if (props.resourceType !== ResourceTypeEnum.CLB) return true;

      // 只对CLB资源的上海可用区进行判断
      const isShanghai = props.region === 'ap-shanghai';
      const isNanchangZone = props.zones === 'ap-shanghai-ez-nanchang-1';

      if (isShanghai) {
        return isNanchangZone ? egress.startsWith('edge') : !egress.startsWith('edge');
      }

      return true;
    };

    watch(
      [() => props.accountId, () => props.region, () => props.networkTypes],
      () => {
        getBandwidthPackageList();
      },
      { immediate: true },
    );

    return () => (
      <Select
        class='bandwidth-package-selector w500'
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {{
          default: () =>
            bandwidthPackageList.value.map(({ id, name, charge_type, network_type, status, egress }) => (
              <Option
                key={id}
                id={id}
                name={`${name}(${id}) (${BANDWIDTH_PACKAGE_CHARGE_TYPE_MAP[charge_type] || charge_type} ${
                  BANDWIDTH_PACKAGE_NETWORK_TYPE_MAP[network_type]
                })`}
                disabled={status !== BANDWIDTH_PACKAGE_STATUS.CREATED || !checkBandwidthPackageAvailable(egress)}
              />
            )),
          extension: () => (
            <div style='width: 100%; color: #63656E; padding: 0 12px;'>
              <div style='display: flex; align-items: center;justify-content: center;'>
                <span style='display: flex; align-items: center;cursor: pointer;' onClick={() => {}}>
                  <Plus style='font-size: 20px;' />
                  新增
                </span>
                <span style='display: flex; align-items: center;position: absolute; right: 12px;'>
                  <Divider direction='vertical' type='solid' />
                  {isDataRefresh.value ? (
                    <Spinner style='font-size: 14px;color: #3A84FF;' />
                  ) : (
                    <RightTurnLine style='font-size: 14px;cursor: pointer;' onClick={handleRefresh} />
                  )}
                </span>
              </div>
            </div>
          ),
        }}
      </Select>
    );
  },
});
