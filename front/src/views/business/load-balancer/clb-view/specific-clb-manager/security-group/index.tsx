import { PropType, TransitionGroup, computed, defineComponent, reactive, ref, watch, watchEffect } from 'vue';
import { useBusinessStore } from '@/store';
import { ILoadBalancerDetails, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { useTable } from '@/hooks/useTable/useTable';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { QueryRuleOPEnum } from '@/typings';
import { LOAD_BALANCER_PASS_TO_TARGET_LIST } from '@/constants';
import { cloneDeep } from 'lodash';
import { IAuthSign } from '@/common/auth-service';

import { Message } from 'bkui-vue';
import { EditLine, Plus, Success } from 'bkui-vue/lib/icon';
import { VueDraggable } from 'vue-draggable-plus';
import ExpandCard from './expand-card';
import DraggableItem from './draggable-item';
import ModalFooter from '@/components/modal/modal-footer.vue';
import './index.scss';

export enum SecurityRuleDirection {
  in = 'ingress',
  out = 'egress',
}

export default defineComponent({
  props: {
    detail: Object as PropType<ILoadBalancerDetails>,
    getDetails: Function,
    id: String,
    currentGlobalBusinessId: Number,
    clbOperationAuthSign: Object as PropType<IAuthSign | IAuthSign[]>,
  },
  setup(props) {
    const businessStore = useBusinessStore();
    const loadBalancerClbStore = useLoadBalancerClbStore();

    // 检查并转义正则特殊字符
    const escapeRegExp = (str: string) => {
      return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    };

    // 安全组放通模式
    const isPassToTarget = ref(false);
    const tmpIsPassToTarget = ref(isPassToTarget.value);
    const isPassToTargetConfigDialogState = reactive({ isShow: false, isLoading: false });
    const securityGroupPassToTargetRender = () => {
      const target = LOAD_BALANCER_PASS_TO_TARGET_LIST.find((item) => item.value === isPassToTarget.value);

      return (
        <div class='pass-to-target-config-display-container'>
          <div>
            <bk-tag theme='warning'>{target.label}</bk-tag>
            <span class='ml16'>{target.description}</span>
          </div>
          <hcm-auth class={'mr5'} sign={props.clbOperationAuthSign}>
            {{
              default: ({ noPerm }: { noPerm: boolean }) => (
                <bk-button
                  theme='primary'
                  text
                  disabled={noPerm}
                  onClick={() => (isPassToTargetConfigDialogState.isShow = true)}>
                  <EditLine class='edit-icon' width={12} height={12} />
                </bk-button>
              ),
            }}
          </hcm-auth>
        </div>
      );
    };
    const handleConfirmPassToTargetConfig = async () => {
      isPassToTargetConfigDialogState.isLoading = true;
      try {
        await loadBalancerClbStore.updateLoadBalancer(
          props.detail.vendor,
          { id: props.id, load_balancer_pass_to_target: tmpIsPassToTarget.value },
          props.currentGlobalBusinessId,
        );
        props.getDetails(props.id);
        isPassToTargetConfigDialogState.isShow = false;
      } finally {
        isPassToTargetConfigDialogState.isLoading = false;
      }
    };
    watch(
      () => props.detail.extension,
      () => {
        // load_balancer_pass_to_target = false, 不放通，检测2次
        isPassToTarget.value = !!props.detail?.extension?.load_balancer_pass_to_target;
        tmpIsPassToTarget.value = isPassToTarget.value;
      },
      { deep: true, immediate: true },
    );

    // 已绑定的安全组
    const expandIdSet = ref<Set<string>>(new Set());
    const activeRuleType = ref(SecurityRuleDirection.in);
    const securityGroups = ref([]);
    const searchValue = ref('');
    const bindedSecurityGroups = ref([]);
    const isAllExpand = ref(false);
    watch(
      () => expandIdSet.value.size,
      (size) => (isAllExpand.value = size === securityGroups.value.length),
    );

    const displaySecurityGroupRules = computed(() => {
      const val = searchValue.value;
      if (!val.trim()) return bindedSecurityGroups.value;
      const reg = new RegExp(escapeRegExp(val));
      return bindedSecurityGroups.value.filter((v) => reg.test(`${v.name} (${v.cloud_id})`));
    });

    const getBindedSecurityList = async () => {
      const res = await businessStore.listCLBSecurityGroups(props.id);
      bindedSecurityGroups.value = cloneDeep(res.data);
      securityGroups.value = res.data;
    };
    watch(() => props.id, getBindedSecurityList, { immediate: true });

    // 安全组配置弹框
    const securityGroupConfigModalState = reactive({
      sidesliderVisible: false,
      sidesliderLoading: false,
      sidesliderSearchValue: '',
      displaySecurityGroupList: [],
      dialogVisible: false,
      dialogLoading: false,
    });
    const selectedSecurityGroupsSet = ref(new Set());

    watchEffect(() => {
      const val = securityGroupConfigModalState.sidesliderSearchValue;
      if (!val.trim()) {
        securityGroupConfigModalState.displaySecurityGroupList = [];
        return;
      }
      const reg = new RegExp(escapeRegExp(val));
      securityGroupConfigModalState.displaySecurityGroupList = securityGroups.value.filter((v) =>
        reg.test(`${v.name} (${v.cloud_id})`),
      );
    });

    const handleUnbind = async (security_group_id: string) => {
      if (selectedSecurityGroupsSet.value.has(security_group_id)) {
        const idx = securityGroups.value.findIndex((v) => v.id === security_group_id);
        selectedSecurityGroupsSet.value.delete(security_group_id);
        securityGroups.value.splice(idx, 1);
        return;
      }
      await businessStore.unbindSecurityToCLB({
        bk_biz_id: props.currentGlobalBusinessId,
        security_group_id,
        lb_id: props.id,
      });
      Message({ message: '解绑成功', theme: 'success' });

      // 移除已绑定的安全组
      bindedSecurityGroups.value.splice(
        bindedSecurityGroups.value.findIndex((v) => v.id === security_group_id),
        1,
      );
      securityGroups.value.splice(
        securityGroups.value.findIndex((v) => v.id === security_group_id),
        1,
      );
    };

    const handleSubmitSecurityGroupConfig = async () => {
      try {
        securityGroupConfigModalState.sidesliderLoading = true;
        await businessStore.bindSecurityToCLB({
          bk_biz_id: props.currentGlobalBusinessId,
          lb_id: props.id,
          security_group_ids: securityGroups.value.map(({ id }) => id),
        });
        Message({ message: '绑定成功', theme: 'success' });
        securityGroupConfigModalState.sidesliderVisible = false;
        getBindedSecurityList();
      } finally {
        securityGroupConfigModalState.sidesliderLoading = false;
      }
    };

    const tableColumns = [
      { type: 'selection', width: 30, minWidth: 30 },
      { label: '安全组名称', field: 'name' },
      { label: 'ID', field: 'cloud_id' },
      { label: '备注', field: 'memo' },
    ];
    const searchData: ISearchItem[] = [
      { id: 'name', name: '安全组名称' },
      { id: 'cloud_id', name: 'ID' },
    ];
    const isRowSelectEnable = ({ row, isCheckAll }: any) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => {
      return (
        !bindedSecurityGroups.value.map((v) => v.id).includes(row.id) && !selectedSecurityGroupsSet.value.has(row.id)
      );
    };
    const { CommonTable, getListData } = useTable({
      searchOptions: { searchData, extra: { searchSelectExtStyle: { width: '100%' } } },
      tableOptions: {
        columns: tableColumns,
        extra: {
          maxHeight: '50vh',
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
            { field: 'vendor', op: QueryRuleOPEnum.EQ, value: props.detail.vendor },
            { field: 'region', op: QueryRuleOPEnum.EQ, value: props.detail.region },
          ],
          // 属性里传入一个配置，选择是不是要模糊查询
          fuzzySwitch: true,
        },
      },
    });
    const { selections, handleSelectionChange, resetSelections } = useSelection();

    const handleBind = async () => {
      for (const item of selections.value) {
        if (selectedSecurityGroupsSet.value.has(item.id)) continue;
        selectedSecurityGroupsSet.value.add(item.id);
        securityGroups.value.unshift(item);
      }
      securityGroupConfigModalState.dialogVisible = false;
    };

    const handleShowSecurityGroupConfigSideslider = () => {
      selectedSecurityGroupsSet.value = new Set();
      securityGroups.value = cloneDeep(bindedSecurityGroups.value);
      securityGroupConfigModalState.sidesliderVisible = true;
    };

    const handleShowSecurityGroupConfigDialog = () => {
      securityGroupConfigModalState.dialogVisible = true;
      resetSelections();
      getListData();
    };

    return () => (
      <div class='clb-security-group-container'>
        {securityGroupPassToTargetRender()}

        <div class={'line'}></div>

        <div class={'security-group-rule-container'}>
          <div class='header'>
            <span class='title'>绑定安全组</span>
            <span class='description'>
              当负载均衡不绑定安全组时，其监听端口默认对所有 IP 放通。此处绑定的安全组是直接绑定到负载均衡上面。
            </span>
          </div>
          <div class='toolbar'>
            <hcm-auth class='mr12' sign={props.clbOperationAuthSign}>
              {{
                default: ({ noPerm }: { noPerm: boolean }) => (
                  <bk-button theme='primary' disabled={noPerm} onClick={handleShowSecurityGroupConfigSideslider}>
                    配置
                  </bk-button>
                ),
              }}
            </hcm-auth>
            <bk-button onClick={() => (isAllExpand.value = !isAllExpand.value)}>
              <i
                class={[
                  'hcm-icon',
                  { 'bkhcm-icon-zoomout': isAllExpand.value, 'bkhcm-icon-fullscreen': !isAllExpand.value },
                ]}
              />
              {isAllExpand.value ? '全部收起' : '全部展开'}
            </bk-button>
            <div class='search'>
              <bk-radio-group v-model={activeRuleType.value}>
                <bk-radio-button label={SecurityRuleDirection.in}>入站规则</bk-radio-button>
                <bk-radio-button label={SecurityRuleDirection.out}>出站规则</bk-radio-button>
              </bk-radio-group>
              <bk-input v-model={searchValue.value} type='search' clearable class={'search-input'} />
            </div>
          </div>
          <div class='rules-display-container'>
            {displaySecurityGroupRules.value.length ? (
              displaySecurityGroupRules.value.map(({ name, cloud_id, id }, idx) => (
                <ExpandCard
                  key={cloud_id}
                  name={name}
                  cloudId={cloud_id}
                  idx={idx}
                  isAllExpand={isAllExpand.value}
                  vendor={props.detail.vendor}
                  direction={activeRuleType.value}
                  id={id}
                  onExpand={() => expandIdSet.value.add(cloud_id)}
                  onCollapse={() => expandIdSet.value.delete(cloud_id)}
                />
              ))
            ) : (
              <bk-exception type='empty' scene='part' description='没有数据' />
            )}
          </div>
        </div>

        <bk-dialog
          class='pass-to-target-config-dialog'
          title='检测配置'
          isShow={isPassToTargetConfigDialogState.isShow}
          width={960}
          onClosed={() => (isPassToTargetConfigDialogState.isShow = false)}>
          {{
            default: () => (
              <div class={'rs-check-selector-container'}>
                {LOAD_BALANCER_PASS_TO_TARGET_LIST.map(({ label, description, value }) => {
                  const active = value === tmpIsPassToTarget.value;
                  const disabled = isPassToTargetConfigDialogState.isLoading;
                  const handleClick = () => {
                    if (disabled) return;
                    tmpIsPassToTarget.value = value;
                  };
                  return (
                    <div
                      class={['rs-check-selector', { 'rs-check-selector-active': active, 'disabled-button': disabled }]}
                      onClick={handleClick}>
                      <span class='label-tag'>
                        <bk-tag theme='warning'>{label}</bk-tag>
                      </span>
                      <span>{description}</span>
                      <Success v-show={active} width={14} height={14} fill='#3A84FF' class={'rs-check-icon'} />
                    </div>
                  );
                })}
              </div>
            ),
            footer: () => (
              <ModalFooter
                loading={isPassToTargetConfigDialogState.isLoading}
                onConfirm={handleConfirmPassToTargetConfig}
                onClosed={() => (isPassToTargetConfigDialogState.isShow = false)}
              />
            ),
          }}
        </bk-dialog>

        <bk-sideslider
          class='security-group-config-sideslider'
          v-model:isShow={securityGroupConfigModalState.sidesliderVisible}
          title='配置安全组'
          width={640}>
          {{
            default: () => (
              <>
                {securityGroups.value.length > 5 && (
                  <bk-alert
                    theme='danger'
                    title=' 一个负载均衡默认只允许绑定5个安全组，如果特殊需求，请联系腾讯云助手调整'
                    class={'mb12'}
                  />
                )}
                <div class='toolbar'>
                  <bk-button onClick={handleShowSecurityGroupConfigDialog}>
                    <Plus class='f22'></Plus>新增绑定
                  </bk-button>
                  <bk-input
                    v-model={securityGroupConfigModalState.sidesliderSearchValue}
                    class={'search-input'}
                    type='search'
                    clearable
                  />
                </div>
                {securityGroupConfigModalState.sidesliderSearchValue.trim().length ? (
                  securityGroupConfigModalState.displaySecurityGroupList.map(({ name, cloud_id, id }, idx) => (
                    <DraggableItem
                      securityItem={{ name, cloud_id, id }}
                      idx={idx}
                      securitySearchVal={securityGroupConfigModalState.sidesliderSearchValue}
                      handleUnbind={handleUnbind}
                      selectedSecuirtyGroupsSet={selectedSecurityGroupsSet.value}
                    />
                  ))
                ) : (
                  <VueDraggable v-model={securityGroups.value} animation={200} class='rules-display-container'>
                    {securityGroups.value.length ? (
                      <TransitionGroup type='transition' name='fade'>
                        {securityGroups.value.map(({ name, cloud_id, id }, idx) => (
                          <DraggableItem
                            securityItem={{ name, cloud_id, id }}
                            idx={idx}
                            securitySearchVal={securityGroupConfigModalState.sidesliderSearchValue}
                            handleUnbind={handleUnbind}
                            selectedSecuirtyGroupsSet={selectedSecurityGroupsSet.value}
                          />
                        ))}
                      </TransitionGroup>
                    ) : (
                      <bk-exception
                        type={securityGroupConfigModalState.sidesliderSearchValue.length ? 'search-empty' : 'empty'}
                        description={
                          securityGroupConfigModalState.sidesliderSearchValue.length ? '搜索为空' : '暂无绑定'
                        }
                      />
                    )}
                  </VueDraggable>
                )}
              </>
            ),
            footer: () => (
              <ModalFooter
                loading={securityGroupConfigModalState.sidesliderLoading}
                disabled={securityGroups.value.length === 0}
                onConfirm={handleSubmitSecurityGroupConfig}
                onClosed={() => resetSelections()}
              />
            ),
          }}
        </bk-sideslider>
        <bk-dialog v-model:isShow={securityGroupConfigModalState.dialogVisible} title={'绑定安全组'} width={640}>
          {{
            default: () => <CommonTable />,
            footer: () => (
              <ModalFooter
                disabled={securityGroups.value.length + selections.value.length > 5}
                tooltips={{
                  content: '一个负载均衡默认只允许绑定5个安全组，如果特殊需求，请联系腾讯云助手调整',
                  disabled: !(securityGroups.value.length + selections.value.length > 5),
                }}
                onConfirm={handleBind}
                onClosed={() => (securityGroupConfigModalState.dialogVisible = false)}
              />
            ),
          }}
        </bk-dialog>
      </div>
    );
  },
});
