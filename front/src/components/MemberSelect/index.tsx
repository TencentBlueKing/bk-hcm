import { useStaffStore } from '@/stores';
import { Staff, StaffType } from '@/typings';
import { Loading, TagInput } from 'bkui-vue';
import { computed, defineComponent, onMounted, PropType } from 'vue';

import './member-select.scss';
import Tpl from './Tpl';

export default defineComponent({
  props: {
    disabled: {
      type: Boolean,
    },
    modelValue: {
      type: Array as PropType<string[]>,
    },
    type: {
      type: String as PropType<StaffType>,
      default: StaffType.RTX,
    },
    multiple: {
      type: Boolean,
      default: true,
    },
    clearable: {
      type: Boolean,
      default: true,
    },
  },
  emits: ['change', 'input'],
  setup(props, ctx) {
    const staffStore = useStaffStore();
    const searchKey = ['english_name', 'chinese_name'];
    const maxData = computed(() => (!props.multiple ? {
      maxData: 1,
    } : {}));
    const popoverProps = {
      boundary: document.body,
      fixOnBoundary: true,
    };

    onMounted(() => {
      if (staffStore.list.length === 0) {
        staffStore.fetchStaffs(props.type);
      }
    });

    function tpl(node: Staff) {
      return (
        <Tpl
          englishName={node.english_name}
          chineseName={node.chinese_name}
        />
      );
    }
    function handleChange(val: Staff[]) {
      ctx.emit('input', val);
      ctx.emit('change', val);
    }
    function handleSearch(lowerCaseValue: string, _: string | string[], list: Staff[]) {
      return list.filter((item) => {
        const english_name = item.english_name.toLowerCase();
        return english_name.includes(lowerCaseValue) || item.chinese_name.includes(lowerCaseValue);
      });
    }
    return () => (
      <TagInput
        {...ctx.attrs}
        {...maxData.value}
        // disabled={props.disabled || staffStore.fetching}
        list={staffStore.list}
        displayKey="chinese_name"
        saveKey="english_name"
        searchKey={searchKey}
        filterCallback={handleSearch}
        modelValue={props.modelValue}
        onChange={handleChange}
        tpl={tpl}
        tagTpl={tpl}
        clearable={props.clearable}
        popoverProps={popoverProps}
      >
          {{
            suffix: () => staffStore.fetching && (
              <Loading
                class="mr8"
                loading={staffStore.fetching}
                mode="spin"
                size="mini"

              />
            ),
          }}
      </TagInput>
    );
  },
});
