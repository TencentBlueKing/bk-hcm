import { defineComponent } from 'vue';
import { Button } from 'bkui-vue';
import './index.scss';
import RsConfigTable from '../../all-groups-manager/rs-config-table';

export default defineComponent({
  name: 'TargetGroupDetail',
  setup() {
    const targetGroupDetail = [
      {
        title: '基本信息',
        content: [
          {
            label: '云账号',
            value: '腾讯云222',
          },
          {
            label: '地域',
            value: '新加坡',
          },
          {
            label: '目标组名称',
            value: '目标组1123',
          },
          {
            label: '网络',
            value: '1290.34.2342',
          },
          {
            label: '协议端口',
            value: 'TCP:4600',
          },
          {
            label: '创建时间',
            value: '2023-07-03 18:00:00',
          },
        ],
      },
      {
        title: 'RS 信息',
        content: <RsConfigTable noOperation />,
      },
    ];
    return () => (
      <div class='target-group-detail-page'>
        <Button class='fixed-operate-btn' outline theme='primary'>
          编辑
        </Button>
        <div class='detail-info-container'>
          {targetGroupDetail.map(({ title, content }) => (
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
