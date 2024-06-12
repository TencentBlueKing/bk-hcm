import { PropType, VNode, defineComponent } from 'vue';
import { Container, Form } from 'bkui-vue';
import { FormItemProps } from 'bkui-vue/lib/form/form-item';

export interface IFormItemProps extends FormItemProps {
  render: () => VNode;
}

export default defineComponent({
  name: 'FilterFormContainer',
  props: {
    col: Number,
    gutter: Number,
    margin: Number,
    formConfig: Array as PropType<Array<IFormItemProps>>,
    formModel: Object,
  },
  setup(props) {
    return () => (
      <Container col={props.col} gutter={props.gutter} margin={props.margin}>
        <Form formType='vertical' model={props.formModel}>
          <Container.Row>
            {props.formConfig.map(({ render, ...rest }) => (
              <Container.Col>
                <Form.FormItem key={rest.label} {...rest}>
                  {render()}
                </Form.FormItem>
              </Container.Col>
            ))}
          </Container.Row>
        </Form>
      </Container>
    );
  },
});
