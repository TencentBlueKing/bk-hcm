import { defineComponent } from 'vue';
import { Button, Tag } from 'bkui-vue';
import './index.scss';
import StatusLoading from '@/assets/image/status_loading.png';

export default defineComponent({
  name: 'ListenerDetail',
  props: {
    protocolType: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const listenerDetailInfo = [
      {
        title: '基本信息',
        content: [
          {
            label: '监听器名称',
            value: '主站web服务',
          },
          {
            label: '监听器ID',
            value: 'lbl-ht46i65c',
          },
          {
            label: '协议端口',
            value: 'HTTP:80',
          },
          {
            label: '域名数量',
            value: '59',
            sub_hidden: ['TCP', 'UDP'].includes(props.protocolType),
          },
          {
            label: 'URL 数量',
            value: '5234',
            sub_hidden: ['TCP', 'UDP'].includes(props.protocolType),
          },
          {
            label: '均衡方式',
            value: '加权轮询',
          },
          {
            label: '目标组',
            value: () => (
              <div class='target-group-wrap'>
                <span class='link-text-btn'>目标组134</span>
                {true && <img class='loading-icon spin-icon' src={StatusLoading} alt='' />}
              </div>
            ),
            sub_hidden: ['HTTP', 'HTTPS'].includes(props.protocolType),
          },
          {
            label: '创建时间',
            value: '2023-07-03  18:00:00',
          },
        ],
      },
      {
        title: '证书信息',
        hidden: ['HTTP', 'TCP', 'UDP'].includes(props.protocolType),
        content: [
          {
            label: '认证方式',
            value: '单向认证',
          },
          {
            label: '服务器证书',
            value: '***证书',
          },
          {
            label: 'CA证书',
            value: '***证书',
          },
        ],
      },
      {
        title: '会话保持',
        open_state: true,
        content: [
          {
            label: '会话时间',
            value: '基于源 IP 30 秒',
          },
        ],
      },
      {
        title: '健康检查',
        // 模拟一下场景
        open_state: ['HTTP', 'TCP', 'UDP'].includes(props.protocolType),
        content: [
          {
            label: '健康探测源IP',
            value: '100.64.0.0/10网段',
          },
          {
            label: '检查方式',
            value: 'TCP',
          },
          {
            label: '检查端口',
            value: '4600',
          },
          {
            label: '检查端口',
            value: ['响应超时(2秒)', '检查间隔(5秒)', '不健康阈值(3次)', '健康阈值(3次),'],
          },
        ],
      },
    ];

    return () => (
      <div class='listener-detail-wrap'>
        <Button class='fixed-edit-btn' outline theme='primary'>
          编辑
        </Button>
        {listenerDetailInfo.map(({ title, open_state, content, hidden }) => {
          if (hidden) {
            return null;
          }
          return (
            <div class='listener-detail-info-wrap'>
              <div class='info-title'>
                {title} {/*  eslint-disable-next-line no-nested-ternary */}
                {open_state === undefined ? null : open_state ? (
                  <Tag theme='success' class='status-tag'>
                    已开启
                  </Tag>
                ) : (
                  <Tag class='status-tag'>未开启</Tag>
                )}{' '}
              </div>
              <div class='info-content'>
                {open_state !== false &&
                  content.map(({ label, value, sub_hidden }) => {
                    if (sub_hidden) {
                      return null;
                    }
                    let valueVNode = null;
                    if (typeof value === 'function') {
                      valueVNode = value();
                    } else {
                      if (Array.isArray(value)) {
                        valueVNode = value.map((v) => (
                          <>
                            {' '}
                            {v};<br />{' '}
                          </>
                        ));
                      } else {
                        valueVNode = value;
                      }
                    }
                    return (
                      <div class='info-item'>
                        <div class='info-item-label'>{label}</div>:
                        <div class={`info-item-content${Array.isArray(value) ? ' multiline' : ''}`}>{valueVNode}</div>
                      </div>
                    );
                  })}
              </div>
            </div>
          );
        })}
      </div>
    );
  },
});
