import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Button, Checkbox, Dialog, Loading, Table } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import './security-group-selector.scss';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import DraggableCard from './DraggableCard';

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
    const isScrollLoading = ref(false);
    const securityGroupRulesMap = ref(new Map());
    const securityGroupKVMap = ref(new Map<string, string>());
    const isRulesTableLoading = ref(false);

    const computedDisabled = computed(() => {
      return !(props.accountId && props.vendor && props.region);
    });

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns;

    const selected = ref([]);

    // const isSelected = computed(() => {
    //   if (selected.value) {
    //     return !!Object.keys(selected.value).length;
    //   }
    //   return false;
    // });

    const handleScrollBottom = () => {
      isScrollLoading.value = true;
    };

    watch(
      [
        () => props.bizId,
        () => props.accountId,
        () => props.region,
        () => props.vpcId,
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
          securityGroupRulesMap.value = new Map();
        }
      },
    );

    return () => (
      <div>
        {selected.value?.length ? (
         <div class={'selected-block-container'}>
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
                <div>
                  无可用的安全组,可 <Button theme='primary' text onClick={() => {
                  const url = '/#/service/service-apply/cvm';
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
            selected.value = [...Array.from(securityGroupKVMap.value).map(([k, _val]) => k)];
            props.onSelectedChange(selected.value);
            isDialogShow.value = false;
          }}
          title='选择安全组'
          width={1500}>
          <div class={'security-container'}>
            <div class={'fixed-security-list'}>
              <Loading loading={loading.value}>
                <Table
                  border={'none'}
                  data={list.value}
                  scrollLoading={isScrollLoading.value}
                  onScrollBottom={handleScrollBottom}
                  columns={[{
                    field: 'name',
                    label: '',
                    render: ({ cell, data }: any) => (
                      <Checkbox label={data.cloud_id} onChange={async (isSelected: boolean) => {
                        if (isSelected) {
                          isRulesTableLoading.value = true;
                          const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${props.vendor}/security_groups/${data.id}/rules/list`, {
                            filter: {
                              op: 'and',
                              rules: [],
                            },
                            page: {
                              count: false,
                              start: 0,
                              limit: 500,
                            },
                          });
                          const arr = res.data?.details || [];
                          securityGroupRulesMap.value.set(data.cloud_id, arr);
                          securityGroupKVMap.value.set(data.cloud_id, data.name);
                          isRulesTableLoading.value = false;
                        } else {
                          securityGroupRulesMap.value.delete(data.cloud_id);
                          securityGroupKVMap.value.delete(data.cloud_id);
                        }
                      }}>
                        { cell }
                      </Checkbox>
                    ),
                  }]}
                />
              </Loading>
            </div>
            <div class={'security-group-rules-container'}></div>
            <div>
              <Loading loading={isRulesTableLoading.value}>
                <div class={'security-group-rules-container'}>
                  {
                    Array.from(securityGroupRulesMap.value).map(([key, value]) => <div>
                      {/* <Card
                        isCollapse
                        collapseStatus={false}
                        title={securityGroupKVMap.value.get(key)}
                        class={'mb12'}
                      >
                        <Table
                          data={value}
                          columns={securityRulesColumns}
                        />
                      </Card> */}
                      <DraggableCard
                        title={securityGroupKVMap.value.get(key)}
                        index={1}
                      >
                        <Table
                            data={value}
                            columns={securityRulesColumns}
                          />
                      </DraggableCard>
                    </div>)
                  }
                </div>
              </Loading>
            </div>
          </div>
        </Dialog>
      </div>
    );
  },
});
