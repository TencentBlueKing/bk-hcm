import { ComputedRef, PropType, computed, defineComponent, inject } from 'vue';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessStore } from '@/store';
import { useCalcTopWithNotice } from '@/views/home/hooks/useCalcTopWithNotice';
import { QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { IAuthSign } from '@/common/auth-service';
import { timeFormatter } from '@/common/util';
import bus from '@/common/bus';

import { Button, Link } from 'bkui-vue';
import { Share } from 'bkui-vue/lib/icon';
import RsConfigTable from '../../components/RsConfigTable';
import AddOrUpdateTGSideslider from '../../components/AddOrUpdateTGSideslider';
import AddRsDialog from '../../components/AddRsDialog';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupDetail',
  props: {
    detail: {
      required: true,
      type: Object,
    },
    getTargetGroupDetail: {
      type: Function as PropType<(...args: any) => any>,
    },
    id: String,
  },
  setup(props) {
    const clbOperationAuthSign = inject<ComputedRef<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');

    // use stores
    const { getRegionName } = useRegionsStore();
    const businessStore = useBusinessStore();

    const targetGroupDetail = computed(() => [
      {
        title: '基本信息',
        content: [
          {
            label: '云账号',
            value: props.detail.account_id,
            copy: true,
          },
          {
            label: '地域',
            value: getRegionName(VendorEnum.TCLOUD, props.detail.region),
            copy: true,
          },
          {
            label: '目标组名称',
            value: props.detail.name,
            copy: true,
          },
          {
            label: '所属vpc',
            value: (
              <Link
                theme='primary'
                href={`/#/resource/detail/vpc?type=tcloud&id=${props.detail.vpc_id}`}
                target='_blank'>
                <div class='flex-row align-items-center'>
                  {props.detail.cloud_vpc_id}
                  <Share class='ml5' />
                </div>
              </Link>
            ),
            copy: true,
            copyContent: props.detail.cloud_vpc_id,
          },
          {
            label: '协议端口',
            value: `${props.detail.protocol}:${props.detail.port}`,
            copy: true,
          },
          {
            label: '创建时间',
            value: timeFormatter(props.detail.created_at),
            copy: true,
          },
        ],
      },
      {
        title: 'RS 信息',
        content: (
          <RsConfigTable
            onlyShow
            rsList={props.detail.target_list}
            getTargetGroupDetail={props.getTargetGroupDetail}
            id={props.id}
          />
        ),
      },
    ]);

    // click-handler - 编辑目标组
    const handleEditTargetGroup = async () => {
      // 根据目标组id获取目标组关联lb_id
      const res = await businessStore.list(
        {
          filter: { op: QueryRuleOPEnum.AND, rules: [{ field: 'id', op: QueryRuleOPEnum.EQ, value: props.detail.id }] },
          page: { count: false, start: 0, limit: 1 },
        },
        'target_groups',
      );
      bus.$emit('editTargetGroup', {
        ...props.detail,
        rs_list: props.detail.target_list,
        lb_id: res.data.details[0].lb_id,
      });
    };

    const [calcTop] = useCalcTopWithNotice(192);

    return () => (
      <div class='target-group-detail-page'>
        <hcm-auth class='fixed-operate-btn' sign={clbOperationAuthSign.value}>
          {{
            default: ({ noPerm }: { noPerm: boolean }) => (
              <Button
                style={{ top: calcTop.value }}
                theme='primary'
                outline
                disabled={noPerm}
                onClick={handleEditTargetGroup}>
                编辑
              </Button>
            ),
          }}
        </hcm-auth>
        <div class='detail-info-container'>
          {targetGroupDetail.value.map(({ title, content }) => (
            <div class='detail-info-wrap'>
              <h3 class='info-title'>{title}</h3>
              <div class='info-content'>
                {Array.isArray(content)
                  ? content.map(({ label, value, copyContent, copy }) => (
                      <div class='info-item'>
                        <span class='info-item-label'>{label}</span>:<span class='info-item-value'>{value}</span>
                        {copy && <CopyToClipboard class='copy-btn' content={copyContent ?? String(value)} />}
                      </div>
                    ))
                  : content}
              </div>
            </div>
          ))}
        </div>
        <AddOrUpdateTGSideslider origin='info' getTargetGroupDetail={props.getTargetGroupDetail} />
        <AddRsDialog />
      </div>
    );
  },
});
