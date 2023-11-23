import { defineComponent } from 'vue';
import SchemePreview from '../components/scheme-preview';

export default defineComponent({
  name: 'SchemeRecommendationPage',
  setup() {
    return () => <div class='scheme-recommendation-page'>
      <SchemePreview/>
    </div>;
  },
});
