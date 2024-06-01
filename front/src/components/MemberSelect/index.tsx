import { useStaffStore } from '@/store';
import { Staff, StaffType } from '@/typings';
import { Loading, TagInput } from 'bkui-vue';
import { computed, defineComponent, onMounted, PropType, ref, watch, nextTick } from 'vue';
import _ from 'lodash';

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
    allowCreate: {
      type: Boolean,
      default: false,
    },
    defaultUserlist: {
      type: Array,
      default: [],
    },
  },
  emits: ['change', 'input', 'blur'],
  setup(props, ctx) {
    const tagInputRef = ref(null);
    const staffStore = useStaffStore();
    const searchKey = ['username'];
    const userList: any = ref(props.defaultUserlist);
    const maxData = computed(() =>
      !props.multiple
        ? {
            maxData: 1,
          }
        : {},
    );
    const popoverProps = {
      boundary: document.body,
      fixOnBoundary: true,
    };

    onMounted(() => {
      if (staffStore.list.length === 0) {
        staffStore.fetchStaffs();
      }
    });

    function tpl(node: Staff) {
      return <Tpl englishName={node.username} chineseName={node.display_name} />;
    }
    function handleChange(val: Staff[]) {
      userList.value = val.map((name) => ({
        username: name,
        display_name: name,
      }));
      ctx.emit('input', val);
      ctx.emit('change', val);
    }

    function handleBlur(val: Staff[]) {
      ctx.emit('blur', val);
    }

    const getUserList = _.debounce((userName: string) => {
      if (staffStore.fetching || !userName) return;
      staffStore.fetchStaffs(userName);
    }, 1000);

    const handleInput = (userName: string) => {
      getUserList(userName);
    };

    watch(
      () => staffStore.list,
      (list) => {
        if (list.length) {
          nextTick(() => {
            const arr = [...userList.value, ...list];
            const set = new Set(arr.map(({ username }) => username));
            userList.value = Array.from(set).map((name) => ({
              username: name,
              display_name: name,
            }));
            // tagInputRef.value?.focusInputTrigger(); // 获取到数据聚焦
          });
        }
      },
      { immediate: true, deep: true },
    );

    return () => (
      <TagInput
        {...ctx.attrs}
        {...maxData.value}
        // disabled={props.disabled || staffStore.fetching}
        list={userList}
        ref={tagInputRef}
        displayKey='display_name'
        saveKey='username'
        is-async-list
        searchKey={searchKey}
        // filterCallback={handleSearch}
        modelValue={props.modelValue}
        onChange={handleChange}
        onBlur={handleBlur}
        onInput={handleInput}
        tpl={tpl}
        tagTpl={tpl}
        clearable={props.clearable}
        allowCreate={props.allowCreate}
        popoverProps={popoverProps}>
        {{
          suffix: () =>
            staffStore.fetching && <Loading class='mr8' loading={staffStore.fetching} mode='spin' size='mini' />,
        }}
      </TagInput>
    );
  },
});
