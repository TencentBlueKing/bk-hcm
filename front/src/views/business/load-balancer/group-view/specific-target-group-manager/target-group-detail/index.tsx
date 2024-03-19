import { computed, defineComponent } from 'vue';
import { Button } from 'bkui-vue';
import './index.scss';
import RsConfigTable from '../../all-groups-manager/rs-config-table';

export default defineComponent({
  name: 'TargetGroupDetail',
  props: {
    detail: {
      required: true,
      type: Object,
    },
  },
  setup(props) {
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
            value: props.detail.region,
          },
          {
            label: '目标组名称',
            value: props.detail.name,
          },
          {
            label: '所属vpc',
            value: props.detail.cloud_vpc_id,
          },
          {
            label: '协议端口',
            value: `${props.detail.protocol}:${props.detail.port}`,
          },
          {
            label: '创建时间',
            value: props.detail.created_at,
          },
        ],
      },
      {
        title: 'RS 信息',
        content: <RsConfigTable noOperation details={props.detail.target_list} />,
      },
    ]);
    return () => (
      <div class='target-group-detail-page'>
        <Button class='fixed-operate-btn' outline theme='primary'>
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
      </div>
    );
  },
});
