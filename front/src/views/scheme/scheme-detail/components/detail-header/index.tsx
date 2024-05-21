import { defineComponent, PropType } from 'vue';
import { DEPLOYMENT_ARCHITECTURE_MAP } from '@/constants';
import { ISchemeSelectorItem } from '@/typings/scheme';
import SchemeSelector from '@/views/scheme/components/scheme-selector';
import CloudServiceTag from '@/views/scheme/components/cloud-service-tag';

import './index.scss';

interface ISchemeData {
  deployment_architecture: string[];
  vendors: string[];
  composite_score: number;
  net_score: number;
  cost_score: number;
}

export default defineComponent({
  name: 'SchemeDetailHeader',
  props: {
    schemeList: Array as PropType<ISchemeSelectorItem[]>,
    schemeListLoading: Boolean,
    schemeData: {
      type: Object as PropType<ISchemeData>,
      default: () => ({}),
    },
    showEditIcon: Boolean,
    selectFn: Function,
    onBack: Function,
  },
  emits: ['update'],
  setup(props, ctx) {
    const scores = [
      { id: 'composite_score', name: '综合评分' },
      { id: 'net_score', name: '网络评分' },
      { id: 'cost_score', name: '成本评分' },
    ];

    return () => (
      <div class='scheme-detail-header'>
        <div class='header-content'>
          <SchemeSelector
            schemeList={props.schemeList}
            schemeListLoading={props.schemeListLoading}
            showEditIcon={props.showEditIcon}
            schemeData={props.schemeData}
            selectFn={props.selectFn}
            onBack={props.onBack}
            onUpdate={(data) => {
              ctx.emit('update', data);
            }}
          />
          <div class='tag-list'>
            {props.schemeData.deployment_architecture.map((item: string) => {
              return <div class='deploy-type'>{DEPLOYMENT_ARCHITECTURE_MAP[item]}</div>;
            })}
            {props.schemeData.vendors.map((item: string) => {
              return <CloudServiceTag class='cloud-service-type' type={item} showIcon={true} />;
            })}
          </div>
          <div class='score-nums'>
            {scores.map((item) => {
              return (
                <div class='num-item' key={item.id}>
                  <span class='label'>{item.name}：</span>
                  <span class='val'>{props.schemeData[item.id]}</span>
                </div>
              );
            })}
          </div>
        </div>
        <div class='operate-area'>{ctx.slots.operate()}</div>
      </div>
    );
  },
});
