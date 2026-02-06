import alertIconSrc from '@/assets/image/alert.svg';
import chromeIconSrc from '@/assets/image/chrome.svg';

// 配置项
export interface BannerConfig {
  animationDuration?: number;
  autoShow?: boolean;
  bannerClass?: string;
  bannerHeight?: number;
  bannerId?: string;
  downloadUrl?: string;
  floating?: boolean;
  zIndex?: number;
}

/**
 * Chrome 浏览器检测横幅组件
 */
class ChromeBanner {
  private bannerElement: HTMLElement | null = null;
  private config: BannerConfig;
  // 默认配置
  private defaultConfig: BannerConfig = {
    floating: false,
    autoShow: true,
    bannerId: 'chrome-detection-banner',
    bannerClass: 'chrome-banner',
    bannerHeight: 52,
    animationDuration: 400,
    zIndex: 9999,
    downloadUrl: 'https://www.google.com/chrome/',
  };

  private isVisible = false;
  private originalPaddingTop: string | undefined = undefined;

  constructor(config: BannerConfig = {}) {
    this.config = { ...this.defaultConfig, ...config };
    // 如果不是 Chrome 浏览器
    if (!this.isChrome()) {
      // 创建横幅元素
      this.createBannerElement();
    }
    // 如果配置了自动显示，则显示横幅
    if (this.config.autoShow) {
      this.show();
    }
  }

  /**
   * 销毁横幅，清理资源
   */
  public destroy(): void {
    if (this.bannerElement?.parentNode) {
      this.bannerElement.parentNode.removeChild(this.bannerElement);
    }

    if (!this.config.floating) {
      document.body.style.paddingTop = this.originalPaddingTop || '';
    }

    this.bannerElement = null;
    this.isVisible = false;
  }

  /**
   * 隐藏横幅
   */
  public hide(): void {
    if (!this.bannerElement || !this.isVisible) return;

    // 添加隐藏动画
    this.bannerElement.style.opacity = '0';

    // 动画完成后隐藏元素
    setTimeout(() => {
      if (this.bannerElement) {
        this.bannerElement.style.display = 'none';
      }

      if (!this.config.floating) {
        document.body.style.paddingTop = this.originalPaddingTop || '';
      }

      this.isVisible = false;
    }, this.config.animationDuration);
  }

  /**
   * 显示横幅
   */
  public show(): void {
    if (!this.bannerElement || this.isVisible) return;

    // 显示横幅
    this.bannerElement.style.display = 'block';

    // 添加显示动画
    setTimeout(() => {
      if (this.bannerElement) {
        this.bannerElement.style.opacity = '1';
      }
    }, 10);

    // 记录原始paddingTop
    if (!this.originalPaddingTop) {
      this.originalPaddingTop = window.getComputedStyle(document.body).paddingTop || '';
    }
    // 非浮动时，增加paddingTop
    if (!this.config.floating) {
      const originalPadding = parseFloat(this.originalPaddingTop) || 0;
      document.body.style.paddingTop = `${originalPadding + (this.config.bannerHeight ?? 0)}px`;
    }

    this.isVisible = true;
  }

  /**
   * 切换横幅显示状态
   */
  public toggle(): void {
    if (this.isVisible) {
      this.hide();
    } else {
      this.show();
    }
  }

  private isChrome(): boolean {
    const userAgent = navigator.userAgent.toLowerCase();
    // Chrome 有特定的 window.chrome 对象
    const hasChromeAPI = !!(window as any).chrome;

    return /chrome/.test(userAgent) && !/edge/.test(userAgent) && hasChromeAPI;
  }

