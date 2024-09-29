import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import { IOption, QueryFilterType, QueryRuleOPEnum } from '@/typings/common';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import { useHostStore } from '@/store/host';
import { isChinese } from '@/language/i18n';
import { getRegionName } from '@pluginHandler/region-selector';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: [String, Array] as PropType<string | string[]>,
    multiple: Boolean,
    type: String as PropType<string>,
    vendor: String as PropType<string>,
    accountId: String as PropType<string>,
    isDisabled: {
      required: false,
      default: false,
      type: Boolean,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const list = ref([]);
    const loading = ref(false);
    const hostStore = useHostStore();

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    watch(
      [() => props.vendor],
      async ([vendor]) => {
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
              [ResourceTypeEnum.SUBNET]: 'vpc',
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
            dataNameKey = isChinese ? 'locales_zh_cn' : 'region_id';
            break;
          }
          case VendorEnum.TCLOUD: {
            filter.rules = [
              {
                field: 'vendor',
                op: QueryRuleOPEnum.EQ,
                value: vendor,
              },
              {
                field: 'status',
                op: QueryRuleOPEnum.EQ,
                value: 'AVAILABLE',
              },
            ];
            dataNameKey = isChinese ? 'region_name' : 'display_name';
            break;
          }
          case VendorEnum.AWS: {
            filter.rules = [
              {
                field: 'vendor',
                op: QueryRuleOPEnum.EQ,
                value: vendor,
              },
              {
                field: 'status',
                op: QueryRuleOPEnum.EQ,
                value: 'opt-in-not-required',
              },
              // {
              //   field: 'account_id',
              //   op: QueryRuleOPEnum.EQ,
              //   value: props.accountId,
              // },
            ];
            break;
          }
          case VendorEnum.GCP:
            filter.rules = [
              {
                field: 'vendor',
                op: QueryRuleOPEnum.EQ,
                value: vendor,
              },
              {
                field: 'status',
                op: QueryRuleOPEnum.EQ,
                value: 'UP',
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
        list.value = details.map((item: any) => ({
          id: item[dataIdKey],
          name: getRegionName(isChinese, vendor as VendorEnum, item[dataIdKey], item[dataNameKey]) || item[dataIdKey],
        }));
        hostStore.regionList = details;

        loading.value = false;
      },
      {
        immediate: true,
      },
    );

    return () => (
      <Select
        multiple={props.multiple}
        clearable={false}
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={(val) => (selected.value = val)}
        loading={loading.value}
        disabled={props.isDisabled}
        {...{ attrs }}>
        {list.value.map(({ id, name }: IOption) => (
          <Option key={id} id={id} name={name} />
        ))}
      </Select>
    );
  },
});
