import { defineComponent, reactive, computed, watch } from "vue";

import './index.scss';

export default defineComponent({
  name: 'scheme-selector',
  props: {
    type: String,
  },
  setup (props) {
    const CLOUD_SERVICE_MAP = {
      tcloud: {
        name: '腾讯云',
        color: '#4193E5',
        bgColor: '#DAE9FD',
        icon: '',
      },
      aws: {
        name: 'AWS',
        color: '#E68D00',
        bgColor: '#FFF2C9',
        icon: '',
      },
      azure: {
        name: '微软云',
        color: '#45A0A5',
        bgColor: '#D8F4F5',
        icon: '',
      },
      gcp: {
        name: '谷歌云',
        color: '#3FAA3B',
        bgColor: '#DAF5C8',
        icon: '',
      },
      huawei: {
        name: '华为云',
        color: '#EA4646',
        bgColor: '#FFDDDD',
        icon: '',
      },
    }

    let cloudData = reactive<{ [key: string]: string; }>({})

    watch(() => props.type, (val) => {
      cloudData = CLOUD_SERVICE_MAP[val] || {};
    }, {
      immediate: true
    });

    const styleObj = computed(() => {
      if (CLOUD_SERVICE_MAP[props.type]) {
        return {
          color: `${cloudData.color}`,
          background: `${cloudData.bgColor}`,
        }
      }
      return {};
    })

    return () => (
      <div class="cloud-service-tag" style={styleObj.value}>
        <span class='name-text'>{ cloudData.name || '--' }</span>
      </div>
    );
  },
});
