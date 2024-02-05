import { Dialog, Form } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const { FormItem } = Form;

export interface IVpcItem {
  cloud_id: string; // 资源ID
  name: string; // 名称
  bk_cloud_id: number; // 管控区域ID
  extension: {
    cidr: Array<{
      cidr: string; // CIDR
    }>;
  };
}

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
    const { whereAmI } = useWhereAmI();
    return () => (
      <Dialog
        class='preview-dialog'
        dialogType='show'
        isShow={props.isShow}
        title={'VPC 预览'}
        onClosed={props.handleClose}>
        <Form labelWidth={108} label-position='right'>
          <FormItem label='资源ID：'>
            <span class={'vpc-dialog-highlit-font'}>{props.data.cloud_id}</span>
            <svg
              onClick={() => {
                if (!props.data.cloud_id) return;
                const url =
                  whereAmI.value === Senarios.resource
                    ? `/#/resource/resource?cloud_id=${props.data.cloud_id}&type=vpc`
                    : `/#/business/vpc?cloud_id=${props.data.cloud_id}`;
                window.open(url, '_blank');
              }}
              class='vpc-dialog-highligt-icon'
              viewBox='0 0 1024 1024'
              version='1.1'
              xmlns='http://www.w3.org/2000/svg'
              style='fill: #3a84ff;'>
              <path d='M864 128h-70.4c0 0 0 0 0 0L640 128c-14.4 0-20.8 17.6-11.2 27.2l88 84.8L508.8 446.4l67.2 67.2 208-208 84.8 81.6c9.6 9.6 27.2 3.2 27.2-11.2V272v-41.6V160C896 142.4 881.6 128 864 128z'></path>
              <path d='M800 512v288H224V224h288V128H160c-17.6 0-32 14.4-32 32v704c0 17.6 14.4 32 32 32h704c17.6 0 32-14.4 32-32V512H800z'></path>
            </svg>
          </FormItem>
          <FormItem label='名称：'>
            <span class={'vpc-dialog-highlit-font'}>{props.data.name}</span>
          </FormItem>
          <FormItem label='管控区域ID：'>
            <span class={'vpc-dialog-highlit-font'}>{props.data.bk_cloud_id}</span>
          </FormItem>
          <FormItem label='CIDR：'>
            <span class={'vpc-dialog-highlit-font'}>
              {props.data?.extension?.cidr?.map((obj) => (
                <p>{obj.cidr}</p>
              ))}
            </span>
          </FormItem>
        </Form>
      </Dialog>
    );
  },
});
