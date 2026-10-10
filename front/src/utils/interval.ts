/**
 * setTimeout 模拟 setInterval
 *
 * @deprecated 轮询统一用 `@/hooks/use-timeout-poll`。这里只返回 start/clear，
 * 回调在创建时就被闭包固定，取不到后续变化的参数（账号切换类场景会打到上一个 id）；
 * 也没有 isActive、次数重置与 scope 自动销毁。
 */
export default function interval(func: Function, wait: number, max = 0) {
  let timer: any = null;
  const count: number = Math.floor(max / wait);
  let now = 0;

  const interv = function (nowTimer: any) {
    if (timer !== nowTimer || (count > 0 && count < (now = now + 1))) return;
    func.call(null);
    setTimeout(() => interv(nowTimer), wait);
  };

  const clearTimeInterval = () => {
    clearTimeout(timer);
    now = 0;
    timer = null;
  };
  const setTimeInterval = () => {
    if (timer) return;
    timer = setTimeout(() => interv(timer), wait);
  };

  return {
    clearTimeInterval,
    setTimeInterval,
  };
}
