import { defineComponent, reactive } from "vue";
import SchemeSelector from "@/views/scheme/components/scheme-selector";

import './index.scss';

export default defineComponent({
  name: 'scheme-detail-header',
  setup () {
    const schemeDetail = reactive({
      composite_score: 80,
      net_score: 70,
      cost_score: 90,
    })
    const scores = [
      { id: 'composite_score', name: '综合评分' },
      { id: 'net_score', name: '网络评分' },
      { id: 'cost_score', name: '方案成本' },
    ]

    return () => (
      <div class="scheme-detail-header">
        <div class="header-content">
          <SchemeSelector />
          <div class="tag-list">
            <div class="tag deploy-type">分布式部署</div>
            <div class="tag cloud-service-type">
              <i class="hcm-icon bkhcm-icon-tengxunyun tencent-cloud-icon"></i>
              腾讯云
            </div>
          </div>
          <div class="score-nums">
            {
              scores.map(item => {
                return (
                  <div class="num-item" key={item.id}>
                    <span class="label">{item.name}：</span>
                    <span class="val">{item.id === 'cost_score' ? '$ ' : ''}{schemeDetail[item.id]}</span>
                  </div>
                )
              })
            }
          </div>
        </div>
        <div class="operate-area">
          <bk-button>删除</bk-button>
        </div>
      </div>
    )
  },
});
