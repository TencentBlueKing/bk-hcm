import { showLoginModal as showModal } from '@blueking/login-modal';

export const showLoginModal = () => {
  const { origin, pathname } = window.location;
  const successUrl = `${origin}${pathname.endsWith('/') ? pathname : `${pathname}/`}static/login_success.html`;

  const loginBaseUrl = window.PROJECT_CONFIG.BK_LOGIN_URL || '';
  if (!loginBaseUrl) {
    console.error('Login URL not configured!');
    return;
  }

  const loginURL = new URL(loginBaseUrl);
  loginURL.searchParams.set('c_url', successUrl);
  const loginPathname = loginURL.pathname.endsWith('/') ? loginURL.pathname : `${loginURL.pathname}/`;
  const loginUrl = `${loginURL.origin}${loginPathname}plain/${loginURL.search}`;

  showModal({
    loginUrl,
    onFail: () => {
      gotoLoginPage();
    },
  });
};

export const gotoLoginPage = (url?: string, isLogout = false) => {
  const rawUrl = url ?? window.PROJECT_CONFIG.BK_LOGIN_URL;
  if (!rawUrl) {
    console.error('The login URL is not configured!');
    return;
  }
  try {
    const loginURL = new URL(rawUrl);
    loginURL.searchParams.set('c_url', location.href);

    if (isLogout) {
      loginURL.searchParams.set('is_from_logout', '1');
    }

    location.href = loginURL.href;
  } catch (_) {
    console.error('The login URL invalid!');
  }
};
