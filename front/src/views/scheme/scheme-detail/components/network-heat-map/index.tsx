import { defineComponent, reactive, ref } from "vue";
import { BkTable, BkTableColumn } from "bkui-vue/lib/table";
import SearchInput from "../../../components/search-input/index";

import './index.scss';

export default defineComponent({
  name: 'network-heat-map',
  setup () {
    const searchStr = ref('');
    const isHighlight = ref(false);
    const idcList = reactive([
      { id: 'New York', name: '纽约机房' },
      { id: 'Silicon Valley', name: '硅谷机房' },
      { id: 'Frank Furt', name: '法兰克福机房' },
      { id: 'Brooklyn', name: '布鲁克林机房' },
      { id: 'Singapore', name: '新加坡机房' },
    ]);
    const idcData = reactive([]);

    const handleSearch = () => {};
    return () => (
      <div class="network-heat-map">
        <h3 class="title">网络热力分析</h3>
        <div class="data-switch-panel">
          <div class="data-type-tabs">
            <div class="tab-item actived">业务数据</div>
            <div class="tab-item">裸 ping 数据</div>
          </div>
          <SearchInput v-model={searchStr.value} width={300} onSearch={handleSearch} />
        </div>
        <div class="data-content-wrapper">
          <div class="network-data">
            <div class="data-item">
              <span class="label">网络平均延迟: </span>
              <span class="value">54ms</span>
            </div>
            <div class="data-item">
              <span class="label">平均丢包率: </span>
              <span class="value">3%</span>
            </div>
            <div class="data-item">
              <span class="label">平均 ping 抖动: </span>
              <span class="value">11ms</span>
            </div>
          </div>
          <bk-checkbox v-model={isHighlight.value}>高亮服务区域</bk-checkbox>
        </div>
        <div class="idc-data-table">
          <BkTable border={['outer']} data={idcData}>
            <BkTableColumn label="国家 \ IDC"></BkTableColumn>
            {
              idcList.map(idc => {
                return (<BkTableColumn key={idc.id} label={idc.name}></BkTableColumn>)
              })
            }
          </BkTable>
        </div>
      </div>
    );
  },
});
