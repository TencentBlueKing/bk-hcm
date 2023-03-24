import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import { IOption, QueryFilterType, QueryRuleOPEnum } from '@/typings/common';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    type: String as PropType<string>,
    vendor: String as PropType<string>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const list = ref([]);
    const loading = ref(false);

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    watch([() => props.vendor], async ([vendor]) => {
      if (!vendor) {
        list.value = [];
        return;
      }

      const filter: QueryFilterType = {
        op: 'and',
        rules: [],
      };
      let dataIdKey = 'region_id';
      let dataNameKey = 'region_name';
      switch (vendor) {
        case VendorEnum.AZURE:
          filter.rules = [
            {
              field: 'type',
              op: QueryRuleOPEnum.EQ,
              value: 'Region',
            },
          ];
          dataIdKey = 'name';
          dataNameKey = 'display_name';
          break;
        case VendorEnum.HUAWEI: {
          const services = {
            [ResourceTypeEnum.CVM]: 'ecs',
            [ResourceTypeEnum.VPC]: 'vpc',
            [ResourceTypeEnum.DISK]: 'ecs',
          };
          filter.rules = [
            {
              field: 'type',
              op: QueryRuleOPEnum.EQ,
              value: 'public',
            },
            {
              field: 'service',
              op: QueryRuleOPEnum.EQ,
              value: services[props.type],
            },
          ];
          dataNameKey = 'region_id';
          break;
        }
        case VendorEnum.TCLOUD:
        case VendorEnum.AWS:
        case VendorEnum.GCP:
          filter.rules = [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: vendor,
            },
          ];
          break;
      }

      loading.value = true;
      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${vendor}/regions/list`, {
        filter,
        page: {
          count: false,
          start: 0,
          limit: 500,
        },
      });

      const details = result?.data?.details ?? [];
      list.value = details
        .filter((_item: any) => {
          // if (item?.status) {
          //   return item.status === 'AVAILABLE';
          // }
          return true;
        })
        .map((item: any) => ({
          id: item[dataIdKey],
          name: item[dataNameKey],
        }));

      loading.value = false;
    });

    return () => (
      <Select
        clearable={false}
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        loading={loading.value}
        {...{ attrs }}
      >
        {
          list.value.map(({ id, name }: IOption) => (
            <Option key={id} value={id} label={name}></Option>
          ))
        }
      </Select>
    );
  },
});
