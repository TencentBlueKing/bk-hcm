import { defineComponent, ref } from 'vue';
import { Popover, Input } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'SimpleSearchSelect',
  props: {
    searchValue: {
      type: String,
      required: true,
    },
    dataList: {
      type: Array<{ id: string; name: string }>,
    },
  },
  emits: ['update:searchValue'],
  setup(props, ctx) {
    const isShow = ref(false);
    const searchVal = ref('');
    const searchRef = ref();
    const handleSearchDataClick = (e: MouseEvent) => {
      searchVal.value = `${e.target.dataset.name}：`;
      isShow.value = false;
      searchRef.value.focus();
    };
    const handleClick = () => {
      isShow.value = true;
    };
    const handleEnter = (v: string) => {
      const [searchName, searchVal] = v.split('：');
      const target = props.dataList.find((item) => item.name === searchName);
      ctx.emit('update:searchValue', `${target.id}:${searchVal}`);
    };
    const handleClear = () => {
      ctx.emit('update:searchValue', '');
      isShow.value = true;
    };
    return () => (
      <div class='simple-search-select'>
        <Popover trigger='click' isShow={isShow.value} theme='light' disableTeleport={true} arrow={false}>
          {{
            default: () => (
              <Input
                ref={searchRef}
                type='search'
                clearable
                v-model={searchVal.value}
                onClick={handleClick}
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
