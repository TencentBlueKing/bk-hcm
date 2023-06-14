import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import { QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

interface IMachineType {
  instance_type: string;
  architecture?: string;
}

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    machineType: Object as PropType<IMachineType>,
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

    watch([
      () => props.vendor,
      () => props.region,
      () => props.machineType,
    ], async ([vendor, region, machineType]) => {
      if (!vendor || !region || (vendor === VendorEnum.AZURE && !machineType?.architecture)) {
        list.value = [];
        return;
      }

      loading.value = true;

      const filter = {
        op: 'and',
        rules: [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: vendor,
          },
          {
            field: 'type',
            op: QueryRuleOPEnum.EQ,
            value: 'public',
          },
        ],
      };

      switch (vendor) {
        case VendorEnum.AWS:
          filter.rules.push({
            field: 'extension.region',
            op: QueryRuleOPEnum.JSON_EQ,
            value: region,
          }, {
            field: 'state',
            op: QueryRuleOPEnum.EQ,
            value: 'available',
          });
          break;
        case VendorEnum.HUAWEI:
          filter.rules.push({
            field: 'extension.region',
            op: QueryRuleOPEnum.JSON_EQ,
            value: region,
          });
          break;
        case VendorEnum.TCLOUD:
          filter.rules.push({
            field: 'state',
            op: QueryRuleOPEnum.EQ,
            value: 'NORMAL',
          });
          break;
        case VendorEnum.AZURE:
          filter.rules.push({
            field: 'architecture',
            op: QueryRuleOPEnum.EQ,
            value: machineType.architecture,
          });
          break;
        case VendorEnum.GCP:
          filter.rules.push({
            field: 'state',
            op: QueryRuleOPEnum.EQ,
            value: 'READY',
          });
          break;
      }

      const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/images/list`, {
        filter,
        page: {
          count: false,
          start: 0,
          limit: 500,
        },
      });
      const details = result?.data?.details ?? [];
      list.value = details
        .map((item: any) => ({
          id: item.cloud_id,
          name: vendor === VendorEnum.AZURE ? `${item.platform} ${item.architecture} ${item.name}` : item.name,
        }));
      loading.value = false;
    });

    return () => (
      <Select
        filterable={true}
        modelValue={selected.value}
        onUpdate:modelValue={val => selected.value = val}
        loading={loading.value}
        {...{ attrs }}
      >
        {
          list.value.map(({ id, name }) => (
            <Option key={id} value={id} label={name}></Option>
          ))
        }
      </Select>
    );
  },
});
