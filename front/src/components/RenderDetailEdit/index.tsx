import { Input, Select } from 'bkui-vue';
import { defineComponent, ref, nextTick, watch, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import OrganizationSelect from '@/components/OrganizationSelect';

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
  },
  emits: ['update:modelValue', 'change', 'input', 'blur'],
  setup(props, ctx) {
    const { t } = useI18n();
    const { Option } = Select;
    const renderEdit = ref(false);
    const inputRef = ref<InstanceType<typeof Input>>(null);
    const selectRef = ref(null);
    const handleEdit = () => {
      // @ts-ignore
      renderEdit.value = true;
      console.log('props.modelValue', props.modelValue);
      nextTick(() => {
        // @ts-ignore
        inputRef.value?.focus();
        selectRef.value?.handleTogglePopover();
      });
    };


    watch(
      () => props.isEdit,
      () => {
        console.log('props.isEdit', props.isEdit);
        renderEdit.value = props.isEdit;
      },
    );

    const handleChange = (val: any) => {
      ctx.emit('change', val);
      ctx.emit('input', val);
      ctx.emit('update:modelValue', val);
    };

    const handleOrganChange = (val: any) => {
      renderEdit.value = false;
      ctx.emit('change', val, 'departmentId');
    };

    const handleBlur = (key: string) => {
      // @ts-ignore
      ctx.emit('blur', renderEdit.value, key);
    };

    const renderComponentsContent = (type: string) => {
      switch (type) {
        case 'input':
          return <Input ref={inputRef} class="w320" placeholder={props.fromPlaceholder} modelValue={props.modelValue}
          onChange={handleChange} onBlur={() => handleBlur(props.fromKey)} />;
        case 'member':
          return <MemberSelect class="w320" v-model={props.modelValue}
          onChange={handleChange} onBlur={() => handleBlur(props.fromKey)}/>;
        case 'department':
          return <OrganizationSelect class="w320" v-model={props.modelValue}
          onChange={handleOrganChange}/>;
        case 'textarea':
          return <Input ref={inputRef} class="w320" placeholder={props.fromPlaceholder} type="textarea" modelValue={props.modelValue}
          onChange={handleChange} onBlur={() => handleBlur(props.fromKey)} />;
        case 'select':
          return <Select ref={selectRef} class="w320" modelValue={props.modelValue}
          filterable multiple show-select-all multiple-mode="tag"
          placeholder={props.fromPlaceholder}
          onChange={handleChange} onBlur={() => handleBlur(props.fromKey)}>
            {props.selectData.map((item: any) => (
              <Option
                key={item.label}
                value={item.value}
                label={item.label}
              >
                {item.label}
              </Option>
            ))
          }
          </Select>;
        default:
          return <Input ref={inputRef} class="w320" placeholder={t('请输入')} modelValue={props.modelValue}
          onChange={handleChange} onBlur={() => handleBlur(props.fromKey)} />;
      }
    };

    const renderTextContent = (type: string) => {
      switch (type) {
        case 'input':
          return <span>{props.modelValue}</span>;
        case 'member':
          return props.modelValue.length
            ? <span>{props.modelValue.join(',')}</span>
            : '暂无';
        case 'select':
          // eslint-disable-next-line no-case-declarations
          let selectModelValue;
          if (Array.isArray(props.modelValue)) {
            selectModelValue = props.selectData.filter((e: any) => props.modelValue.includes(e.value));
          } else {
            selectModelValue = props.selectData.filter((e: any) => e.value === props.modelValue);
          }
          if (selectModelValue.length) {
            selectModelValue = selectModelValue.map((e: any) => e.label);
          }
          return selectModelValue.length
            ? <span>{selectModelValue.join(',')}</span>
            : '暂无';
        default:
          return <span>{props.modelValue}</span>;
      }
    };
    return () => (
        <div class="flex-row align-items-center">
            {renderEdit.value ? (
              renderComponentsContent(props.fromType)
            ) : renderTextContent(props.fromType)}
            {renderEdit.value ? '' : <i onClick={handleEdit} class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>}
        </div>
    );
  },


});
