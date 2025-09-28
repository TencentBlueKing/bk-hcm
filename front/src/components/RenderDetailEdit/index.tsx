import { Input, Select } from 'bkui-vue';
import { defineComponent, ref, nextTick, watch, PropType, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import OrganizationSelect from '@/components/OrganizationSelect';
import cssModule from './index.module.scss';

export default defineComponent({
  props: {
    fromPlaceholder: {
      type: String,
      default: '请输入',
    },
    fromType: {
      type: String,
      default: 'input',
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    isLoading: {
      type: Boolean,
      default: false,
    },
    fromKey: {
      type: String,
      default: '',
    },
    modelValue: {
      type: String as PropType<any>,
      default: [],
    },
    selectData: {
      type: Array as PropType<any>,
      default: [],
    },
    needValidate: {
      type: Boolean,
      default: true,
    },
    hideEdit: {
      type: Boolean,
      default: false,
    },
    trim: {
      type: Boolean,
      default: true,
    },
  },
  emits: ['update:modelValue', 'change', 'input', 'blur'],
  setup(props, ctx) {
    const { t } = useI18n();
    const { Option } = Select;
    const renderEdit = ref(false);
    const inputRef = ref<InstanceType<typeof Input>>(null);
    const selectRef = ref(null);
    const handleEdit = () => {
      renderEdit.value = true;
      nextTick(() => {
        // @ts-ignore
        inputRef.value?.focus();
        selectRef.value?.handleTogglePopover();
      });
    };

    watch(
      () => props.isEdit,
      () => {
        renderEdit.value = props.isEdit;
      },
    );

    const handleChange = (val: any) => {
      const value = props.trim ? val.trim() : val;
      ctx.emit('change', value, props.fromKey);
      ctx.emit('input', value);
      ctx.emit('update:modelValue', value);
    };

    const handleOrganChange = (val: any) => {
      renderEdit.value = false;
      ctx.emit('change', val, 'departmentId');
    };

    const handleBlur = (key: string) => {
      if (props.needValidate) {
        ctx.emit('blur', renderEdit.value, key);
      } else {
        renderEdit.value = false;
      }
    };

    const handleKeyUpEnter = (e: KeyboardEvent) => {
      if (e.key !== 'Enter') return;
      handleBlur(props.fromKey);
    };

    const computedDefaultUserlist = computed(() => {
      let res = props.modelValue;
      if (props.fromType === 'member') {
        res = props.modelValue.map((name: string) => ({
          username: name,
          display_name: name,
        }));
      }
      return res;
    });

    const renderComponentsContent = (type: string) => {
      switch (type) {
        case 'input':
          return (
            <Input
              ref={inputRef}
              class={cssModule.w320}
              placeholder={props.fromPlaceholder}
              modelValue={props.modelValue}
              onChange={handleChange}
              onBlur={() => handleBlur(props.fromKey)}
              onKeyup={(_, e) => handleKeyUpEnter(e)}
            />
          );
        case 'member':
          return (
            <MemberSelect
              class={cssModule.w320}
              v-model={props.modelValue}
              defaultUserlist={computedDefaultUserlist.value}
              onChange={handleChange}
              onBlur={() => handleBlur(props.fromKey)}
            />
          );
        case 'department':
          return <OrganizationSelect class={cssModule.w320} v-model={props.modelValue} onChange={handleOrganChange} />;
        case 'textarea':
          return (
            <Input
              ref={inputRef}
              class={cssModule.w320}
              placeholder={props.fromPlaceholder}
              type='textarea'
              modelValue={props.modelValue}
              onChange={handleChange}
              onBlur={() => handleBlur(props.fromKey)}
            />
          );
        case 'select':
          return (
            <Select
              ref={selectRef}
              class={cssModule.w320}
              modelValue={props.modelValue}
              filterable
              multiple-mode='tag'
              placeholder={props.fromPlaceholder}
              onChange={handleChange}
              onBlur={() => handleBlur(props.fromKey)}>
              {props.selectData.map((item: any) => (
                <Option key={item.id} id={item.id} name={item.name}>
                  {item.name}
                </Option>
              ))}
            </Select>
          );
        default:
          return (
            <Input
              ref={inputRef}
              class={cssModule.w320}
              placeholder={t('请输入')}
              modelValue={props.modelValue}
              onChange={handleChange}
              onBlur={() => handleBlur(props.fromKey)}
            />
          );
      }
    };

    const renderTextContent = (type: string) => {
      switch (type) {
        case 'input':
          return props.modelValue || '--';
        case 'member':
          return props.modelValue.length ? props.modelValue.join(',') : '暂无';
        case 'select':
          // eslint-disable-next-line no-case-declarations
          let selectModelValue;
          if (Array.isArray(props.modelValue)) {
            selectModelValue = props.selectData.filter((e: any) => props.modelValue.includes(e.id));
          } else {
            selectModelValue = props.selectData.filter((e: any) => e.id === props.modelValue);
          }
          if (selectModelValue.length) {
            selectModelValue = selectModelValue.map((e: any) => e.name);
          }
          // eslint-disable-next-line no-nested-ternary
          return selectModelValue.length
            ? selectModelValue.join(',')
            : props.modelValue.join(',') === '-1'
            ? '未分配'
            : '暂无';
        default:
          return props.modelValue;
      }
    };

    ctx.expose({ renderEdit, handleEdit });

    return () => (
      <>
        {renderEdit.value ? renderComponentsContent(props.fromType) : renderTextContent(props.fromType)}
        {!(renderEdit.value || props.hideEdit) && (
          <i onClick={handleEdit} class={['icon hcm-icon bkhcm-icon-bianji', cssModule['edit-icon']]} />
        )}
      </>
    );
  },
});
