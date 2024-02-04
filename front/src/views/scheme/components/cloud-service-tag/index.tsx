import { defineComponent, reactive, computed, watch } from 'vue';

import './index.scss';

export default defineComponent({
  name: 'SchemeSelector',
  props: {
    type: String,
    small: {
      type: Boolean,
      default: false,
    },
    showIcon: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const CLOUD_SERVICE_MAP = {
      tcloud: {
        name: '腾讯云',
        color: '#4193E5',
        bgColor: '#DAE9FD',
        icon: 'bkhcm-icon-tengxunyun',
      },
      aws: {
        name: 'AWS',
        color: '#E68D00',
        bgColor: '#FFF2C9',
        icon: 'bkhcm-icon-yamaxunyun',
      },
      azure: {
        name: '微软云',
        color: '#45A0A5',
        bgColor: '#D8F4F5',
        icon: 'bkhcm-icon-weiruanyun',
      },
      gcp: {
        name: '谷歌云',
        color: '#3FAA3B',
        bgColor: '#DAF5C8',
        icon: 'bkhcm-icon-gugeyun',
      },
      huawei: {
        name: '华为云',
        color: '#EA4646',
        bgColor: '#FFDDDD',
        icon: 'bkhcm-icon-huaweiyun',
      },
    };

    let cloudData = reactive<{ [key: string]: string }>({});

    watch(
      () => props.type,
      (val) => {
        cloudData = CLOUD_SERVICE_MAP[val] || {};
      },
      {
        immediate: true,
      },
    );

    const styleObj = computed(() => {
      if (CLOUD_SERVICE_MAP[props.type]) {
        return {
          color: `${cloudData.color}`,
          background: `${cloudData.bgColor}`,
        };
      }
      return {};
    });

    return () => (
      <div class={['cloud-service-tag', props.small ? 'small-tag' : '']} style={styleObj.value}>
        {props.showIcon ? <i class={['cloud-icon hcm-icon', cloudData.icon]}></i> : null}
        <span class='tag-name-text'>{cloudData.name || '--'}</span>
      </div>
    );
  },
});
