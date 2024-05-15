import { PropType, defineComponent, onMounted, onUnmounted } from 'vue';
// import components
import { Alert, Button, Divider, Form, Input, Select, Switcher, Tag } from 'bkui-vue';
import BkRadio, { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { Plus, RightTurnLine, Spinner } from 'bkui-vue/lib/icon';
import CommonSideslider from '@/components/common-sideslider';
// import stores
import { useLoadBalancerStore, useAccountStore } from '@/store';
// import hooks
import { useI18n } from 'vue-i18n';
import useAddOrUpdateListener from './useAddOrUpdateListener';
// import utils
import bus from '@/common/bus';
import { goAsyncTaskDetail } from '@/utils';
// import constants
import { APPLICATION_LAYER_LIST } from '@/constants';
import { IOriginPage } from '@/typings';
import './index.scss';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  name: 'AddOrUpdateListenerSideslider',
  props: {
    // 标识组件在哪个页面中使用, lb=具体的负载均衡, listener=具体的监听器
    originPage: String as PropType<IOriginPage>,
    getListData: Function as PropType<(...args: any) => any>,
  },
  setup(props) {
    const loadBalancerStore = useLoadBalancerStore();
    const accountStore = useAccountStore();
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
      isTargetGroupListFlashLoading,
      handleTargetGroupListRefreshOptionList,
      isSVRCertListLoading,
      SVRCertList,
      handleSVRCertListScrollEnd,
      isCACertListLoading,
      CACertList,
      handleCACertListScrollEnd,
      isLbLocked,
      lockedLbInfo,
    } = useAddOrUpdateListener(props.getListData, props.originPage);

    // click-handler - 新增目标组
    const handleAddTargetGroup = () => {
      const url = `/#/business/loadbalancer/group-view?bizs=${accountStore.bizs}`;
      window.open(url, '_blank');
    };

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
        isSubmitLoading={isAddOrUpdateListenerSubmit.value}>
        {isLbLocked.value ? (
          <Alert theme='danger' class={'mb24'}>
            当前负载均衡正在变更中，不允许新增监听器，
            <Button text theme='primary' onClick={() => goAsyncTaskDetail(lockedLbInfo.value.flow_id)}>
              查看当前任务
            </Button>
            。
          </Alert>
        ) : null}
        <Form ref={formRef} formType='vertical' model={listenerFormData} rules={rules}>
          <FormItem label={t('负载均衡名称')}>
            <Input
              modelValue={
                props.originPage === 'lb'
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
              </div>
              <FormItem label={t('服务器证书')} required property='certificate.cert_cloud_ids'>
                <Select
                  v-model={listenerFormData.certificate.cert_cloud_ids}
                  multiple
                  scrollLoading={isSVRCertListLoading.value}
                  onScroll-end={handleSVRCertListScrollEnd}>
                  {SVRCertList.value
                    .sort((a, b) => a.cert_status - b.cert_status)
                    .map(({ cloud_id, name, cert_status, domain, encrypt_algorithm }) => (
                      <Option key={cloud_id} id={cloud_id} name={name} disabled={cert_status === '3'}>
                        {name}&nbsp;(主域名 : {domain ? domain[0] : '--'}, 备用域名：{domain ? domain[1] : '--'})
                        {cert_status === '3' ? (
                          <Tag theme='danger' style={{ marginLeft: '12px' }}>
                            已过期
                          </Tag>
                        ) : (
                          <Tag theme='info' style={{ marginLeft: '12px' }}>
                            {encrypt_algorithm}
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
                      <FormItem
                        label={t('会话保持')}
                        required
                        property='session_open'
                        description='会话保持可使得来自同一 IP 的请求被转发到同一台后端服务器上。参考官方文档https://cloud.tencent.com/document/product/214/6154'>
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
                {/* tag: 这里暂时不抽离成公共组件, 等后续相关场景多了再抽离, 此处先写行内样式, 方便抽离时设置css样式 */}
                <Select
                  v-model={listenerFormData.target_group_id}
                  scrollLoading={isTargetGroupListLoading.value}
                  onScroll-end={handleTargetGroupListScrollEnd}>
                  {{
                    default: () =>
                      targetGroupList.value.map(({ id, name, listener_num }) => (
                        <Option key={id} id={id} name={name} disabled={listener_num > 0} />
                      )),
                    extension: () => (
                      <div style='width: 100%; color: #63656E; padding: 0 12px;'>
                        <div style='display: flex; align-items: center;justify-content: center;'>
                          <span
                            style='display: flex; align-items: center;cursor: pointer;'
                            onClick={handleAddTargetGroup}>
                            <Plus style='font-size: 20px;' />
                            新增
                          </span>
                          <span style='display: flex; align-items: center;position: absolute; right: 12px;'>
                            <Divider direction='vertical' type='solid' />
                            {isTargetGroupListFlashLoading.value ? (
                              <Spinner style='font-size: 14px;color: #3A84FF;' />
                            ) : (
                              <RightTurnLine
                                style='font-size: 14px;cursor: pointer;'
                                onClick={handleTargetGroupListRefreshOptionList}
                              />
                            )}
                          </span>
                        </div>
                      </div>
                    ),
                  }}
                </Select>
              </FormItem>
            </>
          )}
          {isEdit.value && (
            <div class='binded-target-group-show-container'>
              <span class='label'>{t('已绑定的目标组')}</span>:{' '}
              <span
                class='ml10 link-text-btn'
                onClick={() => {
                  window.open(
                    `/#/business/loadbalancer/group-view/${listenerFormData.target_group_id}?bizs=${accountStore.bizs}&type=detail`,
                    '_blank',
                    'noopener,noreferrer',
                  );
                }}>
                {listenerFormData.target_group_name || '未命名'}
              </span>
            </div>
          )}
        </Form>
      </CommonSideslider>
    );
  },
});
