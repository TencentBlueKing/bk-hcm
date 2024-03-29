import { defineComponent, watch } from 'vue';
// import components
import { Button, Form, Input, Message, Select, Switcher, Tag } from 'bkui-vue';
import BkRadio, { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { Plus } from 'bkui-vue/lib/icon';
import CommonSideslider from '@/components/common-sideslider';
import BatchOperationDialog from '@/components/batch-operation-dialog';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useResourceStore } from '@/store';
// import custom hooks
import { useTable } from '@/hooks/useTable/useTable';
import { useI18n } from 'vue-i18n';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import useAddOrUpdateListener from './useAddOrUpdateListener';
import useBatchDeleteListener from './useBatchDeleteListener';
// import types
import { DoublePlainObject } from '@/typings';
import './index.scss';
import Confirm from '@/components/confirm';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  setup() {
    // use hooks
    const { t } = useI18n();
    const { whereAmI } = useWhereAmI();
    const { selections, handleSelectionChange, resetSelections } = useSelection();

    const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => {
      if (whereAmI.value === Senarios.business) return true;
      if (row.id) {
        return row.bk_biz_id === -1;
      }
    };

    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const resourceStore = useResourceStore();

    // listener - table
    const { columns, settings } = useColumns('listener');
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [
          {
            name: '监听器名称',
            id: 'name',
          },
          {
            name: '协议',
            id: 'protocol',
          },
          {
            name: '端口',
            id: 'port',
          },
          {
            name: '均衡方式',
            id: 'scheduler',
          },
          {
            name: '域名数量',
            id: 'domain_num',
          },
          {
            name: 'URL数量',
            id: 'url_num',
          },
          {
            name: '同步状态',
            id: 'syncStatus',
          },
        ],
      },
      tableOptions: {
        columns: [
          {
            type: 'selection',
            width: 32,
            minWidth: 32,
            onlyShowOnList: true,
            align: 'right',
          },
          ...columns,
          {
            label: t('操作'),
            field: 'actions',
            render: ({ data }: any) => (
              <div class='operate-groups'>
                <Button text theme='primary' onClick={() => handleEditListener(data.id)}>
                  {t('编辑')}
                </Button>
                <Button text theme='primary' onClick={() => handleDeleteListener(data)}>
                  {t('删除')}
                </Button>
              </div>
            ),
          },
        ],
        extra: {
          settings: settings.value,
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
        },
      },
      requestOption: {
        type: `load_balancers/${loadBalancerStore.currentSelectedTreeNode.id}/listeners`,
      },
    });
    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { id, type } = val;
        if (type !== 'lb') return;
        // 只有当 type='lb' 时, 才去请求对应 lb 下的 listener 列表
        getListData([], `load_balancers/${id}/listeners`);
      },
    );

    // 「新增/编辑」监听器
    const protocolButtonList = ['TCP', 'UDP', 'HTTP', 'HTTPS'];
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
    } = useAddOrUpdateListener(getListData);

    // 删除监听器
    const handleDeleteListener = (data: any) => {
      Confirm('请确定删除监听器', `将删除监听器【${data.name}】`, () => {
        resourceStore.deleteBatch('listeners', { ids: [data.id] }).then(() => {
          Message({ theme: 'success', message: '删除成功' });
          getListData();
        });
      });
    };

    // 批量删除监听器
    const {
      isSubmitLoading,
      isBatchDeleteDialogShow,
      radioGroupValue,
      tableProps,
      handleBatchDeleteListener,
      handleBatchDeleteSubmit,
    } = useBatchDeleteListener(columns, selections, resetSelections, getListData);

    return () => (
      <div>
        {/* 监听器list */}
        <CommonTable class='has-selection'>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'} onClick={handleAddListener}>
                  <Plus class={'f20'} />
                  {t('新增监听器')}
                </Button>
                <Button disabled={selections.value.length === 0} onClick={handleBatchDeleteListener}>
                  {t('批量删除')}
                </Button>
              </div>
            ),
          }}
        </CommonTable>

        {/* 新增/编辑监听器 */}
        <CommonSideslider
          class='listener-sideslider-container'
          v-model:isShow={isSliderShow.value}
          title={isEdit.value ? '编辑监听器' : '新增监听器'}
          width={640}
          onHandleSubmit={handleAddOrUpdateListener}
          isSubmitDisabled={isAddOrUpdateListenerSubmit.value}>
          <Form ref={formRef} formType='vertical' model={listenerFormData} rules={rules}>
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
              <Input v-model={listenerFormData.port} type='number' placeholder={t('请输入')} disabled={isEdit.value} />
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
                    <FormItem label={t('服务器证书')} required property='certificate.ca_cloud_id'>
                      <Select
                        v-model={listenerFormData.certificate.ca_cloud_id}
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
                    <FormItem label={t('CA证书')} required property='certificate.cert_cloud_ids'>
                      <Select
                        v-model={listenerFormData.certificate.cert_cloud_ids}
                        multiple
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
                  </>
                )}
              </>
            )}
            {['HTTP', 'HTTPS'].includes(listenerFormData.protocol) && !isEdit.value && (
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
                    <Option id='IP_HASH' name={t('IP Hash')} />
                  </Select>
                </FormItem>
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
                <FormItem label={t('目标组')} required property='target_group_id'>
                  <Select
                    v-model={listenerFormData.target_group_id}
                    scrollLoading={isTargetGroupListLoading.value}
                    onScroll-end={handleTargetGroupListScrollEnd}>
                    {targetGroupList.value.map(({ id, name }) => (
                      <Option key={id} id={id} name={name} />
                    ))}
                  </Select>
                </FormItem>
              </>
            )}
          </Form>
        </CommonSideslider>

        {/* 批量删除监听器 */}
        <BatchOperationDialog
          class='batch-delete-listener-dialog'
          v-model:isShow={isBatchDeleteDialogShow.value}
          title={t('批量删除监听器')}
          theme='danger'
          confirmText='删除'
          isSubmitLoading={isSubmitLoading.value}
          tableProps={tableProps}
          onHandleConfirm={handleBatchDeleteSubmit}>
          {{
            tips: () => (
              <>
                已选择<span class='blue'>{tableProps.data.length}</span>个监听器，其中
                <span class='red'>
                  {
                    tableProps.data.filter(
                      ({ rs_zero_num, rs_not_zero_num }) => rs_not_zero_num === rs_zero_num + rs_not_zero_num,
                    ).length
                  }
                </span>
                个监听器RS的权重均不为0，在删除监听器前，请确认是否有流量转发，仔细核对后，再提交删除。
              </>
            ),
            tab: () => (
              <BkRadioGroup v-model={radioGroupValue.value}>
                <BkRadioButton label={false}>{t('权重为0')}</BkRadioButton>
                <BkRadioButton label={true}>{t('权重不为0')}</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </BatchOperationDialog>
      </div>
    );
  },
});
