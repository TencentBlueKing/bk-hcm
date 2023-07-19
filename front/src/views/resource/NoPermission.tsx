import { PropType, defineComponent } from 'vue';

export default defineComponent({
  props: {
    message: {
      type: String as PropType<string>,
    },
  },
  setup: (props) => {
    return () => (
      <bk-exception
        class='exception-wrap-item'
        type='403'
        title='无该业务权限'
        description={`请联系 ${props.message} 开通`}>
      </bk-exception>
    );
  },
});
