import { computed, defineComponent, PropType, reactive, watch } from 'vue';
import { Dialog, Form, Select, Input, Radio } from 'bkui-vue';
import { getGcpDataDiskDefaults } from '../../hooks/use-cvm-form-data';
import type { IDiskOption } from '../../hooks/use-cvm-form-data';
import type { IOption } from '@/typings/common';

const { FormItem } = Form;
const { Option } = Select;
const { Group: RadioGroup } = Radio;

export default defineComponent({
  props: {
    isShow: Boolean as PropType<boolean>,
    isEdit: Boolean as PropType<boolean>,
    formData: Object as PropType<IDiskOption>,
    dataDiskTypes: Array as PropType<IOption[]>,
  },
  emits: ['update:isShow', 'save', 'add', 'close'],
  setup(props, { emit }) {
    const localIsShow = computed({
      get() {
        return props.isShow;
      },
      set(val) {
        emit('update:isShow', val);
      },
    });


    let localFormData = reactive<IDiskOption>(getGcpDataDiskDefaults());
    watch(localIsShow, (isShow) => {
      if (isShow) {
        localFormData = {
          ...getGcpDataDiskDefaults(),
          ...props.formData,
        };
      }
    });

    return () => <Dialog
      title="新增磁盘"
      isShow={localIsShow.value}
      quickClose={false}
      theme="primary"
      width="860"
      height="560"
      onConfirm={() => emit(props.isEdit ? 'save' : 'add', localFormData)}
      onClosed={() => emit('close')}
    >
      <Form model={localFormData} labelWidth={90}>
        <FormItem label='名称' property="name">
          <Input v-model={localFormData.disk_name}></Input>
        </FormItem>
        <FormItem label='磁盘来源' required={false}>空白磁盘</FormItem>
        <FormItem label='磁盘类型' property="type" rules={[]}>
          <Select v-model={localFormData.disk_type}>{
              props.dataDiskTypes.map(({ id, name }: IOption) => (
                <Option key={id} value={id} label={name}></Option>
              ))
            }
          </Select>
        </FormItem>
        <FormItem label='大小' property="size" min={10} max={65536}>
          <Input type='number' v-model={localFormData.disk_size_gb} suffix="GB"></Input>
        </FormItem>
        <FormItem label='挂载模式' property="mode">
          <RadioGroup v-model={localFormData.mode}>
            <Radio label="READ_WRITE">读写</Radio>
            <Radio label="READ_ONLY">只读</Radio>
          </RadioGroup>
        </FormItem>
        <FormItem label='删除规则' property="delrule">
          <RadioGroup v-model={localFormData.auto_delete}>
            <Radio label={false}>保留磁盘</Radio>
            <Radio label={true}>删除磁盘</Radio>
          </RadioGroup>
        </FormItem>
      </Form>
    </Dialog>;
  },
});
