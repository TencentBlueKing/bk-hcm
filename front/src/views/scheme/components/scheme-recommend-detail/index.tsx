import { defineComponent } from 'vue';
import './index.scss';
import DetailHeader from '../../scheme-detail/components/detail-header';
import IdcMapDisplay from '../../scheme-detail/components/idc-map-display';
import NetworkHeatMap from '../../scheme-detail/components/network-heat-map';

export default defineComponent({
  setup() {
    return () => (
      <div>
        <DetailHeader/>
        <section>
          <IdcMapDisplay />
          <NetworkHeatMap />
        </section>
      </div>
    );
  },
});
