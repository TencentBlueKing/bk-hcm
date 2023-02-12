import { defineComponent, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Dropdown } from 'bkui-vue';
import './index.scss';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'ApplyHeaderSelect',
  props: {
    title: {
      type: String,
    },
    filterData: {
      type: Object,
      default: () => {
        return {};
      },
    },
    selectContent: {
      type: Object,
    },
    active: {
      type: [Number, String],
    },
  },
  emits: ['on-select'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const state = reactive({
      dataDisplay: props.filterData,
      title: props.title,
      selectValue: props.active,
    });
    const isDropdownShow = ref(false);

    const handleSelect = (payload: Record<string, number | string>) => {
      state.selectValue = payload.value;
      isDropdownShow.value = false;
      emit('on-select', payload);
    };

    const handleShow = () => {
      console.log('打开');
      isDropdownShow.value = true;
    };

    const handleHide = () => {
      console.log('隐藏');
      isDropdownShow.value = false;
    };

    return () => (
      <div>
        <div class="apply-select">
          <div class="title">{state.title}</div>
          <Dropdown trigger="click" onShow={ handleShow } onHide={ handleHide }>
            {{
              default: () => (
                <span class="cursor-pointer flex-row align-items-center ">
                  {
                  state.dataDisplay.length
                  && state.dataDisplay
                    .find((item: Record<string, number | string>) => item.value === state.selectValue)?.label}
                  <i
                    class={[
                      'icon',
                      'hcm-icon',
                      isDropdownShow.value ? 'bkhcm-icon-down-shape transform180' : 'bkhcm-icon-down-shape',
                    ]}
                  />
                </span>
              ),
              content: () => (
                <DropdownMenu ext-cls={[!isDropdownShow.value ? 'hide-dropdown-menu' : '']}>
                  <div>
                    {state.dataDisplay.map((item: Record<string, number | string>) => (
                      <DropdownItem onClick={() => handleSelect(item)}>{
                        t(item.label)}
                      </DropdownItem>
                    ))}
                  </div>
                </DropdownMenu>
              ),
            }}
          </Dropdown>
        </div>
      </div>
    );
  },
});