  /**
   * 添加事件监听器
   */
  private addEventListeners(): void {
    if (!this.bannerElement) return;

    // 关闭按钮点击事件
    const closeBtn = this.bannerElement.querySelector('.close-btn');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => {
        this.hide();
      });
    }

    // 下载按钮点击事件
    const downloadBtn = this.bannerElement.querySelector('.download-btn');
    if (downloadBtn) {
      downloadBtn.addEventListener('click', () => {
        window.open(this.config.downloadUrl, '_blank');
      });
    }
  }

  /**
   * 应用横幅样式
   */
  private applyBannerStyles(): void {
    if (!this.bannerElement) return;

    const styles = `
            #${this.config.bannerId} {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: ${this.config.bannerHeight}px;
                background: #FDF7E8;
                color: #4D4F56;
                display: none;
                z-index: ${this.config.zIndex};
                opacity: 0;
                transition: opacity ${this.config.animationDuration}ms ease;
            }

            #${this.config.bannerId} .banner-content {
                width: 100%;
                height: 100%;
                display: flex;
                align-items: center;
                flex-wrap: wrap;
                gap: 16px;
                font-size: 14px;
                padding: 0 8px 0 18px;
                box-sizing: border-box;
            }
            #${this.config.bannerId} .banner-message {
                display: flex;
                align-items: center;
                gap: 14px;
            }

            #${this.config.bannerId} .alert-icon {
                width: 24px;
            }

            #${this.config.bannerId} .download-btn {
                  position: relative;
                  background: #fff;
                  color: #3A84FF;
                  border-radius: 16px;
                  font-size: 14px;
                  cursor: pointer;
                  text-decoration: none;
                  display: inline-flex;
                  align-items: center;
            }

            #${this.config.bannerId} .download-btn::before {
                content: '';
                position: absolute;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                border: 1px solid transparent;
                border-radius: 16px;
                background: linear-gradient(180deg, #6CBAFF, #3A84FF) border-box;
                -webkit-mask:
                    linear-gradient(#fff 0 0) padding-box,
                    linear-gradient(#fff 0 0);
                -webkit-mask-composite: xor;
                mask-composite: exclude;
                pointer-events: none;
            }

            #${this.config.bannerId} .download-btn-content {
                display: inline-flex;
                align-items: center;
                gap: 4px;
                padding: 6px 24px;
            }

            #${this.config.bannerId} .chrome-icon {
                width: 18px;
                height: 18px;
            }

            #${this.config.bannerId} .download-btn:hover::before {
                background: linear-gradient(180deg, #3A84FF, #6CBAFF) border-box;
            }

            #${this.config.bannerId} .close-btn {
                background: transparent;
                border: none;
                color: #4D4F56;
                font-size: 22px;
                cursor: pointer;
                padding: 0 12px;
                margin-left: auto;
                transition: color 0.2s ease;
                display: flex;
                align-items: center;
                justify-content: center;
                width: 36px;
                height: 36px;
                border-radius: 50%;
            }

            #${this.config.bannerId} .close-btn:hover {
                color: #4D4F56;
                background: rgba(0, 0, 0, 0.05);
            }
        `;

    // 创建样式元素并添加到 head
    const styleElement = document.createElement('style');
    styleElement.textContent = styles;
    document.head.appendChild(styleElement);
  }

  /**
   * 创建横幅内容
   * @returns 横幅内容元素
   */
  private createBannerContent(): HTMLElement {
    const contentDiv = document.createElement('div');
    contentDiv.className = 'banner-content';

    // 创建消息部分
    const messageDiv = document.createElement('div');
    messageDiv.className = 'banner-message';

    // Chrome 图标
    const alertIcon = document.createElement('img');
    alertIcon.className = 'alert-icon';
    alertIcon.src = alertIconSrc;

    // 消息文本
    const messageText = document.createElement('span');
    messageText.textContent =
      '检测到您当前的浏览器非Chrome。为获得最佳兼容性与体验效果，推荐您使用最新版本的Chrome浏览器';

    // 组装消息部分
    messageDiv.appendChild(alertIcon);
    messageDiv.appendChild(messageText);

    // 下载按钮
    const downloadBtn = document.createElement('a');
    downloadBtn.href = this.config.downloadUrl as string;
    downloadBtn.target = '_blank';
    downloadBtn.rel = 'noopener noreferrer';
    downloadBtn.className = 'download-btn';

    const downloadText = document.createElement('div');
    downloadText.className = 'download-btn-content';
    downloadText.innerHTML = `<img src="${chromeIconSrc}" alt="Chrome" class="chrome-icon" />去下载`;

    downloadBtn.appendChild(downloadText);

    // 关闭按钮
    const closeBtn = document.createElement('button');
    closeBtn.className = 'close-btn';
    closeBtn.innerHTML = '&times;';

    // 组装横幅内容
    contentDiv.appendChild(messageDiv);
    contentDiv.appendChild(downloadBtn);
    contentDiv.appendChild(closeBtn);

    return contentDiv;
  }

  /**
   * 创建横幅元素
   */
  private createBannerElement(): void {
    // 如果横幅已存在，先移除
    const existingBanner = document.getElementById(this.config.bannerId as string);
    if (existingBanner) {
      existingBanner.remove();
    }

    // 创建横幅容器
    this.bannerElement = document.createElement('div');
    this.bannerElement.id = this.config.bannerId as string;
    this.bannerElement.className = this.config.bannerClass as string;

    // 设置初始样式
    this.applyBannerStyles();

    // 创建横幅内容
    const bannerContent = this.createBannerContent();
    this.bannerElement.appendChild(bannerContent);

    // 添加到页面顶部
    document.body.insertBefore(this.bannerElement, document.body.firstChild);

    // 添加事件监听器
    this.addEventListeners();
  }
}

export const useChromeNotice = (config: BannerConfig = {}) => {
  const chromeBanner = new ChromeBanner(config);
  return {
    show: () => {
      chromeBanner.show();
    },
    hide: () => {
      chromeBanner.hide();
    },
    destroy: () => {
      chromeBanner.destroy();
    },
    toggle: () => {
      chromeBanner.toggle();
    },
  };
};
