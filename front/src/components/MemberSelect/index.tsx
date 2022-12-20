import { useStaffStore } from '@/store';
import { Staff, StaffType } from '@/typings';
import { Loading, TagInput } from 'bkui-vue';
import { computed, defineComponent, onMounted, PropType, ref, watch, nextTick } from 'vue';

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
  emits: ['change', 'input', 'blur'],
  setup(props, ctx) {
    const tagInputRef = ref(null);
    const staffStore = useStaffStore();
    const searchKey = ['username'];
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
          englishName={node.username}
          chineseName={node.display_name}
        />
      );
    }
    function handleChange(val: Staff[]) {
      ctx.emit('input', val);
      ctx.emit('change', val);
    }

    function handleBlur(val: Staff[]) {
      ctx.emit('blur', val);
    }

    function handleSearch(lowerCaseValue: string, _: string | string[], list: Staff[]) {
      return list.filter((item) => {
        const username = item.username.toLowerCase();
        return username.includes(lowerCaseValue) || item.display_name.includes(lowerCaseValue);
      });
    }

    watch(
      () => staffStore.list,
      (list) => {
        if (list.length) {
          nextTick(() => {
            tagInputRef.value?.focusInputTrigger(); // 获取到数据聚焦
          });
        }
      },
      { immediate: true },
    );

    return () => (
      <TagInput
        {...ctx.attrs}
        {...maxData.value}
        // disabled={props.disabled || staffStore.fetching}
        list={staffStore.list}
        ref={tagInputRef}
        displayKey="display_name"
        saveKey="username"
        searchKey={searchKey}
        filterCallback={handleSearch}
        modelValue={props.modelValue}
        onChange={handleChange}
        onBlur={handleBlur}
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
