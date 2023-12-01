import { defineComponent, reactive, ref, PropType, onMounted, getTransitionRawChildren } from "vue";
import { BkTable, BkTableColumn } from "bkui-vue/lib/table";
import { IAreaInfo } from '@/typings/scheme';
import { IIdcListItem, IIdcLatencyListItem } from "@/typings/scheme";
import { useSchemeStore } from "@/store";
import { getScoreColor } from '@/common/util';
import SearchInput from "../../../components/search-input/index";

import './index.scss';

interface ITableDataItem {
  [key: string]: string|number|boolean;
}


export default defineComponent({
  name: 'network-heat-map',
  props: {
    idcList: Array as PropType<IIdcListItem[]>,
    areaTopo: Array as PropType<IAreaInfo[]>,
  },
  setup (props) {
    const schemeStore = useSchemeStore();

    const searchStr = ref('');
    const isHighlight = ref(false);
    let idcData = reactive<ITableDataItem[]>([]);
    const idcDataLoading = ref(false);
    const activedTab = ref('biz');

    const TABS = [
      { id: 'biz', label: '业务数据' },
      { id: 'ping', label: '裸 ping 数据' },
    ];

    const handleSearch = () => {};

    const getTableData = async() => {
      try {
        idcDataLoading.value = true;
        let res;
        const idcs = props.idcList.map(item => item.id);
        if (activedTab.value === 'biz') {
          res = await schemeStore.queryBizLatency(props.areaTopo, idcs);
        } else {
          res = await schemeStore.queryPingLatency(props.areaTopo, idcs);
        }
        console.log(res);
        idcData = transToTableData(res.data.details);
        idcDataLoading.value = false;
      } catch (e) {
        console.log(e);
      }
    };

    const handleSwitchTab = (id: string) => {
      activedTab.value = id;
      getTableData();
    }

    const transToTableData = (data: IIdcLatencyListItem[]) => {
      const list: ITableDataItem[] = [];
      data.forEach(country => {
        const totalValue = {};
        const regionDataList: ITableDataItem[] = [];
        country.children.forEach(item => {
          const { name, value } = item;
          regionDataList.push({ rowName: name, ...value });
          Object.keys(value).forEach(key => {
            const val = value[key];
            if (!(key in totalValue)) {
              totalValue[key] = val;
            } else {
              totalValue[key] += val;
            }
          });
        });
        if (country.children.length > 0) {
          const averagedValue = Object.keys(totalValue).reduce((obj: { [key: string]: number|string }, key: string) => {
            obj[key] = (totalValue[key] / country.children.length).toFixed(2);
            return obj;
          }, {});
          list.push({ rowName: country.name, isCountry: true, ...averagedValue });
          list.push(...regionDataList);
        }
      });
      return list;
    };

    const renderValCell = (cell: string|number, data: ITableDataItem) => {
      if (data) {
        return (<div class="value-cell" style={{ color: getScoreColor(Number(cell)) }}>{`${cell}ms`}</div>);
      }
    };

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
          <BkTable key={activedTab.value} border={['col', 'outer']} data={idcData}>
            <BkTableColumn label="国家 \ IDC" prop="rowName" fixed>
              {{
                default: ({ cell, data }: { cell: string| number, data: ITableDataItem }) => {
                  if (data) {
                    return <div class={['name-cell', { country: data.isCountry }]}>{cell}</div>
                  }
                }
              }}
            </BkTableColumn>
            {
              props.idcList.map(idc => {
                return (
                  <BkTableColumn key={idc.id} label={idc.name} prop={idc.id}>
                    {{
                      default: ({ cell, data }: { cell: string| number, data: ITableDataItem }) => renderValCell(cell, data)
                    }}
                  </BkTableColumn>
                )
              })
            }
          </BkTable>
        </div>
      </div>
    );
  },
});
