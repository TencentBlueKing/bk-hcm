import { useStore } from '@/store';
import { computed } from 'vue';

export function useUser() {
  const store = useStore();
  const user = computed(() => store.state.user);

  return user;
}
