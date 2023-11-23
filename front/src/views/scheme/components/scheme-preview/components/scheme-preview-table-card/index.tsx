import { defineComponent, ref } from 'vue';
import './index.scss';
import { Table, Tag, Loading, Button } from 'bkui-vue';
import { AngleDown, AngleRight } from 'bkui-vue/lib/icon';
import VendorTcloud from '@/assets/image/vendor-tcloud.png';

export default defineComponent({
  setup() {
    const columns = [
      {
        field: 'name',
        label: '部署点名称',
      },
      {
        field: 'vendor',
        label: '云厂商',
      },
      {
        field: 'abc',
        label: '所在地',
      },
      {
        field: 'edg',
        label: '服务区域',
      },
      {
        field: 'ping',
        label: '平均延迟',
      },
    ];
    const tableData = ref([]);
    const isLoading = ref(false);
    const isExpanded = ref(true);
    return () => (
      <div class={'scheme-preview-table-card-container'}>
        <div
          class={`scheme-preview-table-card-header ${
            isExpanded.value ? '' : 'scheme-preview-table-card-header-closed'
          }`}>
          {isExpanded.value ? (
            <AngleDown
              width={'40px'}
              height={'30px'}
              fill='#63656E'
              onClick={() => (isExpanded.value = !isExpanded.value)}
              class={'scheme-preview-table-card-header-expand-icon'}
            />
          ) : (
            <AngleRight
              width={'40px'}
              height={'30px'}
              fill='#63656E'
              onClick={() => (isExpanded.value = !isExpanded.value)}
              class={'scheme-preview-table-card-header-expand-icon'}
            />
          )}

          <p class={'scheme-preview-table-card-header-title'}>方案一</p>
          <Tag
            theme='success'
            radius='11px'
            class={'scheme-preview-table-card-header-tag'}>
            集中式部署
          </Tag>
          <img
            src={VendorTcloud}
            class={'scheme-preview-table-card-header-icon'}
          />
          <div class={'scheme-preview-table-card-header-score'}>
            <div class={'scheme-preview-table-card-header-score-item'}>
              综合评分： <span class={'score-value'}>81</span>
            </div>
            <div class={'scheme-preview-table-card-header-score-item'}>
              网络评分： <span class={'score-value'}>81</span>
            </div>
            <div class={'scheme-preview-table-card-header-score-item'}>
              方案成本： <span class={'score-value'}>$ 300</span>
            </div>
          </div>
          <div class={'scheme-preview-table-card-header-operation'}>
            <Button class={'mr8'}>查看详情</Button>
            <Button theme='primary'>保存</Button>
          </div>
        </div>
        <div
          class={`scheme-preview-table-card-panel ${
            isExpanded.value ? '' : 'scheme-preview-table-card-panel-invisable'
          }`}>
          <Loading loading={isLoading.value}>
            <Table data={tableData.value} columns={columns} />
          </Loading>
        </div>
      </div>
    );
  },
});
