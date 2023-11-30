import { PropType, defineComponent, onMounted, reactive, ref } from "vue";
import { Scene, RasterLayer } from '@antv/l7';
import { TencentMap } from '@antv/l7-maps';
import { IIdcListItem } from "@/typings/scheme";
import { QueryFilterType, QueryRuleOPEnum } from '@/typings/common';
import { useSchemeStore } from "@/store";
import CloudServiceTag from "@/views/scheme/components/cloud-service-tag";

import './index.scss';

export default defineComponent({
  name: 'idc-map-display',
  props: {
    ids: Array as PropType<string[]>,
  },
  setup(props) {
    const schemeStore = useSchemeStore();

    const mapContainerRef = ref();
    const mapIns = ref();
    let idcList = reactive<IIdcListItem[]>([]);
    const idcLoading = ref(false);
    
    // 查询idc机房列表
    const getIdcList = async () => {
      idcLoading.value = true;
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: [{
          field: 'id',
          op: QueryRuleOPEnum.IN,
          value: props.ids,
        }]
      };
      const res = await schemeStore.listIdc(filterQuery, { start: 0, limit: 500 });
      idcList = res.data.details;
    }

    onMounted(() => {
      mapIns.value = new Scene({
        id: mapContainerRef.value,
        map: new TencentMap ({
          style: 'style1',
          center: [ 107.054293, 35.246265 ],
          zoom: 4.056,
        }),
        logoVisible: false,
      });
      getIdcList();
    })
    return () => (
      <div class="idc-map-display">
        <h3 class="title">地图展示</h3>
        <div ref={mapContainerRef} class="map-container"></div>
        <div class="deploy-list-wrapper">
          <div class="idc-group">
            <div class="group-name">就近部署点</div>
            <div class="idc-list">
              {
                idcList.map(idc => {
                  return (
                    <div class="idc-card">
                      <div class="summary-head">
                        <div class="status-dot"></div>
                        <div class="idc-name">{idc.name}</div>
                        <CloudServiceTag type={idc.vendor} small={true} />
                      </div>
                      <div class="cost">IDC 单位成本: $ {idc.price}</div>
                    </div>
                  )
                })
              }
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot"></div>
                  <div class="idc-name">@todo 待删除</div>
                  <CloudServiceTag type={'tcloud'} small={true} showIcon={true} />
                </div>
                <div class="cost">IDC 单位成本: $ 300</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  },
});
