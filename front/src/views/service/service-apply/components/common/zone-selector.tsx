import http from '@/http';
import { computed, defineComponent, PropType, ref, watchEffect } from 'vue';
import { Tag, Loading } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';

import './zone-selector.scss';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    modelValue: Array as PropType<string[]>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    multiple: Boolean as PropType<boolean>,
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, expose }) {
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

    const isEmptyCond = computed<boolean>(() => !props.vendor?.length || !props.region?.length);

    watchEffect(async () => {
      const filter: QueryFilterType = {
        op: 'and',
        rules: [],
      };

      switch (props.vendor) {
        case VendorEnum.TCLOUD:
          filter.rules = [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: props.vendor,
            },
            {
              field: 'state',
              op: QueryRuleOPEnum.EQ,
              value: 'AVAILABLE',
            },
          ];
          break;
        case VendorEnum.AWS:
          filter.rules = [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: props.vendor,
            },
            {
              field: 'state',
              op: QueryRuleOPEnum.EQ,
              value: 'available',
            },
          ];
          break;
        case VendorEnum.GCP:
          filter.rules = [
            {
              field: 'state',
              op: QueryRuleOPEnum.EQ,
              value: 'UP',
            },
          ];
          break;
      }

      if (!isEmptyCond.value) {
        loading.value = true;
        const result = await http.post(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${props.vendor}/regions/${props.region}/zones/list`,
          {
            filter,
            page: {
              count: false,
              start: 0,
              limit: 500,
            },
          },
        );
        list.value = result?.data?.details ?? [];
        loading.value = false;
      }
    });

    const handleChange = (checked: boolean, name: string) => {
      if (props.multiple) {
        if (checked) {
          selected.value.push(name);
        } else {
          const index = selected.value.findIndex((itemName) => itemName === name);
          selected.value.splice(index, 1);
        }
      } else {
        selected.value = checked ? [name] : [];
      }

      emit('change', name, selected.value);
    };

    expose({
      list,
    });

    return () => (
      <>
        {!loading.value &&
          (!isEmptyCond.value ? (
            [
              list.value.map(({ name, display_name, name_cn }) => (
                <Tag
                  class='tag-checkable'
                  key={name}
                  type='stroke'
                  checkable
                  checked={selected.value.includes(name)}
                  onChange={(checked) => handleChange(checked, name)}>
                  {display_name || name_cn || name}
                </Tag>
              )),
              !list.value.length && <span>该云地域无可用区可选择</span>,
            ]
          ) : (
            <span style={{ color: '#63656e' }}>请先选择云厂商及云地域</span>
          ))}
        {loading.value && <Loading mode='spin' size='small' />}
      </>
    );
  },
});
