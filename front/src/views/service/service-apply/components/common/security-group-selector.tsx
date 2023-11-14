import http from '@/http';
import { computed, defineComponent, PropType, ref, TransitionGroup, watch } from 'vue';
import { Button, Checkbox, Dialog, Input, Loading, Table } from 'bkui-vue';
import { SECURITY_GROUP_RULE_TYPE, VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import './security-group-selector.scss';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import DraggableCard from './DraggableCard';
import { type UseDraggableReturn, VueDraggable } from 'vue-draggable-plus';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { QueryRuleOPEnum } from '@/typings';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

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

    const computedDisabled = computed(() => {
      return !(props.accountId && props.vendor && props.region);
    });

    const computedSecurityGroupRules = computed(() => {
      return securityGroupRules.value.map(({ id, data }) => ({
        id,
        data: data.filter(({ type }: any) => type === selectedSecurityType.value),
      }));
    });

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns.filter(({ field }: {field: string}) => !['updated_at'].includes(field));
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
      [
        () => props.bizId,
        () => props.accountId,
        () => props.region,
        () => props.vpcId,
        () => searchVal.value,
      ],
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
        const result = await http.post(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/security_groups/list`,
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

    watch(
      () => securityGroupRules.value,
      arr => console.log(111, arr.map(({ id }) => securityGroupKVMap.value.get(id))),
      {
        deep: true,
      },
    );

    return () => (
      <div>
        {selected.value?.length ? (
         <div class={'image-selector-selected-block-container'}>
           <div class={'selected-block mr8'}>
            {
              selected.value.map(val => (
                <>
                  {securityGroupKVMap.value.get(val)}<br/>
                </>
              ))
            }
          </div>
          <EditLine
            fill='#3A84FF'
            width={13.5}
            height={13.5}
            onClick={() => (isDialogShow.value = true)}
          />
         </div>
        ) : <div/>}
        {selected.value?.length ? (
          null
        ) : (
          <Button
            onClick={() => (isDialogShow.value = true)}
            disabled={computedDisabled.value}>
            <Plus class='f20' />
            选择安全组
          </Button>
        )}
        <div>
          {
            list.value.length || computedDisabled.value
              ? null
              : (
                <div class={'security-selector-tips'}>
                  无可用的安全组,可 <Button theme='primary' text onClick={() => {
                  const url = '/#/resource/resource?type=security';
                  window.open(url, '_blank');
                }}>新建安全组</Button>
                </div>
              )
          }
        </div>
        <Dialog
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={() => {
            // selected.value = [...Array.from(securityGroupKVMap.value).map(([k, _val]) => k)];
            selected.value = securityGroupRules.value.map(({ id }) => id);
            props.onSelectedChange(selected.value);
            isDialogShow.value = false;
          }}
          title='选择安全组'
          width={1500}>
          <div class={'security-container'}>
            <div class={'fixed-security-list'}>
              <Input
                class={'search-input'}
                placeholder='搜索安全组'
                type='search'
                clearable
                v-model={searchVal.value}/>
              <Loading loading={loading.value}>
                <div>
                  {
                    list.value.length
                      ? list.value.map(item => (
                      <div class={'security-search-item'}>
                        <Checkbox
                          label={'data.cloud_id'}
                          onChange={async (isSelected: boolean) => {
                            if (isSelected) {
                              isRulesTableLoading.value = true;
                              const res = await http.post(
                                `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${props.vendor}/security_groups/${item.id}/rules/list`,
                                {
                                  filter: {
                                    op: 'and',
                                    rules: [],
                                  },
                                  page: {
                                    count: false,
                                    start: 0,
                                    limit: 500,
                                  },
                                },
                              );
                              const arr = res.data?.details || [];
                              securityGroupRules.value.push({
                                id: item.cloud_id,
                                data: arr,
                              });
                              securityGroupKVMap.value.set(
                                item.cloud_id,
                                item.name,
                              );
                              isRulesTableLoading.value = false;
                            } else {
                              securityGroupRules.value = securityGroupRules.value.filter(({
                                id,
                              }) => id !== item.cloud_id);
                              securityGroupKVMap.value.delete(item.cloud_id);
                            }
                          }}>
                          {item.name}
                        </Checkbox>
                      </div>
                      ))
                      : (
                      <bk-exception
                        class="exception-wrap-item exception-part"
                        type="search-empty"
                        scene="part"
                        description="搜索为空"
                      />
                      )
                  }

                </div>
              </Loading>
            </div>
            <div class={'security-group-rules-container'}></div>
            <div>
              <Loading loading={isRulesTableLoading.value}>
                <div class={'security-group-rules-container'}>
                  <div class={'security-group-rules-btn-group-container'}>
                    <BkButtonGroup class={'security-group-rules-btn-group'}>
                      <Button
                        selected={selectedSecurityType.value === SECURITY_GROUP_RULE_TYPE.EGRESS}
                        onClick={() => selectedSecurityType.value = SECURITY_GROUP_RULE_TYPE.EGRESS}
                      >
                        出站规则
                      </Button>
                      <Button
                        selected={selectedSecurityType.value === SECURITY_GROUP_RULE_TYPE.INGRESS}
                        onClick={() => selectedSecurityType.value = SECURITY_GROUP_RULE_TYPE.INGRESS}
                      >
                        入站规则
                      </Button>
                    </BkButtonGroup>
                    <Button
                      class={'security-group-rules-expand-btn'}
                      onClick={() => isAllExpand.value = !isAllExpand.value}
                    >
                      {
                        isAllExpand.value
                          ? '全部收起'
                          : '全部展开'
                      }
                    </Button>
                  </div>
                  {/* @ts-ignore */}
                  <VueDraggable
                    ref={el}
                    v-model={securityGroupRules.value}
                    animation={200}
                    handle='.draggable-card-header-draggable-btn'
                  >
                    {computedSecurityGroupRules.value.length ? (
                      <TransitionGroup type='transition' name='fade'>
                        {computedSecurityGroupRules.value.map(({ id, data }, idx) => (
                            <DraggableCard
                              key={idx}
                              title={securityGroupKVMap.value.get(id)}
                              index={idx + 1}
                              isAllExpand={isAllExpand.value}>
                              <Table
                                data={data}
                                columns={securityRulesColumns}
                              />
                            </DraggableCard>
                        ))}
                      </TransitionGroup>
                    ) : (
                      <bk-exception
                        class="exception-wrap-item exception-part"
                        type="empty"
                        scene="part"
                        description="没有数据"
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
