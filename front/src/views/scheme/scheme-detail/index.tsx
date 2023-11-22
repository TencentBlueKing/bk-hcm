import { defineComponent } from "vue";
import DetailHeader from "./components/detail-header";
import SchemeInfoCard from "./components/scheme-info-card";
import IdcMapDisplay from "./components/idc-map-display";
import NetworkHeatMap from "./components/network-heat-map";

import './index.scss';

export default defineComponent({
  name: 'scheme-detail-page',
  setup () {
    return () => (
      <div class="scheme-detail-page">
        <DetailHeader />
        <section class="detail-content-area">
          <SchemeInfoCard />
          <section  class="chart-content-wrapper">
            <IdcMapDisplay />
            <NetworkHeatMap />
          </section>
        </section>
      </div>
    )
  },
});
