import { type RouteLocationRaw } from 'vue-router';
import router from '@/router';

interface IRedirectOptions {
  history: boolean;
  back: boolean;
  reload: boolean;
  replace: boolean;
}

export default {
  redirect(to: RouteLocationRaw, options?: IRedirectOptions) {
    const { reload = false, replace = false } = options || {};
    if (reload) {
      const { href } = router.resolve(to);
      window.location.href = href;
      window.location.reload();
    } else {
      const action = replace ? 'replace' : 'push';
      router[action](to);
    }
  },
};
