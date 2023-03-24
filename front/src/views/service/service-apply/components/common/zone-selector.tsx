import http from '@/http';
import { computed, defineComponent, PropType, ref, watchEffect } from 'vue';
import { Tag, Loading } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';

import './zone-selector.scss';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    modelValue: Array as PropType<string[]>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    multiple: Boolean as PropType<boolean>,
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
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

    const isEmptyCond = computed(() => !props.vendor.length || !props.region.length);

    watchEffect(async () => {
      if (props.vendor === VendorEnum.AZURE) {
        list.value = [
          { name: 'zone1', display_name: 'Zone1' },
          { name: 'zone2', display_name: 'Zone2' },
          { name: 'zone3', display_name: 'Zone3' },
        ];
        return;
      }

      if (!isEmptyCond.value) {
        loading.value = true;
        const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${props.vendor}/regions/${props.region}/zones/list`, {
          page: {
            count: false,
            start: 0,
            limit: 500,
          },
        });
        list.value = result?.data?.details ?? [];
        loading.value = false;
      }
    });

    const handleChange = (checked: boolean, name: string) => {
      if (props.multiple) {
        if (checked) {
          selected.value.push(name);
        } else {
          const index = selected.value.findIndex(itemName => itemName === name);
          selected.value.splice(index, 1);
        }
      } else {
        selected.value = checked ? [name] : [];
      }

      emit('change', name, selected.value);
    };

    return () => <>
      {
        !loading.value
        && (!isEmptyCond.value
          ? [
            list.value.map(({ name, display_name }) => (
              <Tag
                class="tag-checkable"
                key={name}
                type="stroke"
                checkable
                checked={selected.value.includes(name)}
                onChange={checked => handleChange(checked, name)}
              >
                {display_name ?? name}
              </Tag>
            )),
            !list.value.length && <span>暂无可用区</span>,
          ]
          : <span style={{ color: '#63656e' }}>请先选择云厂商及云地域</span>
        )
      }
      { loading.value && <Loading mode='spin' size='small' /> }
    </>;
  },
});
