import { PropType, computed, defineComponent, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
// import components
import { Alert, Button, Form, Input, Select, Switcher, Tag } from 'bkui-vue';
import BkRadio, { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import CommonSideslider from '@/components/common-sideslider';
import TargetGroupSelector from '../TargetGroupSelector';
import CertSelector from '../CertSelector';
// import stores
import { useLoadBalancerStore, useAccountStore, useBusinessStore } from '@/store';
// import hooks
import { useI18n } from 'vue-i18n';
import useAddOrUpdateListener from './useAddOrUpdateListener';
import { useWhereAmI } from '@/hooks/useWhereAmI';
// import utils
import bus from '@/common/bus';
import { goAsyncTaskDetail } from '@/utils';
// import constants
import { APPLICATION_LAYER_LIST, TRANSPORT_LAYER_LIST } from '@/constants';
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
    const businessStore = useBusinessStore();
    const protocolButtonList = ['TCP', 'UDP', 'HTTP', 'HTTPS'];
    // use hooks
    const { t } = useI18n();
    const { getBizsId } = useWhereAmI();

    const currentBusinessId = computed(() => {
      return loadBalancerStore.currentSelectedTreeNode?.bk_biz_id ?? getBizsId();
    });
    const computedProtocol = computed(() => listenerFormData.protocol);
    const targetGroupSelectorRef = ref();
    const svrCertSelectorRef = ref();

    // 「新增/编辑」监听器
    const {
      isSliderShow,
      isEdit,
      isAddOrUpdateListenerSubmit,
      isSniOpen,
      formRef,
      listenerFormData,
      handleAddListener,
      handleEditListener,
      handleAddOrUpdateListener,
      isLbLocked,
      lockedLbInfo,
    } = useAddOrUpdateListener(props.getListData, props.originPage);

    const rules = {
      name: [
        {
          validator: (value: string) => /^[\u4e00-\u9fa5A-Za-z0-9\-._:]{1,60}$/.test(value),
          message: '不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号',
          trigger: 'change',
        },
      ],
      port: [
        {
          validator: (value: number) => value >= 1 && value <= 65535,
          message: '端口号不符合规范',
          trigger: 'change',
        },
      ],
      domain: [
        {
          validator: (value: string) => /^(?:(?:[a-zA-Z0-9]+-?)+(?:\.[a-zA-Z0-9-]+)+)$/.test(value),
          message: '域名不符合规范',
          trigger: 'change',
        },
      ],
      url: [
        {
          validator: (value: string) => /^\/[\w\-/]*$/.test(value),
          message: 'URL路径不符合规范',
          trigger: 'change',
        },
      ],
      'certificate.cert_cloud_ids': [
        {
          validator: (value: string[]) => value.length <= 2,
          message: '最多选择 2 个证书',
          trigger: 'change',
        },
        {
          validator: (value: string[]) => {
            // 判断证书类型是否重复
            const [cert1, cert2] = svrCertSelectorRef.value.dataList.filter((cert: any) =>
              value.includes(cert.cloud_id),
            );
            return cert1?.encrypt_algorithm !== cert2?.encrypt_algorithm;
          },
          message: '不能选择加密算法相同的证书',
          trigger: 'change',
        },
      ],
    };

    // 当侧边栏显示或协议变更时, 刷新目标组select-option-list
    watch([isSliderShow, () => listenerFormData.protocol], ([isSliderShow]) => {
      if (!isSliderShow || isEdit.value) return;
      // 重置目标组
      listenerFormData.target_group_id = '';
      nextTick(() => {
        targetGroupSelectorRef.value.handleRefresh();
        formRef.value?.clearValidate();
      });
    });

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
            <Button
              text
              theme='primary'
              onClick={() =>
                goAsyncTaskDetail(businessStore.list, lockedLbInfo.value.flow_id, currentBusinessId.value)
              }>
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
              v-model_number={listenerFormData.port}
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
                <CertSelector
                  ref={svrCertSelectorRef}
                  v-model={listenerFormData.certificate.cert_cloud_ids}
                  type='SVR'
                  accountId={listenerFormData.account_id}
                />
              </FormItem>
              {listenerFormData.certificate.ssl_mode === 'MUTUAL' && (
                <FormItem label={t('CA证书')} required property='certificate.ca_cloud_id'>
                  <CertSelector
                    v-model={listenerFormData.certificate.ca_cloud_id}
                    type='CA'
                    accountId={listenerFormData.account_id}
                  />
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
          {/* 新增监听器 */}
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
                          v-model_number={listenerFormData.session_expire}
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
              {/* 四层、七层都支持绑定目标组：四层绑定在监听器上，七层绑定在url上 */}
              <FormItem label={t('目标组')} required property='target_group_id'>
                <TargetGroupSelector
                  ref={targetGroupSelectorRef}
                  v-model={listenerFormData.target_group_id}
                  accountId={listenerFormData.account_id}
                  cloudVpcId={
                    props.originPage === 'lb'
                      ? loadBalancerStore.currentSelectedTreeNode.cloud_vpc_id
                      : loadBalancerStore.currentSelectedTreeNode.lb.cloud_vpc_id
                  }
                  region={
                    props.originPage === 'lb'
                      ? loadBalancerStore.currentSelectedTreeNode.region
                      : loadBalancerStore.currentSelectedTreeNode.lb.region
                  }
                  protocol={computedProtocol.value}
                  isCorsV2={
                    props.originPage === 'lb'
                      ? loadBalancerStore.currentSelectedTreeNode.extension.snat_pro
                      : loadBalancerStore.currentSelectedTreeNode.lb.extension.snat_pro
                  }
                />
              </FormItem>
            </>
          )}
          {/* 编辑监听器，四层监听器显示目标组信息、七层不显示 */}
          {isEdit.value && TRANSPORT_LAYER_LIST.includes(listenerFormData.protocol) && (
            <div class='binded-target-group-show-container'>
              <span class='label'>{t('已绑定的目标组')}</span>:
              {listenerFormData.target_group_id ? (
                <span
                  class='ml10 link-text-btn'
                  onClick={() => {
                    window.open(
                      `/#/business/loadbalancer/group-view/${listenerFormData.target_group_id}?bizs=${accountStore.bizs}&type=detail&vendor=${listenerFormData.vendor}`,
                      '_blank',
                      'noopener,noreferrer',
                    );
                  }}>
                  {listenerFormData.target_group_name || '未命名'}
                </span>
              ) : (
                <span class='ml10'>--</span>
              )}
            </div>
          )}
        </Form>
      </CommonSideslider>
    );
  },
});
