/**
 * setTimeout 模拟 setInterval
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
