import { defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { Button, Exception, InfoBox, Message, Table, Tag } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import SearchInput from '@/views/scheme/components/search-input';
import CommonSideslider from '@/components/common-sideslider';
import CommonDialog from '@/components/common-dialog';
import { useAccountStore, useBusinessStore } from '@/store';
import { Plus } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useTable/useTable';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useLoadBalancerStore } from '@/store/loadbalancer';

export default defineComponent({
  setup() {
    const rsCheckRes = ref('a');
    const securityRuleType = ref('in');
    const isSideSliderShow = ref(false);
    const hanldeSubmit = () => {};
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();
    const { selections, handleSelectionChange } = useSelection();
    const securityGroups = ref([]);
    const isDialogShow = ref(false);
    const bindedSet = reactive(new Set());
    const loadBalancerStore = useLoadBalancerStore();

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

    const { CommonTable } = useTable({
      searchOptions: {
        searchData,
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          onSelectionChange: (selections: any) => handleSelectionChange(selections, () => true),
          onSelectAll: (selections: any) => handleSelectionChange(selections, () => true, true),
        },
      },
      requestOption: {
        type: 'security_groups',
      },
    });

    const handleBind = async () => {
      await businessStore.bindSecurityToCLB({
        bk_biz_id: accountStore.bizs,
        lb_id: loadBalancerStore.lb.id,
        security_group_ids: selections.value.map(({ id }) => id),
      });
      getBindedSecurityList();
      Message({
        message: '绑定成功',
        theme: 'success',
      });
    };

    const handleUnbind = async (security_group_id: string) => {
      await businessStore.unbindSecurityToCLB({
        bk_biz_id: accountStore.bizs,
        security_group_id,
        lb_id: loadBalancerStore.lb.id,
      });
      getBindedSecurityList();
      Message({
        message: '解绑成功',
        theme: 'success',
      });
    };

    const getBindedSecurityList = async () => {
      const res = await businessStore.listCLBSecurityGroups(loadBalancerStore.lb.id);
      securityGroups.value = res.data;
      for (const item of res.data) {
        bindedSet.add(item.id);
      }
    };

    watch(
      () => loadBalancerStore.lb.id,
      async () => {
        getBindedSecurityList();
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div>
        <div class={'rs-check-selector-container'}>
          <div
            class={`${rsCheckRes.value === 'a' ? 'rs-check-selector-active' : 'rs-check-selector'}`}
            onClick={() => {
              rsCheckRes.value = 'a';
            }}>
            <Tag theme='warning'>2 次检测</Tag>
            <span>依次经过负载均衡和RS的安全组 2 次检测</span>
          </div>
          <div
            class={`${rsCheckRes.value === 'b' ? 'rs-check-selector-active' : 'rs-check-selector'}`}
            onClick={() => {
              rsCheckRes.value = 'b';
            }}>
            <Tag theme='warning'>1 次检测</Tag>
            <span>只经过负载均衡的安全组 1 次检测，忽略后端RS的安全组检测</span>
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
            <Button>全部收起</Button>
            <div class={'security-rule-container-searcher'}>
              <BkButtonGroup class={'mr12'}>
                <Button
                  selected={securityRuleType.value === 'in'}
                  onClick={() => {
                    securityRuleType.value = 'in';
                  }}>
                  出站规则
                </Button>
                <Button
                  selected={securityRuleType.value === 'out'}
                  onClick={() => {
                    securityRuleType.value = 'out';
                  }}>
                  入站规则
                </Button>
              </BkButtonGroup>
              <SearchInput placeholder='请输入' />
            </div>
          </div>
          <div class={'specific-security-rule-tables'}>
            {securityGroups.value.length ? (
              securityGroups.value.map(({ name, cloud_id }, idx) => (
                <div class={'security-rule-table-container'}>
                  <div class={'security-rule-table-header'}>
                    <div class={'config-security-item-idx'}>{idx}</div>
                    <span class={'config-security-item-name'}>{name}</span>
                    <span class={'config-security-item-id'}>({cloud_id})</span>
                  </div>
                  <div class={'security-rule-table-panel'}>
                    <Table
                      stripe
                      data={[
                        {
                          target: 'any',
                          source: 'Any',
                          portProtocol: 'abnc',
                          policy: '允许',
                        },
                        {
                          target: '1abc',
                          source: '1ads',
                          portProtocol: 'abc',
                          policy: '拒绝',
                        },
                        {
                          target: 'abc',
                          source: 'abc',
                          portProtocol: 'Tabv3',
                          policy: '允许',
                        },
                      ]}
                      columns={[
                        {
                          label: '目标',
                          field: 'target',
                        },
                        {
                          label: '来源',
                          field: 'source',
                        },
                        {
                          label: '端口协议',
                          field: 'portProtocol',
                        },
                        {
                          label: '策略',
                          field: 'policy',
                        },
                      ]}
                    />
                  </div>
                </div>
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
          onHandleSubmit={hanldeSubmit}>
          <div class={'config-security-rule-contianer'}>
            <div class={'config-security-rule-operation'}>
              <BkButtonGroup>
                <Button onClick={() => (isDialogShow.value = true)}>
                  <Plus class={'f22'}></Plus>新增绑定
                </Button>
              </BkButtonGroup>
              <SearchInput class={'operation-search-input'} />
            </div>
            <div>
              {securityGroups.value.map(({ name, cloud_id, id }, idx) => (
                <div class={'config-security-item'}>
                  <i class={'hcm-icon bkhcm-icon-grag-fill mr16 draggable-card-header-draggable-btn'}></i>
                  <div class={'config-security-item-idx'}>{idx}</div>
                  <span class={'config-security-item-name'}>{name}</span>
                  <span class={'config-security-item-id'}>({cloud_id})</span>
                  <div class={'config-security-item-edit-block'}>
                    <Button text theme='primary' class={'mr27'}>
                      去编辑
                      <span class='icon hcm-icon bkhcm-icon-jump-fill ml5'></span>
                    </Button>
                    <Button
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
                      解绑
                    </Button>
                  </div>
                </div>
              ))}
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
