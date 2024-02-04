import { defineComponent } from 'vue';
import blank_1 from '@/assets/image/scheme-blank-1.png';
import blank_2 from '@/assets/image/scheme-blank-2.png';
import './index.scss';

export default defineComponent({
  name: 'SchemeBlankPage',
  setup() {
    return () => (
      <div class='blank-show-wrap'>
        <div class='item-wrap'>
          <img src={blank_1} alt='' />
          <div class='title-wrap'>
            <span class='serial-number mr8'>1</span>
            <span class='title-text'>配置基本信息</span>
          </div>
          <div class='content-wrap'>配置业务基本属性，并查看初步的部署架构、用户分布、部署方案的推荐结果。</div>
        </div>
        <i class='hcm-icon bkhcm-icon-arrows-up separator'></i>
        <div class='item-wrap'>
          <img src={blank_2} alt='' />
          <div class='title-wrap'>
            <span class='serial-number mr8'>2</span>
            <span class='title-text'>查看方案详情</span>
          </div>
          <div class='content-wrap'>查看方案结果，并基于方案分析与网络、成本数据进一步微调部署方案。</div>
        </div>
      </div>
    );
  },
});
