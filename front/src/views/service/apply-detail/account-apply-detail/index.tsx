import { PropType, computed, defineComponent } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useStatus } from './useStatus';
import { Close, Spinner, Success } from 'bkui-vue/lib/icon';

export interface IDetail {
  id: string;
  source: string;
  sn: string;
  type: string;
  status: string;
  applicant: string;
  // content: {
  //   bk_biz_id: number; // 业务
  //   bak_managers: string[]; // 备份负责人
  //   op_product_id: number; // 运营产品
  //   id: string; // 一级账号ID
  //   vendor: string; // 云厂商
  //   dept_id: number;
  //   managers: string[]; // 主负责人
  // };
  content: string;
  delivery_detail: {
    complete: string;
  };
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  ticket_url: string;
}

export default defineComponent({
  props: {
    detail: {
      required: true,
      type: Object as PropType<IDetail>,
    },
  },
  setup(props) {
    const businessMapStore = useBusinessMapStore();
    const info = computed(() => JSON.parse(props.detail.content));
    const statusMap = useStatus(props.detail.delivery_detail);
    return () => (
      <div class={'account-apply-detail-container'}>
        <div class={'card-wrapper'}>
          <div class={'flex-row align-item-center'}>
            <div class={'flex-row align-item-center'}>
              {statusMap[props.detail.status].tag === 'success' && <Success height={21} width={21} fill='#2DCB56' />}
              {statusMap[props.detail.status].tag === 'abort' && <Close height={21} width={21} fill='#EA3636' />}
              {statusMap[props.detail.status].tag === 'pending' && <Spinner height={21} width={21} fill='#3A84FF' />}
              <div class={'ml4'}>{statusMap[props.detail.status].label}</div>
            </div>
            <div class='approval-process-wrapper' onClick={() => window.open(props.detail.ticket_url, '_blank')}>
              审批单详情
              <i class='hcm-icon bkhcm-icon-jump-fill'></i>
            </div>
          </div>
        </div>
        <div class={'card-wrapper mt24'}>
          <p class={'title'}>申请单信息</p>
          <DetailInfo
            detail={info.value}
            fields={[
              {
                prop: 'vendor',
                name: '云厂商',
              },
              {
                prop: 'bk_biz_id',
                name: '业务',
                render: () => businessMapStore.businessMap.get(info.value.bk_biz_id) || info.value.bk_biz_id || '--',
              },
              {
                prop: 'id',
                name: '一级账号ID',
              },
              {
                prop: 'managers',
                name: '主负责人',
              },
              {
                prop: 'bak_managers',
                name: '备份负责人',
              },
              {
                prop: 'op_product_id',
                name: '运营产品',
              },
            ]}
          />
        </div>

        <div class={'card-wrapper mt24'}>
          <p class={'title'}>账号信息</p>
          <DetailInfo
            wide
            detail={info.value}
            fields={[
              {
                prop: 'id',
                name: '二级账号ID',
              },
              {
                prop: 'name',
                name: '账号名称',
              },
            ]}
          />
        </div>
      </div>
    );
  },
});
