import { defineComponent, ref, onMounted, PropType, watch, computed } from 'vue';
import apiService from '../../../../api/scrApi';
import isEqual from 'lodash/isEqual';
export default defineComponent({
  name: 'AreaSelector',
  props: {
    value: {
      type: [String, Array] as PropType<string | string[]>,
      default: '',
    },
    valueKey: {
      type: String,
      default: '',
    },
    params: {
      type: Object,
      default: () => ({}),
    },
    multiple: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['change'],
  setup(props, { attrs, emit }) {
    const loading = ref(false);
    const selectedValue = ref();
    const options = ref([]);
    const isIdc = computed(() => ['IDCDVM', 'IDCPM'].includes(props.params.resourceType));
    const isQcloud = computed(() => ['QCLOUDCVM', 'QCLOUDDVM'].includes(props.params.resourceType));
    const handleSelectorChange = (value: string | string[]) => {
      emit('change', value);
    };
    const initValue = () => {
      selectedValue.value = props.value;
    };
    const fetchOptions = () => {
      if (isQcloud.value) {
        loadQcloudRegions();
      }

      if (isIdc.value) {
        loadIdcRegions();
      }

      initValue();
    };
    const loadQcloudRegions = async () => {
      const { info } = await apiService.getRegions('qcloud');
      options.value = info;
    };
    const loadIdcRegions = async () => {
      const { info } = await apiService.getRegions('idc');
      options.value = info.map((item: any) => {
        return {
          region: item,
          region_cn: item,
        };
      });
    };

    watch(
      () => props.value,
      (val) => {
        if (!isEqual(val, selectedValue.value) && options.value.length !== 0) {
          initValue();
        }
      },
      { immediate: true },
    );
    watch(
      () => props.params.resourceType,
      (newVal, oldVal) => {
        if (!isEqual(newVal, oldVal)) {
          fetchOptions();
        }
      },
      { immediate: true },
    );
    onMounted(() => {});

    return () => (
      <bk-select
        v-bind={attrs}
        filterable
        default-first-option
        multiple={props.multiple}
        v-model={props.value}
        loading={loading.value}
        onChange={handleSelectorChange}>
        {options.value.map((item: any) => (
          <bk-option key={item.region} value={item.region} label={item.region_cn}></bk-option>
        ))}
      </bk-select>
    );
  },
});
