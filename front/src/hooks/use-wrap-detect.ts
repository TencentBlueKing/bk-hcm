import { onBeforeUnmount, toValue, watch, type MaybeRefOrGetter } from 'vue';

export interface IWrapDetectOptions {
  /** 需要检测换行的容器选择器，如工具栏 `.toolbar`、操作条 `.top-bar`、标签组等 */
  selector: string;
  /** 只在该祖先容器内扫描，默认扫描整个 root */
  scope?: string;
  /** 换行时追加到容器上的类名，默认 `is-wrapped`；传空字符串则只走 onChange、不改 class */
  wrapClass?: string;
  /** 换行状态变化时回调（首次计算也会触发一次），可用于自定义换行后的行为 */
  onChange?: (wrapped: boolean, el: HTMLElement) => void;
}

/**
 * 通用换行检测
 *
 * 纯 CSS 无法表达「是否发生过换行」，这里按容器直接子项是否在同一行（offsetTop 是否相同）判断，
 * 换行时给容器加类名并回调 onChange，适用于工具栏、操作条等靠 flex-wrap 换行的区域。
 *
 * 尺寸变化由 ResizeObserver 覆盖（窗口缩放、侧栏收起等），
 * DOM 重建由 MutationObserver 覆盖（tab 切换、子视图切换等，rAF 去抖），
 * 根节点销毁或组件卸载时自动断开。
 *
 * @param root 扫描起点（一般是页面根节点的模板 ref）
 */
export const useWrapDetect = (root: MaybeRefOrGetter<HTMLElement | null | undefined>, options: IWrapDetectOptions) => {
  const { selector, scope = '', wrapClass = 'is-wrapped', onChange } = options;

  let resizeObserver: ResizeObserver | null = null;
  let mutationObserver: MutationObserver | null = null;
  let observeTarget: HTMLElement | null = null;
  let frame = 0;

  // 缓存每个容器上一次的换行状态，只在状态变化时才切 class / 回调
  const wrapState = new WeakMap<HTMLElement, boolean>();

  const syncWrapState = (el: HTMLElement) => {
    const tops = Array.from(el.children, (child) => (child as HTMLElement).offsetTop);
    const wrapped = tops.some((top) => top !== tops[0]);
    if (wrapState.get(el) === wrapped) return;

    wrapState.set(el, wrapped);
    if (wrapClass) el.classList.toggle(wrapClass, wrapped);
    onChange?.(wrapped, el);
  };

  const resolveScanRoot = () => {
    const el = toValue(root);
    if (!el) return null;
    return (scope && el.querySelector<HTMLElement>(scope)) || el;
  };

  // 面板重建后容器是新的节点，收集时一并重新挂载尺寸监听
  const refresh = () => {
    resizeObserver?.disconnect();
    const scanRoot = resolveScanRoot();
    if (!scanRoot) return;

    scanRoot.querySelectorAll<HTMLElement>(selector).forEach((el) => {
      syncWrapState(el);
      resizeObserver?.observe(el);
    });

    if (mutationObserver && observeTarget !== scanRoot) {
      mutationObserver.disconnect();
      observeTarget = scanRoot;
      mutationObserver.observe(scanRoot, { childList: true, subtree: true });
    }
  };

  const scheduleRefresh = () => {
    if (typeof requestAnimationFrame === 'undefined') {
      refresh();
      return;
    }
    if (frame) return;
    frame = requestAnimationFrame(() => {
      frame = 0;
      refresh();
    });
  };

  watch(
    () => toValue(root),
    (el) => {
      resizeObserver?.disconnect();
      mutationObserver?.disconnect();
      observeTarget = null;
      if (!el || typeof ResizeObserver === 'undefined') return;

      resizeObserver = new ResizeObserver((entries) => {
        entries.forEach(({ target }) => syncWrapState(target as HTMLElement));
      });
      mutationObserver = new MutationObserver(scheduleRefresh);
      refresh();
    },
    { immediate: true, flush: 'post' },
  );

  onBeforeUnmount(() => {
    resizeObserver?.disconnect();
    mutationObserver?.disconnect();
    if (frame) cancelAnimationFrame(frame);
  });

  return { refresh };
};
