import http from '@/http';
import { computed, defineComponent, PropType, ref, TransitionGroup, watch } from 'vue';
import { Button, Checkbox, Dialog, Input, Loading, Table } from 'bkui-vue';
import { SECURITY_GROUP_RULE_TYPE, VendorEnum } from '@/common/constant';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import './security-group-selector.scss';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import DraggableCard from './DraggableCard';
import { type UseDraggableReturn, VueDraggable } from 'vue-draggable-plus';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { QueryRuleOPEnum } from '@/typings';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useResourceStore } from '@/store';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    modelValue: String as PropType<string | string[]>,
    bizId: Number as PropType<number | string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    multiple: Boolean as PropType<boolean>,
    vendor: String as PropType<string>,
    vpcId: String as PropType<string>,
    onSelectedChange: Function as PropType<(val: string[]) => void>,
  },
  emits: ['update:modelValue'],
  setup(props) {
    const resourceStore = useResourceStore();
    const list = ref([]);
    const loading = ref(false);
    const { isServicePage } = useWhereAmI();
    const isDialogShow = ref(false);
    // const isScrollLoading = ref(false);
    const securityGroupRules = ref([]);
    const securityGroupKVMap = ref(new Map<string, string>());
    const isRulesTableLoading = ref(false);
    const el = ref<UseDraggableReturn>();
    const selectedSecurityType = ref(SECURITY_GROUP_RULE_TYPE.INGRESS);
    const { whereAmI } = useWhereAmI();

    const computedDisabled = computed(() => {
      return !(props.accountId && props.vendor && props.region);
    });

    const computedSecurityGroupRules = computed(() => {
      return securityGroupRules.value.map(({ id, data }) => ({
        id,
        data: data.filter(({ type }: any) => type === selectedSecurityType.value),
      }));
    });

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns.filter(
      ({ field }: { field: string }) => !['updated_at'].includes(field),
    );
    // const securityRulesColumns = [
    //   {
    //     label: '目标',
    //     field: 'target_ip',
    //     render: ({ data }: any) => {
    //       return data.ipv4_cidr || data.ipv6_cidr || '--';
    //     },
    //   },
    //   {
    //     label: '端口协议',
    //     field: 'protocol_port',
    //     render: ({ data }: any) => `${data.protocol}:${data.port}`,
    //   },
    //   {
    //     label: '策略',
    //     field: 'action',
    //     render: ({ data }: any) => `${data.action || data.access || '--'}`,
    //   },
    // ];

    const selected = ref([]);
    const searchVal = ref('');
    const isAllExpand = ref(true);

    // const isSelected = computed(() => {
    //   if (selected.value) {
    //     return !!Object.keys(selected.value).length;
    //   }
    //   return false;
    // });

    // const handleScrollBottom = () => {
    //   isScrollLoading.value = true;
    // };

    watch(
      [() => props.bizId, () => props.accountId, () => props.region, () => props.vpcId, () => searchVal.value],
      async ([bizId, accountId, region, vpcId]) => {
        if ((!bizId && isServicePage) || !accountId || !region) {
          list.value = [];
          return;
        }
        loading.value = true;
        // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${bizId}/security_groups/list`, {
        const rules = [
          {
            field: 'account_id',
            op: 'eq',
            value: accountId,
          },
          {
            field: 'region',
            op: 'eq',
            value: region,
          },
        ];
        if (searchVal.value.length) {
          rules.push({
            field: 'name',
            op: QueryRuleOPEnum.CS,
            value: searchVal.value,
          });
        }
        if (props.vendor === VendorEnum.AWS) {
          rules.push({
            field: 'extension.vpc_id',
            op: 'json_eq',
            value: vpcId,
          });
        }
        const result = await resourceStore.getCommonList(
          {
            filter: {
              op: 'and',
              rules,
            },
            page: {
              count: false,
              start: 0,
              limit: 500,
            },
          },
          'security_groups/list',
        );
        list.value = result?.data?.details ?? [];
        loading.value = false;
      },
    );

    watch(
      () => isDialogShow.value,
      (isShow) => {
        if (!isShow) {
          searchVal.value = '';
          securityGroupRules.value = [];
        }
      },
    );

    const currentIndex = ref(-1);
    const handleSecurityGroupChange = async (isSelected: boolean, item: any, index: number) => {
      if (isSelected) {
        currentIndex.value = index;
        isRulesTableLoading.value = true;
        const res = await http.post(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${props.vendor}/security_groups/${item.id}/rules/list`,
          {
            filter: { op: 'and', rules: [] },
            page: { count: false, start: 0, limit: 500 },
          },
        );
        const arr = res.data?.details || [];
        securityGroupRules.value.push({ id: item.cloud_id, data: arr });
        securityGroupKVMap.value.set(item.cloud_id, item.name);
        isRulesTableLoading.value = false;
      } else {
        securityGroupRules.value = securityGroupRules.value.filter(({ id }) => id !== item.cloud_id);
        securityGroupKVMap.value.delete(item.cloud_id);
      }
    };

    return () => (
      <div>
        {selected.value?.length ? (
          <div class={'image-selector-selected-block-container'}>
            <div class={'selected-block mr8'}>
              {selected.value.map((val) => (
                <>
                  {securityGroupKVMap.value.get(val)}
                  <br />
                </>
              ))}
            </div>
            <EditLine fill='#3A84FF' width={13.5} height={13.5} onClick={() => (isDialogShow.value = true)} />
          </div>
        ) : (
          <div />
        )}
        {selected.value?.length ? null : (
          <Button
            onClick={() => (isDialogShow.value = true)}
            disabled={computedDisabled.value || list.value.length === 0}>
            <Plus class='f20' />
            选择安全组
          </Button>
        )}
        <div>
          {list.value.length || computedDisabled.value ? null : (
            <div class={'security-selector-tips'}>
              无可用的安全组，可{' '}
              <Button
                theme='primary'
                text
                onClick={() => {
                  const url =
                    whereAmI.value === Senarios.business
                      ? '/#/business/security'
                      : '/#/resource/resource?type=security';
                  window.open(url, '_blank');
                }}>
                新建安全组
              </Button>
            </div>
          )}
        </div>
        <Dialog
          class={'security-dialog-wrap'}
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={() => {
            // selected.value = [...Array.from(securityGroupKVMap.value).map(([k, _val]) => k)];
            selected.value = securityGroupRules.value.map(({ id }) => id);
            props.onSelectedChange(selected.value);
            // props.onSelectedChange(props.vendor === VendorEnum.AZURE ? selected.value?.[0] : selected.value);
            isDialogShow.value = false;
          }}
          title='选择安全组'
          width={'60vw'}
          height={'80vh'}>
          <div class={'security-container'}>
            <div class={'security-list g-scroller'}>
              <Input
                class={'search-input'}
                placeholder='搜索安全组'
                type='search'
                clearable
                v-model={searchVal.value}
              />
              <Loading loading={loading.value} class={'mt8'}>
                <div class={'security-search-list'}>
                  {list.value.length ? (
                    list.value.map((item, index) => (
                      <div class={'security-search-item'}>
                        <Checkbox
                          disabled={
                            props.vendor === VendorEnum.AZURE &&
                            securityGroupRules.value.length > 0 &&
                            currentIndex.value !== index
                          }
                          label={'data.cloud_id'}
                          onChange={(isSelected: boolean) => handleSecurityGroupChange(isSelected, item, index)}>
                          {item.name}
                        </Checkbox>
                      </div>
                    ))
                  ) : (
                    <bk-exception
                      class='exception-wrap-item exception-part'
                      type='search-empty'
                      scene='part'
                      description='搜索为空'
                    />
                  )}
                </div>
              </Loading>
            </div>
            {/* <div class={'security-group-rules-container'}></div>*/}
            <div class={'security-group-rules-wrap'}>
              <Loading loading={isRulesTableLoading.value}>
                <div class={'security-group-rules-container'}>
                  <div class={'security-group-rules-btn-group-container'}>
                    <BkButtonGroup>
                      <Button
                        style={{ marginLeft: '1px' }}
                        selected={selectedSecurityType.value === SECURITY_GROUP_RULE_TYPE.EGRESS}
                        onClick={() => (selectedSecurityType.value = SECURITY_GROUP_RULE_TYPE.EGRESS)}>
                        出站规则
                      </Button>
                      <Button
                        selected={selectedSecurityType.value === SECURITY_GROUP_RULE_TYPE.INGRESS}
                        onClick={() => (selectedSecurityType.value = SECURITY_GROUP_RULE_TYPE.INGRESS)}>
                        入站规则
                      </Button>
                    </BkButtonGroup>
                    <Button style={{ marginRight: '1px' }} onClick={() => (isAllExpand.value = !isAllExpand.value)}>
                      {isAllExpand.value ? (
                        <>
                          <svg
                            width={14}
                            height={14}
                            class='bk-icon'
                            style='fill: #979BA5; margin-right: 8px;'
                            viewBox='0 0 64 64'
                            version='1.1'
                            xmlns='http://www.w3.org/2000/svg'>
                            <path
                              fill='#979BA5'
                              d='M56,6H8C6.9,6,6,6.9,6,8v48c0,1.1,0.9,2,2,2h48c1.1,0,2-0.9,2-2V8C58,6.9,57.1,6,56,6z M54,54H10V10	h44V54z'></path>
                            <path
                              fill='#979BA5'
                              d='M49.6,17.2l-2.8-2.8L38,23.2l0-5.2h-4v12h12v-4h-5.2L49.6,17.2z M38,26L38,26L38,26L38,26z'></path>
                            <path
                              fill='#979BA5'
                              d='M14.4,46.8l2.8,2.8l8.8-8.8l0,5.2h4V34H18v4h5.2L14.4,46.8z M26,38L26,38L26,38L26,38z'></path>
                          </svg>
                          <span>全部收起</span>
                        </>
                      ) : (
                        <>
                          <svg
                            width={14}
                            height={14}
                            class='bk-icon'
                            style='fill: #979BA5; margin-right: 8px;'
                            viewBox='0 0 64 64'
                            version='1.1'
                            xmlns='http://www.w3.org/2000/svg'>
                            <path
                              fill='#979BA5'
                              d='M56,6H8C6.9,6,6,6.9,6,8v48c0,1.1,0.9,2,2,2h48c1.1,0,2-0.9,2-2V8C58,6.9,57.1,6,56,6z M54,54H10V10	h44V54z'></path>
                            <path
                              fill='#979BA5'
                              d='M34,27.2l2.8,2.8l8.8-8.8v5.2h4v-12h-12v4h5.2L34,27.2z M45.6,18.4L45.6,18.4L45.6,18.4L45.6,18.4z'></path>
                            <path
                              fill='#979BA5'
                              d='M30,36.8L27.2,34l-8.8,8.8v-5.2h-4v12h12v-4h-5.2L30,36.8z M18.4,45.6L18.4,45.6L18.4,45.6	L18.4,45.6z'></path>
                          </svg>
                          <span>全部展开</span>
                        </>
                      )}
                    </Button>
                  </div>
                  {/* @ts-ignore */}
                  <VueDraggable
                    ref={el}
                    v-model={securityGroupRules.value}
                    animation={200}
                    handle='.draggable-card-header-draggable-btn'
                    class={'security-group-rules-list g-scroller'}>
                    {computedSecurityGroupRules.value.length ? (
                      <TransitionGroup type='transition' name='fade'>
                        {computedSecurityGroupRules.value.map(({ id, data }, idx) => (
                          <DraggableCard
                            key={idx}
                            title={securityGroupKVMap.value.get(id)}
                            index={idx + 1}
                            isAllExpand={isAllExpand.value}>
                            <Table data={data} columns={securityRulesColumns} showOverflowTooltip stripe={true} />
                          </DraggableCard>
                        ))}
                      </TransitionGroup>
                    ) : (
                      <bk-exception
                        class='exception-wrap-item exception-part'
                        type='empty'
                        scene='part'
                        description='没有数据'
                      />
                    )}
                  </VueDraggable>
                </div>
              </Loading>
            </div>
          </div>
        </Dialog>
      </div>
    );
  },
});
