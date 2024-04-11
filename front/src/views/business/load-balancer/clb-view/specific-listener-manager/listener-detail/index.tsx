import { computed, defineComponent, reactive, watch } from 'vue';
// import components
import { Button, Tag } from 'bkui-vue';
import StatusLoading from '@/assets/image/status_loading.png';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useResourceStore } from '@/store';
// import hooks
import { useI18n } from 'vue-i18n';
// import utils
import { timeFormatter } from '@/common/util';
// import constants
import { SCHEDULER_MAP, SESSION_TYPE_MAP, SSL_MODE_MAP, TRANSPORT_LAYER_LIST } from '@/constants/clb';
import './index.scss';

export default defineComponent({
  name: 'ListenerDetail',
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const resourceStore = useResourceStore();

    // define data
    const listenerDetail = reactive<any>({}); // 监听器详情

    const listenerDetailInfoOption = computed(() => [
      {
        title: t('基本信息'),
        content: [
          {
            label: t('监听器名称'),
            value: listenerDetail.lbl_name,
          },
          {
            label: t('监听器ID'),
            value: listenerDetail.cloud_lbl_id,
          },
          {
            label: t('协议端口'),
            value: `${listenerDetail.protocol}:${listenerDetail.port}`,
          },
          {
            label: t('域名数量'),
            value: listenerDetail.domain_num,
            sub_hidden: TRANSPORT_LAYER_LIST.includes(listenerDetail.protocol),
          },
          {
            label: t('URL 数量'),
            value: listenerDetail.url_num,
            sub_hidden: TRANSPORT_LAYER_LIST.includes(listenerDetail.protocol),
          },
          {
            label: t('均衡方式'),
            value: SCHEDULER_MAP[listenerDetail.scheduler],
          },
          {
            label: t('目标组'),
            value: () => (
              <div class='target-group-wrap'>
                <span class='link-text-btn'>{listenerDetail.target_group_name}</span>
                {true && <img class='loading-icon spin-icon' src={StatusLoading} alt='' />}
              </div>
            ),
            sub_hidden: ['HTTP', 'HTTPS'].includes(listenerDetail.protocol),
          },
          {
            label: t('创建时间'),
            value: timeFormatter(listenerDetail.created_at),
          },
        ],
      },
      {
        title: t('证书信息'),
        hidden: ['HTTP', 'TCP', 'UDP'].includes(listenerDetail.protocol),
        content: [
          {
            label: t('认证方式'),
            value: SSL_MODE_MAP[listenerDetail.certificate?.ssl_mode],
          },
          {
            label: t('服务器证书'),
            value: listenerDetail.certificate?.ca_cloud_id,
          },
          {
            label: t('CA证书'),
            value: listenerDetail.certificate?.cert_cloud_ids?.join(','),
          },
        ],
      },
      {
        title: t('会话保持'),
        open_state: listenerDetail.session_expire === 0 ? 0 : 1,
        content: [
          {
            label: t('会话时间'),
            value: `${SESSION_TYPE_MAP[listenerDetail.session_type]}${listenerDetail.session_expire} 秒`,
          },
        ],
      },
      {
        title: t('健康检查'),
        open_state: listenerDetail.health_check?.health_switch === 1 ? 1 : 0,
        content: [
          {
            label: t('健康探测源IP'),
            value: listenerDetail.health_check?.source_ip_type === 0 ? '负载均衡 VIP' : '100.64.0.0/10网段',
          },
          {
            label: t('检查方式'),
            value: listenerDetail.health_check?.check_type,
          },
          {
            label: t('检查端口'),
            value: listenerDetail.health_check?.check_port,
          },
          {
            label: t('检查选型'),
            value: [
              `响应超时(${listenerDetail.health_check?.time_out}秒)`,
              `检查间隔(${listenerDetail.health_check?.interval_time}秒)`,
              `不健康阈值(${listenerDetail.health_check?.un_health_num}秒)`,
              `健康阈值(${listenerDetail.health_check?.health_num}秒)`,
            ],
          },
        ],
      },
    ]);

    // 获取监听器详情
    const getListenerDetail = async (id: string) => {
      const res = await resourceStore.detail('listeners', id);
      Object.assign(listenerDetail, res.data);
    };

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { id, type } = val;
        if (type !== 'listener') return;
        // 只有 type='listener' 时, 才请求对应 listener 的详情
        getListenerDetail(id);
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='listener-detail-wrap'>
        <Button class='fixed-edit-btn' outline theme='primary'>
          {t('编辑')}
        </Button>
        {listenerDetailInfoOption.value.map(({ title, open_state, content, hidden }) => {
          if (hidden) {
            return null;
          }
          return (
            <div class='listener-detail-info-wrap'>
              <div class='info-title'>
                {title}
                {open_state === 1 && (
                  <Tag theme='success' class='status-tag'>
                    {t('已开启')}
                  </Tag>
                )}
                {open_state === 0 && <Tag class='status-tag'>{t('未开启')}</Tag>}
              </div>
              <div class='info-content'>
                {open_state !== 0 &&
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
                            {v};<br />
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
