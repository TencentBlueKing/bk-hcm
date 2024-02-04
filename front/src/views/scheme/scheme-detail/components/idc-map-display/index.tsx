import { PropType, defineComponent, onMounted, ref, watch } from 'vue';
import { Scene, PolygonLayer, LineLayer, Zoom, Popup, PointLayer } from '@antv/l7';
import { Mapbox } from '@antv/l7-maps';
import { IIdcInfo } from '@/typings/scheme';
import geoData from '@/constants/geo-data';
import IdcReginData from '@/constants/idc-region-data';
import CloudServiceTag from '@/views/scheme/components/cloud-service-tag';

import './index.scss';

export default defineComponent({
  name: 'IdcMapDisplay',
  props: {
    list: Array as PropType<IIdcInfo[]>,
  },
  setup(props) {
    const REGION_MAP_COLORS = [
      '#3762B8',
      '#3E96C2',
      '#61B2C2',
      '#85CCA8',
      '#FFC685',
      '#FFA66B',
      '#F5876C',
      '#D66F6B',
      '#3A84FF',
      '#699DF4',
      '#2DCB56',
      '#6AD57C',
      '#FF9C01',
      '#FFB848',
      '#EA3636',
      '#EA5858',
    ];

    const mapContainerRef = ref();
    const mapIns = ref();

    watch(
      () => props.list,
      () => {
        renderMap();
      },
    );

    const renderMap = () => {
      if (mapIns.value) {
        mapIns.value.destroy();
      }

      if (props.list.length === 0) {
        return;
      }

      const countryColors = {};
      // eslint-disable-next-line array-callback-return
      props.list.map((item, index) => {
        const countryName = transCountryName(item.country);
        countryColors[countryName] = REGION_MAP_COLORS[index % REGION_MAP_COLORS.length];
      });

      mapIns.value = new Scene({
        id: mapContainerRef.value,
        map: new Mapbox({
          pinch: 0,
          minZoom: 0.5,
          center: [75.054293, 35.246265],
          style: 'blank',
        }),
        logoVisible: false,
      });
      mapIns.value.on('loaded', () => {
        // 行政区块
        const polygonLayer = new PolygonLayer({})
          .source(geoData)
          .color('name', (value) => {
            return countryColors?.[value] || '#d0d0d0';
          })
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

        const regions: { type: string; features: any[] } = {
          type: 'FeatureCollection',
          features: [],
        };
        props.list.forEach((item) => {
          const index = IdcReginData.features.findIndex((idc) => idc.properties.region === item.region);
          if (index > -1) {
            regions.features.push(IdcReginData.features[index]);
          }
        });

        const pointLayer = new PointLayer({
          zIndex: 3,
        })
          .source(regions)
          .shape('simple')
          .size(5)
          .color('#FF5656')
          .style({
            strokeColor: '#ffffff',
            strokeWidth: 2,
          });

        // 缩放组件
        const zoom = new Zoom({
          zoomInTitle: '放大',
          zoomOutTitle: '缩小',
        });

        mapIns.value.addLayer(polygonLayer);
        mapIns.value.addLayer(lineLayer);
        mapIns.value.addLayer(pointLayer);
        mapIns.value.addControl(zoom);

        let popup: Popup;
        polygonLayer.on('mousemove', (e) => {
          const { name } = e.feature.properties;
          console.log(name);
          if (props.list.find((item) => transCountryName(item.country) === name)) {
            popup = new Popup({
              offsets: [0, 0],
              closeButton: false,
            })
              .setLnglat(e.lngLat)
              .setHTML(`<span>${e.feature.properties.name}</span>`);

            mapIns.value.addPopup(popup);
          } else {
            if (popup) {
              mapIns.value.removePopup(popup);
            }
          }
        });
      });
    };

    const transCountryName = (name: string) => {
      if (['中国香港', '中国澳门', '中国台湾'].includes(name)) {
        return '中国';
      }
      return name;
    };

    onMounted(() => {
      renderMap();
    });

    return () => (
      <div class='idc-map-display'>
        <h3 class='title'>地图展示</h3>
        <div class='map-area'>
          <div ref={mapContainerRef} id='map-container' class='map-container'></div>
        </div>
        <div class='deploy-list-wrapper'>
          <div class='idc-group'>
            <div class='group-name'>就近部署点</div>
            <div class='idc-list'>
              {props.list.map((idc) => {
                return (
                  <div class='idc-card'>
                    <div class='summary-head'>
                      <div class='status-dot'></div>
                      <div class='idc-name'>{idc.region}机房</div>
                      <CloudServiceTag type={idc.vendor} small={true} />
                    </div>
                    <div class='cost'>IDC 单位成本: {idc.price}</div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    );
  },
});
