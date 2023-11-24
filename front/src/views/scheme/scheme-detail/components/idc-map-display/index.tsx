import { defineComponent, onMounted, ref } from "vue";
import { Scene, RasterLayer } from '@antv/l7';
import { TencentMap } from '@antv/l7-maps';

import './index.scss';

export default defineComponent({
  name: 'idc-map-display',
  setup () {

    const mapContainerRef = ref();
    const mapIns = ref();

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
    })
    return () => (
      <div class="idc-map-display">
        <h3 class="title">地图展示</h3>
        <div ref={mapContainerRef} class="map-container"></div>
        <div class="deploy-list-wrapper">
          <div class="idc-group">
            <div class="group-name">集中部署点</div>
            <div class="idc-list">
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot blue"></div>
                  <div class="idc-name">法兰克福机房</div>
                  <div class="cloud-tag"><span class="name">腾讯云</span></div>
                </div>
                <div class="cost">IDC 单位成本: $ 300</div>
              </div>
            </div>
          </div>
          <div class="idc-group">
            <div class="group-name">就近部署点</div>
            <div class="idc-list">
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot"></div>
                  <div class="idc-name">布宜诺斯艾利斯</div>
                  <div class="cloud-tag"><span class="name">腾讯云</span></div>
                </div>
                <div class="cost">IDC 单位成本: $ 300</div>
              </div>
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot"></div>
                  <div class="idc-name">华盛顿机房</div>
                  <div class="cloud-tag"><span class="name">腾讯云</span></div>
                </div>
                <div class="cost">IDC 单位成本: $ 300</div>
              </div>
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot"></div>
                  <div class="idc-name">华盛顿机房</div>
                  <div class="cloud-tag"><span class="name">腾讯云</span></div>
                </div>
                <div class="cost">IDC 单位成本: $ 300</div>
              </div>
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot"></div>
                  <div class="idc-name">华盛顿机房</div>
                  <div class="cloud-tag"><span class="name">腾讯云</span></div>
                </div>
                <div class="cost">IDC 单位成本: $ 300</div>
              </div>
              <div class="idc-card">
                <div class="summary-head">
                  <div class="status-dot"></div>
                  <div class="idc-name">华盛顿机房</div>
                  <div class="cloud-tag"><span class="name">腾讯云</span></div>
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
