import { Button, Table } from 'bkui-vue';
import { PropType, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { AngleDown, AngleUp } from 'bkui-vue/lib/icon';

import http from '@/http';
import { VendorEnum } from '@/common/constant';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { SecurityRuleDirection } from '..';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    idx: {
      type: Number,
      required: true,
    },
    name: {
      type: String,
      required: true,
    },
    cloudId: {
      type: String,
      required: true,
    },
    id: {
      type: String,
      required: true,
    },
    isAllExpand: {
      type: Boolean,
      required: true,
    },
    direction: {
      type: String as PropType<SecurityRuleDirection>,
      required: true,
    },
    vendor: {
      type: String as PropType<VendorEnum>,
      required: true,
    },
  },
  setup(props) {
    const isExpand = ref(true);
    const tableData = ref([]);
    watch(
      () => props.isAllExpand,
      (isAllExpand) => {
        isExpand.value = isAllExpand;
      },
    );

    const getSecurityRules = async () => {
      if (!props.direction) return;
      const res = await http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${props.vendor}/security_groups/${props.id}/rules/list`,
        {
          filter: {
            op: 'and',
            rules: [
              {
                field: 'type',
                op: 'eq',
                value: props.direction,
              },
            ],
          },
          page: { count: false, start: 0, limit: 500 },
        },
      );
      tableData.value = res.data.details;
      if (!tableData.value) isExpand.value = false;
    };

    watch(
      () => props.direction,
      () => {
        getSecurityRules();
      },
      {
        immediate: true,
      },
    );

    const securityRulesColumns = useColumns('securityCommon', false, props.vendor).columns.filter(
      ({ field }: { field: string }) => !['updated_at', 'memo'].includes(field),
    );

    return () => (
      <div>
        <div class={'security-rule-table-container'}>
          <div class={'security-rule-table-header'}>
            <div onClick={() => (isExpand.value = !isExpand.value)} class={'header-icon'}>
              {isExpand.value ? <AngleUp width={34} height={28} /> : <AngleDown width={34} height={28} />}
            </div>
            <div class={'config-security-item-idx'}>{props.idx}</div>
            <span class={'config-security-item-name'}>{props.name}</span>
            <span class={'config-security-item-id'}>({props.cloudId})</span>
            <div class={'config-security-item-btn'}>
              <Button
                theme='primary'
                text
                onClick={() => {
                  const url = `/#/business/security?cloud_id=${props.cloudId}`;
                  window.open(url, '_blank');
                }}>
                查看更多
              </Button>
              <span class='icon hcm-icon bkhcm-icon-jump-fill ml5'></span>
            </div>
          </div>
          {isExpand.value ? (
            <div class={'security-rule-table-panel'}>
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
            </div>
          ) : null}
        </div>
      </div>
    );
  },
});
