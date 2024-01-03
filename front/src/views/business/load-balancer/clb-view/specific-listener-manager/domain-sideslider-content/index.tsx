import { computed, defineComponent, reactive } from 'vue';
import { Input, Select, Form } from 'bkui-vue';
import './index.scss';

const { Option } = Select;
const { FormItem } = Form;

export default defineComponent({
  name: 'DomainSidesliderContent',
  setup() {
    const formData = reactive({
      domain: '',
      url: '',
      mode: '',
    });
    const formItemOptions = computed(() => [
      {
        label: '域名',
        property: 'domain',
        required: true,
        content: () => <Input v-model={formData.domain} />,
      },
      {
        label: 'URL 路径',
        property: 'url',
        required: true,
        content: () => <Input v-model={formData.url} />,
      },
      {
        label: '模式',
        property: 'mode',
        required: true,
        content: () => (
          <Select v-model={formData.mode} placeholder='请选择模式'>
            <Option id='1' name='1' />
            <Option id='2' name='2' />
          </Select>
        ),
      },
    ]);

    return () => (
      <>
        <p class='readonly-info'>
          <span class='label'>监听器名称</span>:<span class='value'>web站点</span>
        </p>
        <p class='readonly-info'>
          <span class='label'>协议端口</span>:<span class='value'>HTTP:50</span>
        </p>
        <Form formType='vertical'>
          {formItemOptions.value.map(({ label, required, property, content }) => {
            return (
              <FormItem label={label} required={required} key={property}>
                {content()}
              </FormItem>
            );
          })}
        </Form>
      </>
    );
  },
});
