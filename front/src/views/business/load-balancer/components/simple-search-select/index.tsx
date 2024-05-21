import { PropType, defineComponent, ref, watch } from 'vue';
import { Popover, Input } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'SimpleSearchSelect',
  props: {
    modelValue: String,
    dataList: Array<{ id: string; name: string }>,
    clearHandler: Function as PropType<(...args: any) => any>,
  },
  emits: ['update:modelValue'],
  setup(props, ctx) {
    const searchVal = ref('');
    const searchRef = ref();
    const popoverRef = ref();

    const handleSearchDataClick = (e: MouseEvent) => {
      searchVal.value = `${e.target.dataset.name}：`;
      popoverRef.value.hide();
      searchRef.value.focus();
    };

    const handleEnter = (v: string) => {
      const [searchName, searchVal] = v.split('：');
      const target = props.dataList.find((item) => item.name === searchName);
      ctx.emit('update:modelValue', `${target.id}：${searchVal}`);
    };

    const handleClear = () => {
      ctx.emit('update:modelValue', '');
      popoverRef.value.hide();
      typeof props.clearHandler === 'function' && props.clearHandler();
    };

    watch(
      () => props.modelValue,
      (val) => {
        if (val) {
          const [searchK, searchV] = val.split('：');
          const searchName = props.dataList.find((item) => item.id === searchK).name;
          searchVal.value = `${searchName}：${searchV}`;
        } else {
          searchVal.value = '';
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='simple-search-select'>
        <Popover trigger='click' theme='light' disableTeleport={true} arrow={false} ref={popoverRef}>
          {{
            default: () => (
              <Input
                ref={searchRef}
                type='search'
                clearable
                v-model={searchVal.value}
                onEnter={handleEnter}
                onClear={handleClear}
              />
            ),
            content: () => (
              <div class='search-data-list' onClick={handleSearchDataClick}>
                {props.dataList.map((item) => (
                  <div class='search-data-item' key={item.id} data-name={item.name}>
                    {item.name}
                  </div>
                ))}
              </div>
            ),
          }}
        </Popover>
      </div>
    );
  },
});
