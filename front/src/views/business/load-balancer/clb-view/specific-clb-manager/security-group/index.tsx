import { PropType, computed, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { Button, Exception, InfoBox, Input, Message, Tag } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import SearchInput from '@/views/scheme/components/search-input';
import CommonSideslider from '@/components/common-sideslider';
import CommonDialog from '@/components/common-dialog';
import { useAccountStore, useBusinessStore } from '@/store';
import { Plus, Success } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useTable/useTable';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useLoadBalancerStore } from '@/store/loadbalancer';
import ExpandCard from './expand-card';
import { QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { IDetail } from '@/hooks/useRouteLinkBtn';

export const SecuirtyRuleDirection = {
  in: 'ingress',
  out: 'egress',
};

export default defineComponent({
  props: {
    detail: Object as PropType<IDetail>,
    getDetails: Function,
    updateLb: Function,
  },
  setup(props) {
    const rsCheckRes = ref(false);
    const securityRuleType = ref(SecuirtyRuleDirection.in);
    const isSideSliderShow = ref(false);
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();
    const { selections, handleSelectionChange } = useSelection();
    const isAllExpand = ref(true);
    const securitySearchVal = ref('');
    const selectedSecuirtyGroups = ref([]);
    const bindedSecurityGroups = ref([]);
    const securityGroups = computed(() => {
      const groups = [].concat(selectedSecuirtyGroups.value).concat(bindedSecurityGroups.value);
      return groups;
    });
    const isDialogShow = ref(false);
    const bindedSet = reactive(new Set());
    const loadBalancerStore = useLoadBalancerStore();
    const hanldeSubmit = async () => {
      await businessStore.bindSecurityToCLB({
        bk_biz_id: accountStore.bizs,
        lb_id: loadBalancerStore.currentSelectedTreeNode.id,
        security_group_ids: selections.value.map(({ id }) => id),
      });
      getBindedSecurityList();
      selectedSecuirtyGroups.value = [];
      isSideSliderShow.value = false;
      Message({
        message: '绑定成功',
        theme: 'success',
      });
    };

    // 检查并转义正则特殊字符
    const escapeRegExp = (str: string) => {
      return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    };

    // 高亮命中关键词
    const getHighLightNameText = (name: string, rootCls: string) => {
      return (
        <div
          class={rootCls}
          v-html={name?.replace(
            new RegExp(securitySearchVal.value, 'g'),
            `<span class='search-result-highlight'>${securitySearchVal.value}</span>`,
          )}></div>
      );
    };

    const securitySearchedList = computed(() => {
      const val = securitySearchVal.value;
      if (!val.trim()) return securityGroups.value;
      const reg = new RegExp(escapeRegExp(val));
      return securityGroups.value.filter((v) => reg.test(`${v.name} (${v.cloud_id})`));
    });
    const tableColumns = [
      {
        type: 'selection',
        width: 32,
        minWidth: 32,
      },
      {
        label: '安全组名称',
        field: 'name',
      },
      {
        label: 'ID',
        field: 'id',
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
        id: 'id',
        name: 'ID',
      },
    ];

    const isRowSelectEnable = ({ row, isCheckAll }: any) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };

    const isCurRowSelectEnable = (row: any) => {
      return !bindedSecurityGroups.value.map((v) => v.id).includes(row.id);
    };

    const { CommonTable } = useTable({
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
          onSelectionChange: (selections: any) => handleSelectionChange(selections, () => true),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          isRowSelectEnable,
          isSelectedFn: ({ row }: any) => {
            return selectedSecuirtyGroups.value.map((v) => v.id).includes(row.id);
          },
        },
      },
      requestOption: {
        type: 'security_groups',
        filterOption: {
          rules: [
            {
              field: 'vendor',
              op: QueryRuleOPEnum.EQ,
              value: VendorEnum.TCLOUD,
            },
            {
              field: 'region',
              op: QueryRuleOPEnum.EQ,
              value: loadBalancerStore.currentSelectedTreeNode.region,
            },
          ],
        },
      },
    });

    const handleBind = async () => {
      const arr = selections.value;
      selectedSecuirtyGroups.value = arr;
    };

    const handleUnbind = async (security_group_id: string) => {
      if (selectedSecuirtyGroups.value.map((v) => v.id).includes(security_group_id)) {
        const idx = selectedSecuirtyGroups.value.findIndex((v) => v.id === security_group_id);
        selectedSecuirtyGroups.value.splice(idx, 1);
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
      const res = await businessStore.listCLBSecurityGroups(loadBalancerStore.currentSelectedTreeNode.id);
      bindedSecurityGroups.value = res.data;
      for (const item of res.data) {
        bindedSet.add(item.id);
      }
    };

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { id, type } = val;
        if (type === 'lb' && id) getBindedSecurityList();
      },
      {
        immediate: true,
      },
    );

    watch(
      () => props.detail?.extension?.load_balancer_pass_to_target,
      (isPass) => {
        rsCheckRes.value = !!isPass;
      },
    );

    return () => (
      <div>
        <div class={'rs-check-selector-container'}>
          <div
            class={`${rsCheckRes.value ? 'rs-check-selector-active' : 'rs-check-selector'}`}
            onClick={() => {
              rsCheckRes.value = true;
              props.updateLb({
                load_balancer_pass_to_target: true,
              });
            }}>
            <Tag theme='warning'>2 次检测</Tag>
            <span>依次经过负载均衡和RS的安全组 2 次检测</span>
            <Success
              width={14}
              height={14}
              fill='#3A84FF'
              style={{ visibility: !rsCheckRes.value ? 'hidden' : 'visible' }}
              class={'rs-check-icon'}
            />
          </div>
          <div
            class={`${!rsCheckRes.value ? 'rs-check-selector-active' : 'rs-check-selector'}`}
            onClick={() => {
              rsCheckRes.value = false;
              props.updateLb({
                load_balancer_pass_to_target: false,
              });
            }}>
            <Tag theme='warning'>1 次检测</Tag>
            <span>只经过负载均衡的安全组 1 次检测，忽略后端RS的安全组检测</span>
            <Success
              width={14}
              height={14}
              fill='#3A84FF'
              style={{ visibility: rsCheckRes.value ? 'hidden' : 'visible' }}
              class={'rs-check-icon'}
            />
          </div>
        </div>
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
                全部展开
              </Button>
            )}
            <div class={'security-rule-container-searcher'}>
              <BkButtonGroup class={'mr12'}>
                <Button
                  selected={securityRuleType.value === SecuirtyRuleDirection.in}
                  onClick={() => {
                    securityRuleType.value = SecuirtyRuleDirection.in;
                  }}>
                  出站规则
                </Button>
                <Button
                  selected={securityRuleType.value === SecuirtyRuleDirection.out}
                  onClick={() => {
                    securityRuleType.value = SecuirtyRuleDirection.out;
                  }}>
                  入站规则
                </Button>
              </BkButtonGroup>
              <SearchInput placeholder='请输入' />
            </div>
          </div>
          <div class={'specific-security-rule-tables'}>
            {bindedSecurityGroups.value.length ? (
              bindedSecurityGroups.value.map(({ name, cloud_id, id }, idx) => (
                <ExpandCard
                  name={name}
                  cloudId={cloud_id}
                  idx={idx}
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
          isSubmitDisabled={!selectedSecuirtyGroups.value.length}
          onHandleSubmit={hanldeSubmit}>
          <div class={'config-security-rule-contianer'}>
            <div class={'config-security-rule-operation'}>
              <BkButtonGroup>
                <Button onClick={() => (isDialogShow.value = true)}>
                  <Plus class={'f22'}></Plus>新增绑定
                </Button>
              </BkButtonGroup>
              <Input class={'search-input'} type='search' clearable v-model={securitySearchVal.value}></Input>
            </div>
            <div class={'config-item-wrapper'}>
              {securitySearchedList.value.length ? (
                securitySearchedList.value.map(({ name, cloud_id, id }, idx) => (
                  <div
                    class={
                      selectedSecuirtyGroups.value.map((v) => v.id).includes(id)
                        ? 'config-security-item-new'
                        : 'config-security-item'
                    }>
                    <i class={'hcm-icon bkhcm-icon-grag-fill mr8 draggable-card-header-draggable-btn'}></i>
                    <div class={'config-security-item-idx'}>{idx}</div>
                    <span class={'config-security-item-name'}>
                      {securitySearchVal.value ? getHighLightNameText(name, '') : name}
                    </span>
                    <span class={'config-security-item-id'}>({cloud_id})</span>
                    <div class={'config-security-item-edit-block'}>
                      <Button
                        text
                        theme='primary'
                        class={'mr27'}
                        onClick={() => {
                          const url = `/#/business/security?cloud_id=${cloud_id}`;
                          window.open(url, '_blank');
                        }}>
                        去编辑
                        <span class='icon hcm-icon bkhcm-icon-jump-fill ml5'></span>
                      </Button>
                      <Button
                        class={'mr24'}
                        text
                        theme='danger'
                        onClick={() => {
                          InfoBox({
                            infoType: 'warning',
                            title: '是否确定解绑当前安全组',
                            onConfirm() {
                              handleUnbind(id);
                            },
                          });
                        }}>
                        <svg
                          viewBox='0 0 1024 1024'
                          width={11.45}
                          height={11.45}
                          class={'mr4'}
                          fill='#EA3636'
                          version='1.1'
                          xmlns='http://www.w3.org/2000/svg'>
                          <path
                            fill-rule='evenodd'
                            d='M286.2275905828571 466.8617596342857L346.4517251657142 527.0656921599999 195.92332068571426 677.5959793371429 346.4297720685714 828.1262665142856 496.9801303771428 677.5959793371429 557.1823111314285 737.7989134628571 376.5528159085714 918.4526189714285C368.56767780571425 926.4409673142857 357.73563904 930.9290422857142 346.44074861714284 930.9290422857142 335.1458581942857 930.9290422857142 324.3138194285714 926.4409673142857 316.3286813257143 918.4526189714285L105.57614218971428 707.6974460342857C88.95576212479999 691.0712107885714 88.95576212479999 664.1207471542857 105.57614218971428 647.4945119085714L286.2275905828571 466.8617596342857ZM271.2404106971428 210.97583908571426L813.0869855085714 752.8720991085713 752.8848040228571 813.0750332342856 211.0162761142857 271.17877321142856 271.2404106971428 210.97583908571426ZM392.31718473142854 571.8582944914285L452.17451739428566 631.7174696228572 362.39992393142853 721.47449344 302.5654023314285 661.6361472 392.31718473142854 571.8582944914285ZM707.7107156114284 105.57629805714285L918.463253942857 316.3314731885714C935.0843026285713 332.9578225371429 935.0843026285713 359.9090556342857 918.463253942857 376.53540498285713L737.8138024228571 557.1891126857142 677.6106232685714 496.9642247314285 828.1390299428571 346.4548937142857 677.6106232685714 195.90265270857142 527.0812203885713 346.4349359542857 466.85708580571423 286.23100342857146 647.5085348571429 105.57729594514285C664.1345616457143 88.95571016411428 691.0846888228571 88.95571016411428 707.7107156114284 105.57629805714285ZM661.6181525942857 302.5654023314285L721.47449344 362.4037493028571 631.7008925257143 452.18160128 571.8653791085713 392.32242614857137 661.6181525942857 302.5654023314285Z'
                          />
                        </svg>
                        解绑
                      </Button>
                    </div>
                  </div>
                ))
              ) : (
                <Exception
                  type={securitySearchVal.value.length ? 'search-empty' : 'empty'}
                  description={securitySearchVal.value.length ? '搜索为空' : '暂无绑定'}
                />
              )}
            </div>
          </div>
        </CommonSideslider>
        <CommonDialog v-model:isShow={isDialogShow.value} title={'绑定安全组'} width={640} onHandleConfirm={handleBind}>
          <CommonTable />
        </CommonDialog>
      </div>
    );
  },
});
