import { PropType, defineComponent, onMounted, onUnmounted } from 'vue';
// import components
import { Form, Input, Select, Switcher, Tag } from 'bkui-vue';
import BkRadio, { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import CommonSideslider from '@/components/common-sideslider';
// import stores
import { useLoadBalancerStore } from '@/store';
// import hooks
import { useI18n } from 'vue-i18n';
import useAddOrUpdateListener from './useAddOrUpdateListener';
// import utils
import bus from '@/common/bus';
// import constants
import { APPLICATION_LAYER_LIST } from '@/constants';
import './index.scss';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  name: 'AddOrUpdateListenerSideslider',
  props: {
    getListData: Function as PropType<(...args: any) => any>,
  },
  setup(props) {
    const loadBalancerStore = useLoadBalancerStore();
    const protocolButtonList = ['TCP', 'UDP', 'HTTP', 'HTTPS'];
    // use hooks
    const { t } = useI18n();

    // 「新增/编辑」监听器
    const {
      isSliderShow,
      isEdit,
      isAddOrUpdateListenerSubmit,
      isSniOpen,
      formRef,
      rules,
      listenerFormData,
      handleAddListener,
      handleEditListener,
      handleAddOrUpdateListener,
      isTargetGroupListLoading,
      targetGroupList,
      handleTargetGroupListScrollEnd,
      isSVRCertListLoading,
      SVRCertList,
      handleSVRCertListScrollEnd,
      isCACertListLoading,
      CACertList,
      handleCACertListScrollEnd,
    } = useAddOrUpdateListener(props.getListData);

    onMounted(() => {
      bus.$on('showAddListenerSideslider', handleAddListener);
      bus.$on('showEditListenerSideslider', handleEditListener);
    });

    onUnmounted(() => {
      bus.$off('showAddListenerSideslider');
      bus.$off('showEditListenerSideslider');
    });

    return () => (
      <CommonSideslider
        class='listener-sideslider-container'
        v-model:isShow={isSliderShow.value}
        title={isEdit.value ? '编辑监听器' : '新增监听器'}
        width={640}
        onHandleSubmit={handleAddOrUpdateListener}
        isSubmitDisabled={isAddOrUpdateListenerSubmit.value}>
        <Form ref={formRef} formType='vertical' model={listenerFormData} rules={rules}>
          <FormItem label={t('负载均衡名称')}>
            <Input
              modelValue={
                loadBalancerStore.currentSelectedTreeNode.type === 'lb'
                  ? loadBalancerStore.currentSelectedTreeNode.name
                  : loadBalancerStore.currentSelectedTreeNode.lb.name
              }
              disabled
            />
          </FormItem>
          <FormItem label={t('监听器名称')} required property='name'>
            <Input v-model={listenerFormData.name} placeholder={t('请输入')} />
            <div class='form-item-tips'>
              {t('不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号')}
            </div>
          </FormItem>

          <FormItem label={t('监听协议')} required property='protocol'>
            <BkRadioGroup v-model={listenerFormData.protocol} type='card' disabled={isEdit.value}>
              {protocolButtonList.map((protocol) => (
                <BkRadioButton key={protocol} label={protocol}>
                  {protocol}
                </BkRadioButton>
              ))}
            </BkRadioGroup>
          </FormItem>
          <FormItem label={t('监听端口')} required property='port'>
            <Input
              v-model={listenerFormData.port}
              type='number'
              placeholder={t('请输入')}
              disabled={isEdit.value}
              class='no-number-control'
            />
          </FormItem>
          {listenerFormData.protocol === 'HTTPS' && (
            <>
              <div class={'flex-row justify-content-between'}>
                <FormItem label={t('SNI')} required property='sni_switch'>
                  <Switcher
                    disabled={isSniOpen.value}
                    theme='primary'
                    v-model={listenerFormData.sni_switch}
                    trueValue={1}
                    falseValue={0}
                  />
                </FormItem>
                {listenerFormData.sni_switch === 0 && (
                  <FormItem label={t('SSL解析方式')} required property='certificate.ssl_mode'>
                    <BkRadioGroup v-model={listenerFormData.certificate.ssl_mode}>
                      <BkRadio label='UNIDIRECTIONAL'>
                        {t('单向认证')}
                        <Tag theme='info' class='recommend-tag'>
                          {t('推荐')}
                        </Tag>
                      </BkRadio>
                      <BkRadio label='MUTUAL' class='ml24'>
                        {t('双向认证')}
                      </BkRadio>
                    </BkRadioGroup>
                  </FormItem>
                )}
              </div>
              {listenerFormData.sni_switch === 0 && (
                <>
                  <FormItem label={t('服务器证书')} required property='certificate.cert_cloud_ids'>
                    <Select
                      v-model={listenerFormData.certificate.cert_cloud_ids}
                      multiple
                      scrollLoading={isSVRCertListLoading.value}
                      onScroll-end={handleSVRCertListScrollEnd}>
                      {SVRCertList.value
                        .sort((a, b) => a.cert_status - b.cert_status)
                        .map(({ cloud_id, name, cert_status }) => (
                          <Option key={cloud_id} id={cloud_id} name={name} disabled={cert_status === '3'}>
                            {name}
                            {cert_status === '3' && (
                              <Tag theme='danger' style={{ marginLeft: '12px' }}>
                                已过期
                              </Tag>
                            )}
                          </Option>
                        ))}
                    </Select>
                  </FormItem>
                  {listenerFormData.certificate.ssl_mode === 'MUTUAL' && (
                    <FormItem label={t('CA证书')} required property='certificate.ca_cloud_id'>
                      <Select
                        v-model={listenerFormData.certificate.ca_cloud_id}
                        scrollLoading={isCACertListLoading.value}
                        onScroll-end={handleCACertListScrollEnd}>
                        {CACertList.value
                          .sort((a, b) => a.cert_status - b.cert_status)
                          .map(({ cloud_id, name, cert_status }) => (
                            <Option key={cloud_id} id={cloud_id} name={name} disabled={cert_status === '3'}>
                              {name}
                              {cert_status === '3' && (
                                <Tag theme='danger' style={{ marginLeft: '12px' }}>
                                  已过期
                                </Tag>
                              )}
                            </Option>
                          ))}
                      </Select>
                    </FormItem>
                  )}
                </>
              )}
            </>
          )}
          {APPLICATION_LAYER_LIST.includes(listenerFormData.protocol) && !isEdit.value && (
            <>
              <FormItem label={t('默认域名')} required property='domain'>
                <Input v-model={listenerFormData.domain} placeholder={t('请输入')} />
              </FormItem>
              <FormItem label={t('URL路径')} required property='url'>
                <Input v-model={listenerFormData.url} placeholder={t('请输入')} />
              </FormItem>
            </>
          )}
          {!isEdit.value && (
            <>
              <FormItem label={t('均衡方式')} required property='scheduler'>
                <Select v-model={listenerFormData.scheduler}>
                  <Option id='WRR' name={t('按权重轮询')} />
                  <Option id='LEAST_CONN' name={t('最小连接数')} />
                  {APPLICATION_LAYER_LIST.includes(listenerFormData.protocol) && (
                    <Option id='IP_HASH' name={t('IP Hash')} />
                  )}
                </Select>
              </FormItem>
              {
                // 七层协议无会话保持
                !APPLICATION_LAYER_LIST.includes(listenerFormData.protocol) &&
                  // 均衡方式为加权最小连接数，不支持配置会话保持
                  listenerFormData.scheduler !== 'LEAST_CONN' && (
                    <div class={'flex-row'}>
                      <FormItem label={t('会话保持')} required property='session_open'>
                        <Switcher theme='primary' v-model={listenerFormData.session_open} />
                      </FormItem>
                      <FormItem label={t('保持时间')} class={'ml40'} required property='session_expire'>
                        <Input
                          v-model={listenerFormData.session_expire}
                          disabled={!listenerFormData.session_open}
                          placeholder={t('请输入')}
                          type='number'
                          min={30}
                          suffix='秒'
                        />
                      </FormItem>
                    </div>
                  )
              }
              <FormItem label={t('目标组')} required property='target_group_id'>
                <Select
                  v-model={listenerFormData.target_group_id}
                  scrollLoading={isTargetGroupListLoading.value}
                  onScroll-end={handleTargetGroupListScrollEnd}>
                  {targetGroupList.value.map(({ id, name, listener_num }) => (
                    <Option key={id} id={id} name={name} disabled={listener_num > 0} />
                  ))}
                </Select>
              </FormItem>
            </>
          )}
        </Form>
      </CommonSideslider>
    );
  },
});
