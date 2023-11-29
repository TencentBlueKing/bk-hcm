import { defineComponent, reactive, ref, PropType, onMounted } from "vue";
import { BkTable, BkTableColumn } from "bkui-vue/lib/table";
import { IAreaInfo } from '@/typings/scheme';
import { useSchemeStore } from "@/store";
import SearchInput from "../../../components/search-input/index";

import './index.scss';

export default defineComponent({
  name: 'network-heat-map',
  props: {
    ids: Array as PropType<string[]>,
    areaTopo: Array as PropType<IAreaInfo[]>,
  },
  setup (props) {
    const schemeStore = useSchemeStore();

    const searchStr = ref('');
    const isHighlight = ref(false);
    const idcList = reactive([]);
    const idcData = reactive([]);
    const activedTab = ref('biz');

    const TABS = [
      { id: 'biz', label: '业务数据' },
      { id: 'ping', label: '裸 ping 数据' },
    ];

    const handleSearch = () => {};

    const getTableData = async() => {
      let res;
      if (activedTab.value === 'biz') {
        res = await schemeStore.queryBizLatency(props.areaTopo, props.ids);
      } else {
        res = await schemeStore.queryPingLatency(props.areaTopo, props.ids);
      }
      console.log(res);
    };

    const handleSwitchTab = (id: string) => {
      activedTab.value = id;
      getTableData();
    }

    onMounted(() => {
      getTableData();
    });

    return () => (
      <div class="network-heat-map">
        <h3 class="title">网络热力分析</h3>
        <div class="data-switch-panel">
          <div class="data-type-tabs">
            {
              TABS.map(tab => {
                return (
                  <div class={['tab-item', activedTab.value === tab.id ? 'actived' : '']} onClick={() => { handleSwitchTab(tab.id) }}>{tab.label}</div>
                )
              })
            }
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
