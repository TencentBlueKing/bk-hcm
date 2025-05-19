import { Dialog, Form } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';

const { FormItem } = Form;

export interface ISubnetItem {
  cloud_id: string; // 资源ID
  name: string; // 名称
  region: string; // 可用区
  vpc_id: string; // 所属VPC
  ipv4_cidr: Array<string>; // IPv4 CIDR
  ipv6_cidr?: Array<string>; // IPv6 CIDR
  used_ip_count: number; // IP数---已使用
  total_ip_count: number; // IP数---总共
  available_ip_count: number; // IP数---剩余
}

export default defineComponent({
  props: {
    isShow: {
      type: Boolean,
      required: true,
    },
    data: {
      type: Object as PropType<ISubnetItem>,
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
        class='preview-dialog'
        dialogType='show'
        isShow={props.isShow}
        title={'子网预览'}
        onClosed={props.handleClose}>
        <Form labelWidth={108} label-position='right'>
          <FormItem label='资源ID：'>
            <span class={'subnet-dialog-highlight-font'}>{props.data.cloud_id}</span>
          </FormItem>
          <FormItem label='名称：'>
            <span class={'subnet-dialog-highlight-font'}>{props.data.name}</span>
          </FormItem>
          <FormItem label='可用区：'>
            <span class={'subnet-dialog-highlight-font'}>{props.data.region}</span>
          </FormItem>
          <FormItem label='所属VPC：'>
            <span class={'subnet-dialog-highlight-font'}>{props.data.vpc_id}</span>
          </FormItem>
          {/* TODO：替换为flex-tag */}
          <FormItem label='IPv4 CIDR：'>
            <span class={'subnet-dialog-highlight-font'}>
              {props.data?.ipv4_cidr?.map((str) => (
                <>
                  <span>{str}</span>
                  <br />
                </>
              ))}
            </span>
          </FormItem>
          {/* TODO：替换为flex-tag */}
          <FormItem label='IPv6 CIDR：'>
            <span class={'subnet-dialog-highlight-font'}>
              {props.data?.ipv6_cidr?.map((str) => (
                <>
                  <span>{str}</span>
                  <br />
                </>
              ))}
            </span>
          </FormItem>
          <FormItem label='IP数：'>
            <span class={'subnet-dialog-highlight-font'}>
              {`总${props.data.total_ip_count}个，已使用${props.data.used_ip_count}个，剩余${props.data.available_ip_count}个`}
            </span>
          </FormItem>
        </Form>
      </Dialog>
    );
  },
});
