import { defineComponent } from 'vue';
import './index.scss';
import DetailHeader from '../../scheme-detail/components/detail-header';
import { useSchemeStore } from '@/store';
import { Button } from 'bkui-vue';
// import IdcMapDisplay from '../../scheme-detail/components/idc-map-display';
// import NetworkHeatMap from '../../scheme-detail/components/network-heat-map';

export default defineComponent({
  setup() {
    const schemeStore = useSchemeStore();
    return () => (
      <div>
        <DetailHeader
          schemeData={schemeStore.schemeData}
          schemeList={[]}
        >
          {{
            operate: () => <Button theme='primary'>保存</Button>,
          }}
        </DetailHeader>
      </div>
    );
  },
});
