import { defineComponent, onMounted, reactive, ref } from 'vue';
import { Divider, Select } from 'bkui-vue';
import { Plus, RightTurnLine, Spinner } from 'bkui-vue/lib/icon';
import http from '@/http';
import {
  BANDWIDTH_PACKAGE_CHARGE_TYPE_MAP,
  BANDWIDTH_PACKAGE_NETWORK_TYPE_MAP,
  BANDWIDTH_PACKAGE_STATUS,
} from '@/constants';

const { Option } = Select;

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

/**
 * 共享带宽包选择器
 * @todo 由于调用云上接口, 与本地接口协议不同, 所以不使用 useSingleList 获取 options 数据, 后续看是否可以在 useSingleList 中进行优化
 */
export default defineComponent({
  name: 'BandwidthPackageSelector',
  props: {
    accountId: String,
    region: String,
  },
  setup(props) {
    const getDefaultPage = () => ({ offset: 0, limit: 50 });

    const bandwidthPackageList = ref([]);
    const totalCount = ref(0);
    const page = reactive(getDefaultPage());
    const isDataLoad = ref(false);
    const isDataRefresh = ref(false);

    const getBandwidthPackageList = async () => {
      isDataLoad.value = true;
      try {
        const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bandwidth_packages/query`, {
          account_id: props.accountId,
          region: props.region,
          page,
        });
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

    onMounted(() => {
      getBandwidthPackageList();
    });

    return () => (
      <Select
        class='bandwidth-package-selector w500'
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {{
          default: () =>
            bandwidthPackageList.value.map(({ id, name, charge_type, network_type, status }) => (
              <Option
                key={id}
                id={id}
                name={`${name}(${BANDWIDTH_PACKAGE_CHARGE_TYPE_MAP[charge_type] || charge_type} ${
                  BANDWIDTH_PACKAGE_NETWORK_TYPE_MAP[network_type]
                })`}
                disabled={status !== BANDWIDTH_PACKAGE_STATUS.CREATED}
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
