import { InfoBox } from 'bkui-vue';
import useTimeoutPoll, { type TimeoutPollAction } from '@/hooks/use-timeout-poll';

const STATIC_PATH = '/static/';
const VERSION_FILE = 'build-hash.txt';

let localVersion = '';
let isShown = false;
let dialog: any = null;
let checkVersionPoll: TimeoutPollAction = null;

const checkVersionChannel = new BroadcastChannel('check-version');
checkVersionChannel.addEventListener('message', ({ data }) => {
  if (data?.type !== 'toggle') {
    return;
  }

  // 共享弹出状态，视之后有无需求多个标签页只允许一次弹出提示使用
  // isShown = data.payload.isShown;
});

const hideDialog = (silent = false) => {
  if (dialog) {
    dialog.hide();
    dialog.destroy();
  }

  isShown = false;
  dialog = null;

  checkVersionPoll.resume();

  if (!silent) {
    checkVersionChannel.postMessage({ type: 'toggle', payload: { isShown: false } });
  }
};

const fetchBuildHash = async () => {
  const response = await fetch(`${STATIC_PATH}${VERSION_FILE}`);

  if (!response.ok) {
    throw new Error('Failed to get build hash');
  }

  const newVersion = await response.text();

  // localVersion 还不存在，认为是第一次加载
  if (!localVersion) {
    localVersion = newVersion;
  }

  if (newVersion !== localVersion) {
    showVersionDialog(newVersion);
  }
};

const checkVersion = async () => {
  try {
    await fetchBuildHash();
  } catch (error) {
    console.error(error);
  }
};

const handleVisibilityChange = () => {
  if (!isShown && document.visibilityState === 'visible') {
    checkVersionPoll.resume();
  } else {
    checkVersionPoll.pause();
  }
};

const showVersionDialog = (newVersion: string) => {
  if (isShown) {
    return;
  }

  dialog = InfoBox({
    width: 520,
    title: '版本更新',
    content: '建议「刷新页面」体验新的特性，「暂不刷新」可能会遇到未知异常，可手动刷新解决。',
    confirmText: '刷新',
    cancelText: '取消',
    onConfirm: () => {
      window.location.reload();
    },
    onClose: () => {
      hideDialog();
    },
  });

  if (dialog) {
    isShown = true;
    localVersion = newVersion;
    checkVersionPoll.pause();
    checkVersionChannel.postMessage({ type: 'toggle', payload: { isShown: true } });
  }
};

const addVisibilityChangeListener = () => {
  document.addEventListener('visibilitychange', handleVisibilityChange);
};

const removeVisibilityChangeListener = () => {
  document.removeEventListener('visibilitychange', handleVisibilityChange);
};

window.addEventListener('beforeunload', () => {
  hideDialog();
  removeVisibilityChangeListener();
  checkVersionPoll.reset();
  checkVersionChannel.close();
});

export const watchVersion = () => {
  checkVersionPoll = useTimeoutPoll(checkVersion, 5 * 60 * 1000, { immediate: true, max: -1 });
  addVisibilityChangeListener();
};
