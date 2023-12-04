import { PropType, defineComponent, onMounted, reactive, ref } from "vue";
import { Scene, PolygonLayer, LineLayer, Zoom  } from '@antv/l7';
import { Mapbox } from '@antv/l7-maps';
import { IIdcInfo } from "@/typings/scheme";
import geoData from "@/constants/geo-data";
import CloudServiceTag from "@/views/scheme/components/cloud-service-tag";

import './index.scss';

export default defineComponent({
  name: 'idc-map-display',
  props: {
    list: Array as PropType<IIdcInfo[]>,
  },
  setup(props) {
    const mapContainerRef = ref();
    const mapIns = ref();

    const renderMap = () => {
      mapIns.value = new Scene({
        id: mapContainerRef.value,
        map: new Mapbox ({
          pinch: 0,
          center: [107.054293, 35.246265],
          style: 'blank',
        }),
        logoVisible: false,
      });
      mapIns.value.on('loaded', () => {
        // 行政区块
        const polygonLayer = new PolygonLayer({})
          .source(geoData)
          .color('name', [
            // '#3762B8', '#3E96C2', '#61B2C2', '#85CCA8', '#FFC685', '#FFA66B', '#F5876C', '#D66F6B',
            // '#3A84FF', '#699DF4', '#2DCB56', '#6AD57C', '#FF9C01', '#FFB848', '#EA3636', '#EA5858',
            '#d0d0d0'
          ])
          .shape('fill')
          .style({ opacity: 1 });

        // 区块分界线
        const lineLayer = new LineLayer({
          zIndex: 2,
        })
          .source(geoData)
          .color('#ffffff')
          .size(0.6)
          .style({ opacity: 1 });

        // 缩放组件
        const zoom = new Zoom({
          zoomInTitle: '放大',
          zoomOutTitle: '缩小',
        });

        mapIns.value.addLayer(polygonLayer);
        mapIns.value.addLayer(lineLayer);
        mapIns.value.addControl(zoom);
      })
    };

    onMounted(() => {
      renderMap();
    })

    return () => (
      <div class="idc-map-display">
        <h3 class="title">地图展示</h3>
        <div class="map-area">
          <div ref={mapContainerRef} id="map-container" class="map-container"></div>
        </div>
        <div class="deploy-list-wrapper">
          <div class="idc-group">
            <div class="group-name">就近部署点</div>
            <div class="idc-list">
              {
                props.list.map(idc => {
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
            </div>
          </div>
        </div>
      </div>
    )
  },
});
