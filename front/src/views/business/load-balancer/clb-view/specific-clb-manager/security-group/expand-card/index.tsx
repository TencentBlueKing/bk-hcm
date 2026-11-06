import { PropType, defineComponent, onMounted, ref, watch } from 'vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { SecurityRuleDirection } from '..';
import { VendorEnum } from '@/common/constant';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT } from '@/constants/menu-symbol';
import QueryString from 'qs';
import http from '@/http';

import { Table } from 'bkui-vue';
import { AngleDown, AngleUp } from 'bkui-vue/lib/icon';
import './index.scss';

export default defineComponent({
  props: {
    idx: { type: Number, required: true },
    name: { type: String, required: true },
    cloudId: { type: String, required: true },
    id: { type: String, required: true },
    isExpand: { type: Boolean, required: true },
    direction: { type: String as PropType<SecurityRuleDirection>, required: true },
    vendor: { type: String as PropType<VendorEnum>, required: true },
  },
  emits: ['expand', 'collapse'],
  setup(props, { emit }) {
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const tableData = ref([]);

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns.filter(
      ({ field }: { field: string }) => !['updated_at', 'memo'].includes(field),
    );

    const getSecurityRules = async () => {
      if (!props.direction) return;
      const res = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}vendors/${props.vendor}/security_groups/${props.id}/rules/list`,
        {
          filter: { op: 'and', rules: [{ field: 'type', op: 'eq', value: props.direction }] },
          page: { count: false, start: 0, limit: 500 },
        },
      );
      tableData.value = res.data?.details ?? [];
      // 展开状态由父级集合统一维护，子组件只上报结果，避免父子互相回写
      emit(tableData.value.length > 0 ? 'expand' : 'collapse');
    };

    onMounted(() => {
      getSecurityRules();
    });

    watch(() => props.direction, getSecurityRules);

    const handleToggleExpand = () => {
      emit(props.isExpand ? 'collapse' : 'expand');
    };

    const openSecurityGroupManagementPage = (e: Event) => {
      e.stopPropagation();
      routerAction.open({
        name: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
        params: { bizId: getBizsId() },
        query: {
          filter: QueryString.stringify(
            { cloud_id: props.cloudId },
            { arrayFormat: 'comma', encode: false, allowEmptyArrays: true },
          ),
        },
      });
    };

    return () => (
      <div class={'security-rule-table-container'}>
        <div class={'security-rule-table-header'} onClick={handleToggleExpand}>
          {props.isExpand ? <AngleUp width={22} height={22} /> : <AngleDown width={22} height={22} />}
          <div class={'config-security-item-idx'}>{props.idx + 1}</div>
          <span class={'config-security-item-name'}>{props.name}</span>
          <span class={'config-security-item-id'}>({props.cloudId})</span>
          <bk-button class='config-security-item-btn' theme='primary' text onClick={openSecurityGroupManagementPage}>
            查看更多
            <i class='hcm-icon bkhcm-icon-jump-fill ml4'></i>
          </bk-button>
        </div>
        {props.isExpand && (
          <Table stripe data={tableData.value} columns={securityRulesColumns}>
            {{
              empty: () => (
                <bk-exception
                  class='exception-wrap-item exception-part'
                  type='empty'
                  scene='part'
                  description='无规则，默认拒绝所有流量'
                />
              ),
            }}
          </Table>
        )}
      </div>
    );
  },
});
