import { computed, reactive, ref } from 'vue';
// import components
import { Input, Select } from 'bkui-vue';
// import hooks
import { useI18n } from 'vue-i18n';

const { Option } = Select;

export default () => {
  // use hooks
  const { t } = useI18n();

  const isShow = ref(false);
  const action = ref<number>(); // 0 - add, 1 - update
  const formData = reactive({
    domain: '',
    url: '',
    mode: '',
  });

  // 清空表单参数
  const clearParams = () => {
    Object.assign(formData, {
      domain: '',
      url: '',
      mode: '',
    });
  };

  /**
   * 显示 dialog
   * @param data 域名信息, 如果为 undefined, 表示新增
   */
  const handleShow = (data?: any) => {
    isShow.value = true;
    clearParams();
    if (data) {
      action.value = 1;
      Object.assign(formData, data);
    } else {
      action.value = 0;
    }
  };

  const handleSubmit = () => {};

  const formItemOptions = computed(() => [
    {
      label: t('域名'),
      property: 'domain',
      required: true,
      content: () => <Input v-model={formData.domain} />,
    },
    {
      label: t('URL 路径'),
      property: 'url',
      required: true,
      hidden: action.value === 1,
      content: () => <Input v-model={formData.url} />,
    },
    {
      label: t('模式'),
      property: 'mode',
      required: true,
      hidden: action.value === 1,
      content: () => (
        <Select v-model={formData.mode} placeholder={t('请选择模式')}>
          <Option id='1' name='1' />
          <Option id='2' name='2' />
        </Select>
      ),
    },
  ]);

  return {
    isShow,
    action,
    formItemOptions,
    handleShow,
    handleSubmit,
  };
};
