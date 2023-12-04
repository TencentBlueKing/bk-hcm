import { defineComponent, reactive, ref, watch, PropType, onMounted } from "vue";
import { BkTable, BkTableColumn } from "bkui-vue/lib/table";
import { IIdcInfo, IIdcLatencyListItem, IAreaInfo, IIdcServiceAreaRel } from "@/typings/scheme";
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
    idcList: Array as PropType<IIdcInfo[]>,
    areaTopo: Array as PropType<IAreaInfo[]>,
  },
  setup (props) {
    const schemeStore = useSchemeStore();

    const searchStr = ref('');
    const isHighlight = ref(false);
    let idcData = ref<ITableDataItem[]>([]);
    const idcDataLoading = ref(false);
    const activedTab = ref('biz');
    let idcServiceArea = reactive<IIdcServiceAreaRel[]>([]);
    const containerRef = ref(null);

    const TABS = [
      { id: 'biz', label: '业务数据' },
      { id: 'ping', label: '裸 ping 数据' },
    ];

    watch(() => props.idcList, val => {
      if (val.length > 0) {
        getTableData();
      }
    });

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
        idcData.value = transToTableData(res.data);
        idcDataLoading.value = false;
      } catch (e) {
        console.log(e);
      }
    };

    const handleSwitchTab = (id: string) => {
      activedTab.value = id;
      getTableData();
    }

    const handleHighlightChange  = async (val: boolean) => {
      isHighlight.value = val;
      if (val) {
        const idcs = props.idcList.map(item => item.id);
        const res = await schemeStore.queryIdcServiceArea(idcs, props.areaTopo);
        idcServiceArea = res.data;
      } else {
        idcServiceArea = [];
      }
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

    const renderValCell = (data: ITableDataItem, id: string) => {
      if (data) {
        const value = Number(data[id]);
        return (<div class="value-cell" style={{ color: getScoreColor(value) }}>{`${value.toFixed(2)}ms`}</div>);
      }
    };

    onMounted(() => {
      if (props.idcList.length > 0) {
        getTableData();
      }
    });

    return () => (
      <div ref={containerRef.value} class="network-heat-map">
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
          <bk-checkbox v-model={isHighlight.value} onChange={handleHighlightChange}>高亮服务区域</bk-checkbox>
        </div>
        <div class="idc-data-table">
          <bk-loading loading={idcDataLoading.value} style={{ height: '100%' }}>
            <BkTable key={activedTab.value} border={['col', 'outer']} data={idcData.value}>
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
                        default: ({ data }: { data: ITableDataItem }) => renderValCell(data, idc.name)
                      }}
                    </BkTableColumn>
                  )
                })
              }
            </BkTable>
          </bk-loading>
        </div>
      </div>
    );
  },
});
