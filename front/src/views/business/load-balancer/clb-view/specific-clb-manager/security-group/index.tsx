import { PropType, TransitionGroup, computed, defineComponent, reactive, ref, watch, watchEffect } from 'vue';
import './index.scss';
import { Alert, Button, Dialog, Exception, Input, Message, Tag } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import CommonSideslider from '@/components/common-sideslider';
import CommonDialog from '@/components/common-dialog';
import { useAccountStore, useBusinessStore } from '@/store';
import { EditLine, Plus, Success } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useTable/useTable';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useLoadBalancerStore } from '@/store/loadbalancer';
import ExpandCard from './expand-card';
import { QueryRuleOPEnum } from '@/typings';
import { IDetail } from '@/hooks/useRouteLinkBtn';
import { VueDraggable } from 'vue-draggable-plus';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import DraggableItem from './draggable-item';
import { cloneDeep } from 'lodash';

export enum SecurityRuleDirection {
  in = 'ingress',
  out = 'egress',
}

export default defineComponent({
  props: {
    detail: Object as PropType<IDetail>,
    getDetails: Function,
    updateLb: Function,
    id: String,
  },
  setup(props) {
    const isPassToTarget = ref(false);
    const securityRuleType = ref(SecurityRuleDirection.in);
    const isSideSliderShow = ref(false);
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();
    const { selections, handleSelectionChange, resetSelections } = useSelection();
    const isAllExpand = ref(true);
    const securitySearchVal = ref('');
    const searchVal = ref('');
    const selectedSecuirtyGroupsSet = ref(new Set());
    const bindedSecurityGroups = ref([]);
    const isUpdating = ref(false);
    const isSubmitLoading = ref(false);
    const isConfigDialogShow = ref(false);
    const tmpIsPassToTarget = ref(isPassToTarget.value);
    const securityGroups = ref([]);
    const isDialogShow = ref(false);
    const bindedSet = reactive(new Set());
    const loadBalancerStore = useLoadBalancerStore();
    const el = ref();
    const hanldeSubmit = async () => {
      try {
        isSubmitLoading.value = true;
        await businessStore.bindSecurityToCLB({
          bk_biz_id: accountStore.bizs,
          lb_id: loadBalancerStore.currentSelectedTreeNode.id,
          security_group_ids: securityGroups.value.map(({ id }) => id),
        });
        getBindedSecurityList();
        selectedSecuirtyGroupsSet.value = new Set();
        isSideSliderShow.value = false;
        resetSelections();
        Message({
          message: '绑定成功',
          theme: 'success',
        });
      } finally {
        isSubmitLoading.value = false;
      }
    };

    // 检查并转义正则特殊字符
    const escapeRegExp = (str: string) => {
      return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    };

    const securityRulesSearchedResults = computed(() => {
      const val = searchVal.value;
      if (!val.trim()) return bindedSecurityGroups.value;
      const reg = new RegExp(escapeRegExp(val));
      return bindedSecurityGroups.value.filter((v) => reg.test(`${v.name} (${v.cloud_id})`));
    });

    const securitySearchedList = ref([]);

    watchEffect(() => {
      const val = securitySearchVal.value;
      if (!val.trim()) {
        securitySearchedList.value = [];
        return;
      }
      const reg = new RegExp(escapeRegExp(val));
      securitySearchedList.value = securityGroups.value.filter((v) => reg.test(`${v.name} (${v.cloud_id})`));
    });

    const tableColumns = [
      { type: 'selection', width: 30, minWidth: 30 },
      {
        label: '安全组名称',
        field: 'name',
      },
      {
        label: 'ID',
        field: 'cloud_id',
      },
      {
        label: '备注',
        field: 'memo',
      },
    ];
    const searchData: ISearchItem[] = [
      {
        id: 'name',
        name: '安全组名称',
      },
      {
        id: 'cloud_id',
        name: 'ID',
      },
    ];

    const isRowSelectEnable = ({ row, isCheckAll }: any) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };

    const isCurRowSelectEnable = (row: any) => {
      return (
        !bindedSecurityGroups.value.map((v) => v.id).includes(row.id) && !selectedSecuirtyGroupsSet.value.has(row.id)
      );
    };

    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData,
        extra: {
          searchSelectExtStyle: {
            width: '100%',
          },
        },
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          selectionKey: 'cloud_id',
        },
      },
      requestOption: {
        type: 'security_groups',
        filterOption: {
          rules: [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: props.detail.vendor,
            },
            {
              field: 'region',
              op: QueryRuleOPEnum.EQ,
              value: props.detail.region,
            },
          ],
          // 属性里传入一个配置，选择是不是要模糊查询
          fuzzySwitch: true,
        },
      },
    });

    const handleBind = async () => {
      for (const item of selections.value) {
        if (selectedSecuirtyGroupsSet.value.has(item.id)) continue;
        selectedSecuirtyGroupsSet.value.add(item.id);
        securityGroups.value.unshift(item);
      }
    };

    const handleUnbind = async (security_group_id: string) => {
      if (selectedSecuirtyGroupsSet.value.has(security_group_id)) {
        const idx = securityGroups.value.findIndex((v) => v.id === security_group_id);
        selectedSecuirtyGroupsSet.value.delete(security_group_id);
        securityGroups.value.splice(idx, 1);
        return;
      }
      await businessStore.unbindSecurityToCLB({
        bk_biz_id: accountStore.bizs,
        security_group_id,
        lb_id: loadBalancerStore.currentSelectedTreeNode.id,
      });
      getBindedSecurityList();
      isSideSliderShow.value = false;
      Message({
        message: '解绑成功',
        theme: 'success',
      });
    };

    const getBindedSecurityList = async () => {
      const res = await businessStore.listCLBSecurityGroups(props.id);
      bindedSecurityGroups.value = cloneDeep(res.data);
      securityGroups.value = res.data;
      for (const item of res.data) {
        bindedSet.add(item.id);
      }
    };

    watch(
      () => props.id,
      () => {
        // 获取已绑定的安全组列表
        getBindedSecurityList();
      },
      {
        immediate: true,
      },
    );

    watch(
      () => props.detail.extension,
      () => {
        // load_balancer_pass_to_target = false, 不放通，检测2次
        isPassToTarget.value = !!props.detail?.extension?.load_balancer_pass_to_target;
        tmpIsPassToTarget.value = isPassToTarget.value;
      },
      {
        deep: true,
        immediate: true,
      },
    );

    watch(
      () => isDialogShow.value,
      () => {
        getListData();
        resetSelections();
      },
    );

    return () => (
      <div>
        <div class={'config-res-wrapper mb24'}>
          {!isPassToTarget.value ? (
            <div>
              <Tag theme='warning' class={'mr16'}>
                2 次检测
              </Tag>
              <span>依次经过负载均衡和RS的安全组 2 次检测</span>
            </div>
          ) : (
            <div>
              <Tag theme='warning' class={'mr16'}>
                1 次检测
              </Tag>
              <span>只经过负载均衡的安全组 1 次检测，忽略后端RS的安全组检测</span>
            </div>
          )}
          <EditLine
            onClick={() => (isConfigDialogShow.value = true)}
            class={'ml12 edit-icon'}
            fill='#3A84FF'
            width={12}
            height={12}
          />
        </div>

        <div class={'line'}></div>

        <div class={'security-rule-container'}>
          <p>
            <span class={'security-rule-container-title'}>绑定安全组</span>
            <span class={'security-rule-container-desc'}>
              当负载均衡不绑定安全组时，其监听端口默认对所有 IP 放通。此处绑定的安全组是直接绑定到负载均衡上面。
            </span>
          </p>
          <div class={'security-rule-container-operations'}>
            <Button theme='primary' class={'mr12'} onClick={() => (isSideSliderShow.value = true)}>
              配置
            </Button>
            {isAllExpand.value ? (
              <Button onClick={() => (isAllExpand.value = false)}>
                <svg
                  width={14}
                  height={14}
                  class='bk-icon'
                  style='fill: #979BA5; margin-right: 8px;'
                  viewBox='0 0 64 64'
                  version='1.1'
                  xmlns='http://www.w3.org/2000/svg'
                >
                  <path
                    fill='#979BA5'
                    d='M56,6H8C6.9,6,6,6.9,6,8v48c0,1.1,0.9,2,2,2h48c1.1,0,2-0.9,2-2V8C58,6.9,57.1,6,56,6z M54,54H10V10	h44V54z'
                  ></path>
                  <path
                    fill='#979BA5'
                    d='M49.6,17.2l-2.8-2.8L38,23.2l0-5.2h-4v12h12v-4h-5.2L49.6,17.2z M38,26L38,26L38,26L38,26z'
                  ></path>
                  <path
                    fill='#979BA5'
                    d='M14.4,46.8l2.8,2.8l8.8-8.8l0,5.2h4V34H18v4h5.2L14.4,46.8z M26,38L26,38L26,38L26,38z'
                  ></path>
                </svg>
                全部收起
              </Button>
            ) : (
              <Button onClick={() => (isAllExpand.value = true)}>
                <svg
                  width={14}
                  height={14}
                  class='bk-icon'
                  style='fill: #979BA5; margin-right: 8px;'
                  viewBox='0 0 64 64'
                  version='1.1'
                  xmlns='http://www.w3.org/2000/svg'
                >
                  <path
                    fill='#979BA5'
                    d='M56,6H8C6.9,6,6,6.9,6,8v48c0,1.1,0.9,2,2,2h48c1.1,0,2-0.9,2-2V8C58,6.9,57.1,6,56,6z M54,54H10V10	h44V54z'
                  ></path>
                  <path
                    fill='#979BA5'
                    d='M34,27.2l2.8,2.8l8.8-8.8v5.2h4v-12h-12v4h5.2L34,27.2z M45.6,18.4L45.6,18.4L45.6,18.4L45.6,18.4z'
                  ></path>
                  <path
                    fill='#979BA5'
                    d='M30,36.8L27.2,34l-8.8,8.8v-5.2h-4v12h12v-4h-5.2L30,36.8z M18.4,45.6L18.4,45.6L18.4,45.6	L18.4,45.6z'
                  ></path>
                </svg>
                全部展开
              </Button>
            )}
            <div class={'security-rule-container-searcher'}>
              <BkRadioGroup v-model={securityRuleType.value} class={'mr12'}>
                <BkRadioButton label={SecurityRuleDirection.in}>入站规则</BkRadioButton>
                <BkRadioButton label={SecurityRuleDirection.out}>出站规则</BkRadioButton>
              </BkRadioGroup>
              <Input class={'search-input'} type='search' clearable v-model={searchVal.value}></Input>
            </div>
          </div>
          <div class={'specific-security-rule-tables'}>
            {securityRulesSearchedResults.value.length ? (
              securityRulesSearchedResults.value.map(({ name, cloud_id, id }, idx) => (
                <ExpandCard
                  name={name}
                  cloudId={cloud_id}
                  idx={idx + 1}
                  isAllExpand={isAllExpand.value}
                  vendor={loadBalancerStore.currentSelectedTreeNode.vendor}
                  direction={securityRuleType.value}
                  id={id}
                />
              ))
            ) : (
              <Exception type='empty' scene='part' description='没有数据'></Exception>
            )}
          </div>
        </div>
        <CommonSideslider
          v-model:isShow={isSideSliderShow.value}
          title='配置安全组'
          width={'640'}
          isSubmitLoading={isSubmitLoading.value}
          handleClose={resetSelections}
          onUpdate:isShow={() => {
            resetSelections();
            selectedSecuirtyGroupsSet.value = new Set();
            securityGroups.value = cloneDeep(bindedSecurityGroups.value);
          }}
          onHandleSubmit={hanldeSubmit}
        >
          {securityGroups.value.length > 5 ? (
            <Alert
              theme='danger'
              title=' 一个负载均衡默认只允许绑定5个安全组，如果特殊需求，请联系腾讯云助手调整'
              class={'mb12'}
            />
          ) : null}
          <div class={'config-security-rule-contianer'}>
            <div class={'config-security-rule-operation'}>
              <BkButtonGroup>
                <Button onClick={() => (isDialogShow.value = true)}>
                  <Plus class={'f22'}></Plus>新增绑定
                </Button>
              </BkButtonGroup>
              <Input class={'search-input'} type='search' clearable v-model={securitySearchVal.value}></Input>
            </div>
            {securitySearchVal.value.trim().length ? (
              securitySearchedList.value.map(({ name, cloud_id, id }, idx) => (
                <DraggableItem
                  securityItem={{ name, cloud_id, id }}
                  idx={idx}
                  securitySearchVal={securitySearchVal.value}
                  handleUnbind={handleUnbind}
                  selectedSecuirtyGroupsSet={selectedSecuirtyGroupsSet.value}
                />
              ))
            ) : (
              <VueDraggable ref={el} v-model={securityGroups.value} animation={200} class={'config-item-wrapper'}>
                {securityGroups.value.length ? (
                  <TransitionGroup type='transition' name='fade'>
                    {securityGroups.value.map(({ name, cloud_id, id }, idx) => (
                      <DraggableItem
                        securityItem={{ name, cloud_id, id }}
                        idx={idx}
                        securitySearchVal={securitySearchVal.value}
                        handleUnbind={handleUnbind}
                        selectedSecuirtyGroupsSet={selectedSecuirtyGroupsSet.value}
                      />
                    ))}
                  </TransitionGroup>
                ) : (
                  <Exception
                    type={securitySearchVal.value.length ? 'search-empty' : 'empty'}
                    description={securitySearchVal.value.length ? '搜索为空' : '暂无绑定'}
                  />
                )}
              </VueDraggable>
            )}
          </div>
        </CommonSideslider>
        <CommonDialog v-model:isShow={isDialogShow.value} title={'绑定安全组'} width={640} onHandleConfirm={handleBind}>
          {{
            default: () => <CommonTable />,
            footer: () => (
              <div>
                <Button
                  theme='primary'
                  class={'mr6'}
                  disabled={securityGroups.value.length + selections.value.length > 5}
                  v-bk-tooltips={{
                    content: '一个负载均衡默认只允许绑定5个安全组，如果特殊需求，请联系腾讯云助手调整',
                    disabled: !(securityGroups.value.length + selections.value.length > 5),
                  }}
                  onClick={() => {
                    handleBind();
                    isDialogShow.value = false;
                  }}
                >
                  确定
                </Button>
                <Button onClick={() => (isDialogShow.value = false)}>取消</Button>
              </div>
            ),
          }}
        </CommonDialog>
        <Dialog
          title='检测配置'
          isShow={isConfigDialogShow.value}
          width={720}
          onClosed={() => (isConfigDialogShow.value = false)}
        >
          {{
            default: () => (
              <div class={'rs-check-selector-container'}>
                <div
                  class={[
                    'rs-check-selector',
                    { 'rs-check-selector-active': !tmpIsPassToTarget.value, 'disabled-button': isUpdating.value },
                  ]}
                  onClick={async () => {
                    if (!tmpIsPassToTarget.value || isUpdating.value) return;
                    tmpIsPassToTarget.value = false;
                  }}
                >
                  <Tag theme='warning'>2 次检测</Tag>
                  <span>依次经过负载均衡和RS的安全组 2 次检测</span>
                  <Success
                    v-show={!tmpIsPassToTarget.value}
                    width={14}
                    height={14}
                    fill='#3A84FF'
                    class={'rs-check-icon'}
                  />
                </div>
                <div
                  class={[
                    'rs-check-selector',
                    { 'rs-check-selector-active': tmpIsPassToTarget.value, 'disabled-button': isUpdating.value },
                  ]}
                  onClick={() => {
                    if (tmpIsPassToTarget.value || isUpdating.value) return;
                    tmpIsPassToTarget.value = true;
                  }}
                >
                  <Tag theme='warning'>1 次检测</Tag>
                  <span>只经过负载均衡的安全组 1 次检测，忽略后端RS的安全组检测</span>
                  <Success
                    v-show={tmpIsPassToTarget.value}
                    width={14}
                    height={14}
                    fill='#3A84FF'
                    class={'rs-check-icon'}
                  />
                </div>
              </div>
            ),
            footer: () => (
              <div>
                <Button
                  loading={isUpdating.value}
                  theme='primary'
                  class={'mr8'}
                  onClick={async () => {
                    isUpdating.value = true;
                    try {
                      await props.updateLb({
                        load_balancer_pass_to_target: tmpIsPassToTarget.value,
                      });
                      isConfigDialogShow.value = false;
                    } finally {
                      isUpdating.value = false;
                    }
                  }}
                >
                  确认
                </Button>
                <Button onClick={() => (isConfigDialogShow.value = false)}>取消</Button>
              </div>
            ),
          }}
        </Dialog>
      </div>
    );
  },
});
