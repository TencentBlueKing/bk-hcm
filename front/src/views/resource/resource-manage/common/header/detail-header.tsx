import {
  defineComponent,
} from 'vue';

import {
  ArrowsLeft,
} from 'bkui-vue/lib/icon';

import {
  useRouter,
} from 'vue-router';

import './detail-header.scss';

export default defineComponent({
  components: {
    ArrowsLeft,
  },

  setup() {
    const router = useRouter();

    const goBack = () => {
      router.back();
    };

    return {
      goBack,
    };
  },

  render() {
    return <>
      <section class="detail-header-main">
        <span>
          <arrows-left
            class="detail-header-arrows-left"
            onClick={this.goBack}
          />
          {
            this.$slots.default?.()
          }
        </span>
        <span>
          {
            this.$slots.right?.()
          }
        </span>
      </section>
    </>;
  },
});
