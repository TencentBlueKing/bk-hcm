import { PropType, computed, defineComponent, ref, watch, watchEffect } from 'vue';
import { Select, Tag } from 'bkui-vue';
import './index.scss';
import { useBusinessStore } from '@/store';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';

const { Option } = Select;

export default defineComponent({
  name: 'PrimaryStandZoneSelector',
  props: {
    modelValue: Array as PropType<string[]>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
  },
  emits: ['update:modelValue'],
  setup(props, ctx) {
    const businessStore = useBusinessStore();

    const primaryZone = ref('');
    const standZone = ref('');
    const selectedValue = computed(() => {
      return [primaryZone.value, standZone.value];
    });
    const handleSelect = (value: string) => {
      if (primaryZone.value && !standZone.value) {
        standZone.value = value;
      } else {
        primaryZone.value = value;
      }
      ctx.emit('update:modelValue', selectedValue.value);
    };
    const handleDeSelect = (value: string) => {
      if (primaryZone.value === value) {
        primaryZone.value = '';
      } else {
        standZone.value = '';
      }
      ctx.emit('update:modelValue', selectedValue.value);
    };
    const handleExchange = (e: MouseEvent) => {
      e.stopPropagation();
      const temp = primaryZone.value;
      primaryZone.value = standZone.value;
      standZone.value = temp;
      ctx.emit('update:modelValue', selectedValue.value);
    };

    const zonesList = ref([]);
    const loading = ref(null);
    const zonePage = ref(0);
    const hasMoreData = ref(true);
    const filter = ref<QueryFilterType>({ op: 'and', rules: [] });
    const getZonesData = async () => {
      if (!hasMoreData.value || !props.vendor || !props.region) return;
      loading.value = true;
      const res = await businessStore.getZonesList({
        vendor: props.vendor,
        region: props.region,
        data: {
          filter: filter.value,
          page: {
            start: zonePage.value * 100,
            limit: 100,
          },
        },
      });
      zonePage.value += 1;
      zonesList.value.push(...(res?.data?.details || []));
      hasMoreData.value = res?.data?.details?.length >= 100; // 100条数据说明还有数据 可翻页
      loading.value = false;
    };
    const resetData = () => {
      zonePage.value = 0;
      hasMoreData.value = true;
      zonesList.value = [];
    };

    watchEffect(
      void (async () => {
        getZonesData();
      })(),
    );

    watch(
      () => props.vendor,
      (val) => {
        switch (val) {
          case VendorEnum.TCLOUD:
            filter.value.rules = [
              { field: 'vendor', op: QueryRuleOPEnum.EQ, value: val },
              { field: 'state', op: QueryRuleOPEnum.EQ, value: 'AVAILABLE' },
            ];
            break;
          default:
            filter.value.rules = [];
        }
        resetData();
        getZonesData();
      },
    );

    watch(
      () => props.region,
      () => {
        resetData();
        getZonesData();
      },
    );

    return () => (
      <Select
        class='primary-stand-zone-selector'
        modelValue={selectedValue.value}
        multiple
        multipleMode='tag'
        filterable
        inputSearch={false}
        clearable={false}
        onSelect={handleSelect}
        onDeselect={handleDeSelect}
        loading={loading.value}
        onScroll-end={getZonesData}>
        {{
          default: () => {
            return zonesList.value.map((item) => {
              return (
                <Option key={item.id} id={item.name} name={item.name_cn || item.name}>
                  <span>{item.name_cn || item.name}</span>
                  {primaryZone.value === item.name && (
                    <Tag class='ml12' theme='info'>
                      主可用区
                    </Tag>
                  )}
                  {standZone.value === item.name && (
                    <Tag class='ml12' theme='warning'>
                      备可用区
                    </Tag>
                  )}
                </Option>
              );
            });
          },
          tag: () => (
            <div class='selected-tag-value-container'>
              <Tag closable>主&nbsp;&nbsp;:&nbsp;&nbsp;{primaryZone.value || '请选择'}</Tag>
              <i class='hcm-icon bkhcm-icon-exchange-line' onClick={handleExchange}></i>
              <Tag closable>备&nbsp;&nbsp;:&nbsp;&nbsp;{standZone.value || '请选择'}</Tag>
            </div>
          ),
        }}
      </Select>
    );
  },
});
