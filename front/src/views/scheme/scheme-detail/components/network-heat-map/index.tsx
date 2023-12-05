import { defineComponent, reactive, ref, watch, PropType, onMounted } from "vue";
import { BkTable, BkTableColumn } from "bkui-vue/lib/table";
import { IIdcInfo, IIdcLatencyListItem, IAreaInfo, IIdcServiceAreaRel } from "@/typings/scheme";
import { useSchemeStore } from "@/store";
import { getScoreColor } from '@/common/util';
import SearchInput from "../../../components/search-input/index";
import RenderTable from './render-table/index';

import './index.scss';

interface ITableDataItem {
  [key: string]: any;
  children?: ITableDataItem[];
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
    const idcData = ref<ITableDataItem[]>([]);
    const idcDataLoading = ref(true);
    const IdcAreaDataLoading = ref(true);
    const activedTab = ref('biz');
    let highlightArea = reactive<IIdcServiceAreaRel[]>([]);
    let avePing = ref(0);
    const containerRef = ref(null);

    const TABS = [
      { id: 'biz', label: '业务数据' },
      { id: 'ping', label: '裸 ping 数据' },
    ];

    watch(() => props.idcList, val => {
      if (val.length > 0) {
        getTableData();
        getIdcAreaData();
      }
    });

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

    // 获取idc区域数据，计算平均延迟以及高亮区域
    const getIdcAreaData  = async () => {
      IdcAreaDataLoading.value = true;
      const idcs = props.idcList.map(item => item.id);
      const res = await schemeStore.queryIdcServiceArea(idcs, props.areaTopo);
      highlightArea = res.data;
      avePing.value = res.data.reduce((acc, crt) => {
        return acc + crt.avg_latency;
      }, 0) / res.data.length;
      IdcAreaDataLoading.value = false;
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
          list.push({
            rowName: country.name,
            isCountry: true,
            isFold: false,
            ...averagedValue,
            children: regionDataList
          })
        }
      });
      return list;
    };

    const handleToggleCountry = (name: string) => {
      const country = idcData.value.find(item => item.rowName === name);
      if (country) {
        country.isFold = !country.isFold;
      }
    }

    onMounted(() => {
      if (props.idcList.length > 0) {
        getTableData();
        getIdcAreaData();
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
          <SearchInput v-model={searchStr.value} width={300} placeholder="请输入国家 \ IDC名称" />
        </div>
        <div class="data-content-wrapper">
          <div class="network-data">
            <div class="data-item">
              <span class="label">网络平均延迟: </span>
              <span class="value">
                {
                  IdcAreaDataLoading.value
                  ?
                  <bk-loading size="mini" theme="primary" mode="spin" />
                  :
                  `${avePing.value.toFixed(2)} ms`
                }
              </span>
            </div>
          </div>
          <bk-checkbox v-model={isHighlight.value}>高亮服务区域</bk-checkbox>
        </div>
        <div class="idc-data-table">
          <bk-loading loading={idcDataLoading.value} style={{ height: '100%' }}>
            {
              idcDataLoading.value
              ?
              null
              :
              <RenderTable
                idcList={props.idcList}
                data={idcData.value}
                searchStr={searchStr.value}
                isHighlight={isHighlight.value}
                highlightArea={highlightArea}
                onToggleFold={handleToggleCountry} />
            }
          </bk-loading>
        </div>
      </div>
    );
  },
});
