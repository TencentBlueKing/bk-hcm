import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Button, Card, Checkbox, Dialog, Loading, Table } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import './security-group-selector.scss';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    multiple: Boolean as PropType<boolean>,
    vendor: String as PropType<string>,
    vpcId: String as PropType<string>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const list = ref([]);
    const loading = ref(false);
    const { isServicePage } = useWhereAmI();
    const isDialogShow = ref(false);
    const isScrollLoading = ref(false);
    const securityGroupRulesMap = ref(new Map());
    const isRulesTableLoading = ref(false);

    const computedDisabled = computed(() => {
      return !(props.accountId && props.vendor && props.region);
    });

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns;

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

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

    return () => (
      // <Select
      //   filterable={true}
      //   modelValue={selected.value}
      //   onUpdate:modelValue={val => selected.value = val}
      //   multiple={props.multiple}
      //   loading={loading.value}
      //   class={isSelected.value && 'security-group-cls'}
      //   {...{ attrs }}
      // >
      //   {
      //     list.value.map(({ cloud_id, name }) => (
      //       <Option key={cloud_id} value={cloud_id} label={name}></Option>
      //     ))
      //   }
      // </Select>
      <div class={'selected-block-container'}>
        {selected.value?.length ? (
          <div class={'selected-block mr8'}>{selected.value}</div>
        ) : null}
        {selected.value?.length ? (
          <EditLine
            fill='#3A84FF'
            width={13.5}
            height={13.5}
            onClick={() => (isDialogShow.value = true)}
          />
        ) : (
          <Button
            onClick={() => (isDialogShow.value = true)}
            disabled={computedDisabled.value}>
            <Plus class='f20' />
            选择安全组
          </Button>
        )}
        <Dialog
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={() => {
            // selected.value = checkedImageId.value;
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
                          isRulesTableLoading.value = false;
                        } else {
                          securityGroupRulesMap.value.delete(data.cloud_id);
                        }
                      }}>
                        { cell }
                      </Checkbox>
                    ),
                  }]}
                />
              </Loading>
            </div>
            <div></div>
            <div>
              <Loading loading={isRulesTableLoading.value}>
                <div>
                  {
                    Array.from(securityGroupRulesMap.value).map(([key, value]) => <div>
                      {key}
                      <Card
                        isCollapse
                        collapseStatus={true}
                      >
                        <Table
                          data={value}
                          columns={securityRulesColumns}
                        />
                      </Card>
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
