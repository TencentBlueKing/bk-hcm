import { defineComponent, reactive } from "vue";

import './index.scss';
import { getRandomValues } from "crypto";

export default defineComponent({
  name: 'scheme-info-card',
  setup () {
    const schemeDetail = reactive({
      "biz_type": "游戏",
      "user_distribution": [
        {
          "name": "country_1",
          "children": [
            {
              "name": "province_1_1",
              "value": 0.1
            },
            {
              "name": "province_1_2",
              "value": 0.1
            }
          ]
        },
        {
          "name": "country_2",
          "children": [
            {
              "name": "province_2_1",
              "value": 0.1
            },
            {
              "name": "province_2_2",
              "value": 0.1
            }
          ]
        }
      ],
      "cover_ping": 120,
      "cover_rate": 80,
      "deployment_architecture": [
        "xxxx"
      ]
    });

    const infos = [
      { id: 'user_distribution', name: '用户据分布地区' },
      { id: 'biz_type', name: '业务类型' },
      { id: 'network', name: '用户网络容忍' },
      { id: 'user_rate', name: '用户分布占比' },
      { id: 'deployment_architecture', name: '部署架构' },
    ];

    const getValue = (id: string) => {
      switch (id) {
        case 'user_distribution':
          return schemeDetail.user_distribution.map(item => item.name).join(', ');
        case 'network':
          return `网络延迟 < @todo待确认、ping抖动 < ${schemeDetail.cover_ping}、丢包率 < ${schemeDetail.cover_rate}%`
        default:
          return schemeDetail[id];
      }
    };

    return () => (
      <div class="scheme-info-card">
        <div class="info-list">
          {
            infos.map(item => {
              return (
                <div class="info-item" key={item.id}>
                  <span class="label">{item.name}：</span>
                  <span class="value">{item.id === 'user_rate' ? <bk-button text theme="primary">查看详情</bk-button> : getValue(item.id)}</span>
                </div>
              )
            })
          }
        </div>
        <div class="recreate-btn">
          <bk-button outline theme="primary">重新生成</bk-button>
        </div>
      </div>
    )
  },
});
