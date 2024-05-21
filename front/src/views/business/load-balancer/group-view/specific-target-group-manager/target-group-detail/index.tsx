import { PropType, computed, defineComponent } from 'vue';
// import components
import { Button, Link } from 'bkui-vue';
import { Share } from 'bkui-vue/lib/icon';
import RsConfigTable from '../../components/RsConfigTable';
import AddOrUpdateTGSideslider from '../../components/AddOrUpdateTGSideslider';
import AddRsDialog from '../../components/AddRsDialog';
// import stores
import { useRegionsStore } from '@/store/useRegionsStore';
// import utils
import bus from '@/common/bus';
import { timeFormatter } from '@/common/util';
// import constants
import { VendorEnum } from '@/common/constant';
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
  },
  setup(props) {
    // use stores
    const { getRegionName } = useRegionsStore();

    const targetGroupDetail = computed(() => [
      {
        title: '基本信息',
        content: [
          {
            label: '云账号',
            value: props.detail.account_id,
          },
          {
            label: '地域',
            value: getRegionName(VendorEnum.TCLOUD, props.detail.region),
          },
          {
            label: '目标组名称',
            value: props.detail.name,
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
          },
          {
            label: '协议端口',
            value: `${props.detail.protocol}:${props.detail.port}`,
          },
          {
            label: '创建时间',
            value: timeFormatter(props.detail.created_at),
          },
        ],
      },
      {
        title: 'RS 信息',
        content: <RsConfigTable onlyShow rsList={props.detail.target_list} />,
      },
    ]);

    // click-handler - 编辑目标组
    const handleEditTargetGroup = () => {
      bus.$emit('editTargetGroup', { ...props.detail, rs_list: props.detail.target_list });
    };

    return () => (
      <div class='target-group-detail-page'>
        <Button class='fixed-operate-btn' outline theme='primary' onClick={handleEditTargetGroup}>
          编辑
        </Button>
        <div class='detail-info-container'>
          {targetGroupDetail.value.map(({ title, content }) => (
            <div class='detail-info-wrap'>
              <h3 class='info-title'>{title}</h3>
              <div class='info-content'>
                {Array.isArray(content)
                  ? content.map(({ label, value }) => (
                      <div class='info-item'>
                        <span class='info-item-label'>{label}</span>:<span class='info-item-value'>{value}</span>
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
