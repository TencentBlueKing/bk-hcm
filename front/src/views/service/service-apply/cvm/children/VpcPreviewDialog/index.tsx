import { Dialog, Form } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';

const { FormItem } = Form;

export interface IVpcItem {
  cloud_id: string; // 资源ID
  name: string; // 名称
  bk_cloud_id: number; // 管控区域ID
  extension: {
    cidr: Array<{
      cidr: string;   // CIDR
    }>
  }
};

export default defineComponent({
  props: {
    isShow: {
      type: Boolean,
      required: true,
    },
    data: {
      type: Object as PropType<IVpcItem>,
      required: true,
    },
    handleClose: {
      type: Function as PropType<() => void>,
      required: true,
    },
  },
  setup(props) {
    return () => (
      <Dialog
        dialogType='show'
        isShow={props.isShow}
        title={'VPC 预览'}
        onClosed={props.handleClose}
      >
        <Form>
          <FormItem
            label='资源ID: '
          >
            <span class={'vpc-dialog-highlit-font'}>
              { props.data.cloud_id }
            </span>
            <i class="vpc-dialog-highligt-icon icon bk-icon icon-arrows--right--line"
              onClick={() => {
                if (!props.data.cloud_id) return;
                const url = `/#/business/vpc?cloud_id=${props.data.cloud_id}`;
                window.open(url, '_blank');
              }}
            ></i>
          </FormItem>
          <FormItem
            label='名称: '
          >
            <span class={'vpc-dialog-highlit-font'}>
              { props.data.name }
            </span>
          </FormItem>
          <FormItem
            label='管控区域ID: '
          >
            <span class={'vpc-dialog-highlit-font'}>
              { props.data.bk_cloud_id }
            </span>
          </FormItem>
          <FormItem
            label='CIDR: '
          >
            <span class={'vpc-dialog-highlit-font'}>
              {
                props.data?.extension?.cidr?.map(obj => (
                  <p>
                    { obj.cidr }
                  </p>
                ))
              }
            </span>
          </FormItem>
        </Form>
      </Dialog>
    );
  },
});

