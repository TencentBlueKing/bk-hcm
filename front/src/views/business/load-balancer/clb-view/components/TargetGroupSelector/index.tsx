import { PropType, defineComponent, ref, watch } from 'vue';
import { Divider, Select } from 'bkui-vue';
import { Plus, RightTurnLine, Spinner } from 'bkui-vue/lib/icon';
import { useAccountStore } from '@/store';
import { useSingleList } from '@/hooks/useSingleList';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { Protocol, QueryRuleOPEnum } from '@/typings';

const { Option } = Select;

/**
 * 目标组选择器
 * @prop modelValue 目标组 ID
 * @prop accountId 负载均衡的云账户 ID
 * @prop cloudVpcId 负载均衡的云VPC ID
 * @prop region 负载均衡的地域
 * @prop protocol 协议 (TCP, UDP, HTTP, HTTPS)
 */
export default defineComponent({
  name: 'TargetGroupSelector',
  props: {
    modelValue: String,
    accountId: String,
    cloudVpcId: String,
    region: String,
    protocol: String as PropType<Protocol>,
    isCorsV2: Boolean,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const { getBusinessApiPath } = useWhereAmI();
    const accountStore = useAccountStore();

    const targetGroupId = ref(props.modelValue);
    const { dataList, isDataLoad, isDataRefresh, handleScrollEnd, handleRefresh } = useSingleList({
      url: `/api/v1/cloud/${getBusinessApiPath()}target_groups/list`,
      rules: () => {
        const baseRules = [
          { field: 'account_id', op: QueryRuleOPEnum.EQ, value: props.accountId },
          { field: 'protocol', op: QueryRuleOPEnum.EQ, value: props.protocol },
        ];
        if (props.isCorsV2) return baseRules;
        return [
          ...baseRules,
          { field: 'region', op: QueryRuleOPEnum.EQ, value: props.region },
          { field: 'cloud_vpc_id', op: QueryRuleOPEnum.EQ, value: props.cloudVpcId },
        ];
      },
    });

    // click-handler - 新增目标组
    const handleAddTargetGroup = () => {
      const url = `/#/business/loadbalancer/group-view?bizs=${accountStore.bizs}`;
      window.open(url, '_blank');
    };

    // select-handler - 选择目标组后更新 props.modelValue
    watch(targetGroupId, (val) => emit('update:modelValue', val));

    // 对外暴露刷新列表的方法
    expose({ handleRefresh });

    return () => (
      <Select
        class='target-group-selector'
        v-model={targetGroupId.value}
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {{
          default: () =>
            dataList.value.map(({ id, name, listener_num }) => (
              <Option key={id} id={id} name={name} disabled={listener_num > 0} />
            )),
          extension: () => (
            <div style='width: 100%; color: #63656E; padding: 0 12px;'>
              <div style='display: flex; align-items: center;justify-content: center;'>
                <span style='display: flex; align-items: center;cursor: pointer;' onClick={handleAddTargetGroup}>
                  <Plus style='font-size: 20px;' />
                  新增
                </span>
                <span style='display: flex; align-items: center;position: absolute; right: 12px;'>
                  <Divider direction='vertical' type='solid' />
                  {isDataRefresh.value ? (
                    <Spinner style='font-size: 14px;color: #3A84FF;' />
                  ) : (
                    <RightTurnLine style='font-size: 14px;cursor: pointer;' onClick={handleRefresh} />
                  )}
                </span>
              </div>
            </div>
          ),
        }}
      </Select>
    );
  },
});
