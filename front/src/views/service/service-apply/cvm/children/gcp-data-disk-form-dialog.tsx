import { computed, defineComponent, PropType, reactive, ref, watch } from 'vue';
import { Dialog, Form, Select, Input, Radio } from 'bkui-vue';
import { getGcpDataDiskDefaults } from '../../hooks/use-cvm-form-data';
import { InfoLine as InfoLineIcon } from 'bkui-vue/lib/icon';

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
    const formRef = ref(null);

    const localIsShow = computed({
      get() {
        return props.isShow;
      },
      set(val) {
        emit('update:isShow', val);
      },
    });

    const localFormData = reactive<IDiskOption>(getGcpDataDiskDefaults());
    watch(localIsShow, (isShow) => {
      if (isShow) {
        const defaultData = {
          ...getGcpDataDiskDefaults(),
          ...props.formData,
        };

        Object.keys(defaultData).forEach((key) => {
          localFormData[key] = defaultData[key];
        });
      }
    });

    const handleConfirm = async () => {
      await formRef.value.validate();
      emit(props.isEdit ? 'save' : 'add', { ...localFormData });
    };

    return () => (
      <Dialog
        title='新增磁盘'
        isShow={localIsShow.value}
        quickClose={false}
        theme='primary'
        width='860'
        height='560'
        onConfirm={handleConfirm}
        onClosed={() => emit('close')}>
        <Form ref={formRef} model={localFormData} labelWidth={90}>
          <FormItem label='磁盘来源' required={false}>
            空白磁盘
          </FormItem>
          <FormItem label='磁盘类型' property='disk_type' required>
            <Select v-model={localFormData.disk_type}>
              {props.dataDiskTypes.map(({ id, name }: IOption) => (
                <Option key={id} value={id} label={name}></Option>
              ))}
            </Select>
          </FormItem>
          <FormItem label='大小' property='disk_size_gb' min={10} max={65536}>
            <Input type='number' v-model={localFormData.disk_size_gb} suffix='GB'></Input>
          </FormItem>
          <FormItem label='挂载模式' property='mode'>
            <RadioGroup v-model={localFormData.mode}>
              <Radio label='READ_WRITE'>读写</Radio>
              <Radio label='READ_ONLY'>只读</Radio>
            </RadioGroup>
          </FormItem>
          <FormItem label='删除规则' property='auto_delete'>
            <RadioGroup v-model={localFormData.auto_delete}>
              <Radio label={false}>保留磁盘</Radio>
              <Radio label={true}>删除磁盘</Radio>
            </RadioGroup>
          </FormItem>
        </Form>
        <div style={{ display: 'flex', alignItems: 'center', padding: '0 12px' }}>
          <InfoLineIcon />
          说明：新增的数据盘，需要登录机器挂载和格式化，
          <a
            target='_blank'
            style={{ color: '#3a84ff' }}
            href='https://cloud.google.com/compute/docs/disks/add-persistent-disk?hl=zh-cn'>
            参考文档
          </a>
        </div>
      </Dialog>
    );
  },
});
