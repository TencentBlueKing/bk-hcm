import { PropType, defineComponent } from 'vue';
import { Button, Popover } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import './index.scss';

export interface VpcDetail {
  id: string;
  vendor: string;
  account_id: string;
  cloud_id: string;
  name: string;
  region: string;
  category: string;
  memo: string;
  bk_cloud_id: number;
  bk_biz_id: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  extension: VpcDetailExtension;
}

interface VpcDetailExtension {
  cidr: Cidr[];
  is_default: boolean;
  enable_multicast: boolean;
  dns_server_set: string[];
}

interface Cidr {
  type: string;
  cidr: string;
  category: string;
}

export default defineComponent({
  name: 'VpcReviewPopover',
  props: {
    data: {
      type: Object as PropType<VpcDetail>,
      required: true,
    },
  },
  setup(props) {
    const { t } = useI18n();

    return () => (
      <Popover theme='light' trigger='click' placement='bottom-start' extCls='vpc-review-popover'>
        {{
          default: () => (
            <Button style={{ marginLeft: '16px' }} text theme='primary' disabled={!props.data?.id}>
              {t('预览')}
            </Button>
          ),
          content: () => (
            <div class='review-detail'>
              <div class='detail-item'>
                <div class='item-label'>资源 ID</div>
                <div class='item-value'>{props.data.cloud_id}</div>
              </div>
              <div class='detail-item'>
                <div class='item-label'>名称</div>
                <div class='item-value'>{props.data.name}</div>
              </div>
              <div class='detail-item'>
                <div class='item-label'>管控区域 ID</div>
                <div class='item-value'>{props.data.bk_cloud_id}</div>
              </div>
              {/* TODO：替换为flex-tag */}
              <div class='detail-item'>
                <div class='item-label'>IPv4 CIDR</div>
                <div class='item-value'>
                  {props.data.extension?.cidr?.map((obj: any) => (
                    <p>{obj.cidr}</p>
                  ))}
                </div>
              </div>
            </div>
          ),
        }}
      </Popover>
    );
  },
});
