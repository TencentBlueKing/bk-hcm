import { type RouteLocationRaw, RouteQueryAndHash } from 'vue-router';
import router from '@/router';
import { HistoryStorage } from './history-storage';

interface IRedirectOptions {
  history?: boolean;
  back?: boolean;
  reload?: boolean;
  replace?: boolean;
}

export default {
  redirect(to: RouteLocationRaw, options?: IRedirectOptions) {
    const { query = {}, ...other } = { ...(to as RouteQueryAndHash) };
    const { currentRoute } = router;
    const { reload = false, replace = false, history = false, back = false } = options || {};

    // 当前页非history模式则先清空历史记录
    if (!Object.hasOwn(currentRoute.value.query, '_f')) {
      HistoryStorage.clear();
    }

    if (history) {
      const data = {
        name: currentRoute.value.name,
        params: { ...currentRoute.value.params },
        query: { ...currentRoute.value.query },
      };
      // 置入标志参数
      query._f = '1';

      HistoryStorage.append(data);
    } else if (back) {
      // 后退的目标就是栈顶记录（当前页本身从未入栈），出栈后目标页的栈顶才是它自己的来源页。
      // 不能按当前页 name 删除：会删掉不该删的同名记录（如 主机→硬盘→主机），并让栈顶残留目标页自身，导致目标页的返回原地跳转无反应
      try {
        HistoryStorage.removeLast();
      } catch (err) {
        console.error(err);
      }
    }

    const newTo = {
      query,
      ...other,
    };

    if (reload) {
      const { href } = router.resolve(newTo);
      window.location.href = href;
      window.location.reload();
    } else {
      const action = replace ? 'replace' : 'push';
      router[action](newTo);
    }
  },
  back() {
    if (Object.hasOwn(router.currentRoute.value.query, '_f')) {
      try {
        HistoryStorage.pop();
      } catch (error) {
        router.go(-1);
      }
    } else {
      router.go(-1);
    }
  },
  open(to: RouteLocationRaw) {
    const { href } = router.resolve(to);
    window.open(href);
  },
};
