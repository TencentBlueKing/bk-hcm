import { watch } from 'vue';
import { APPLICATION_LAYER_LIST } from '@/constants';

export default (listenerFormData: any) => {
  watch(
    () => listenerFormData.protocol,
    (val) => {
      // 七层监听器不支持会话保持
      APPLICATION_LAYER_LIST.includes(val) && (listenerFormData.session_open = false);
    },
  );

  watch(
    () => listenerFormData.scheduler,
    (val) => {
      // 如果均衡方式为加权最小连接数, 不支持配置会话保持
      val === 'LEAST_CONN' && (listenerFormData.session_open = false);
    },
  );

  watch(
    () => listenerFormData.session_open,
    (val) => {
      // session_expire传0即为关闭会话保持
      val ? (listenerFormData.session_expire = 30) : (listenerFormData.session_expire = 0);
    },
  );

  watch(
    () => listenerFormData.certificate.ssl_mode,
    (val) => {
      if (!listenerFormData.certificate) return;
      // 如果需要客户端也提供证书, 则需要SSL认证类型为双向认证
      val !== 'MUTUAL' && (listenerFormData.certificate.ca_cloud_id = '');
    },
    { deep: true },
  );
};
