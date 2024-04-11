import { watch } from 'vue';

export default (listenerFormData: any) => {
  watch(
    () => listenerFormData.session_open,
    (val) => {
      // session_expire传0即为关闭会话保持
      val ? (listenerFormData.session_expire = 30) : (listenerFormData.session_expire = 0);
    },
  );

  watch(
    () => listenerFormData.sni_switch,
    (val) => {
      // 如果sni开启, 则不需要提供证书信息
      val === 1 &&
        Object.assign(listenerFormData.certificate, {
          ca_cloud_id: '',
          cert_cloud_ids: [],
        });
    },
  );

  watch(
    () => listenerFormData.certificate.ssl_mode,
    (val) => {
      // 如果需要客户端也提供证书, 则需要SSL认证类型为双向认证
      val !== 'MUTUAL' && (listenerFormData.certificate.ca_cloud_id = '');
    },
    { deep: true },
  );
};
