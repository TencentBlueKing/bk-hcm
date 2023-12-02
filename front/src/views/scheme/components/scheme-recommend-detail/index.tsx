import { defineComponent } from 'vue';
import './index.scss';
import DetailHeader from '../../scheme-detail/components/detail-header';
import { useSchemeStore } from '@/store';
import { Button } from 'bkui-vue';
// import IdcMapDisplay from '../../scheme-detail/components/idc-map-display';
// import NetworkHeatMap from '../../scheme-detail/components/network-heat-map';

export default defineComponent({
  props: {
    onBack: {
      type: Function,
      required: true,
    },
  },
  setup(props) {
    const schemeStore = useSchemeStore();
    return () => (
      <div>
        <DetailHeader
          schemeData={schemeStore.schemeData}
          schemeList={schemeStore.recommendationSchemes.map((v, idx) => ({
            id: idx,
            name: `方案${idx}`,
            ...v,
          }))}
          onBack={props.onBack}
        >
          {{
            operate: () => <Button theme='primary'>保存</Button>,
          }}
        </DetailHeader>
      </div>
    );
  },
});
