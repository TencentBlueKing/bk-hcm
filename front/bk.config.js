const path = require('path');
const fs = require('fs');
const argv = require('minimist')(process.argv.slice(2));
const BuildHashPlugin = require('./build-hash-plugin');

const appDir = fs.realpathSync(process.cwd());
const resolveBase = (relativePath) => path.resolve(appDir, relativePath);

const getConfig = (custom = {}) => ({
  assetsDir: '',
  outputAssetsDirName: '',
  outputDir: custom.outputDir ?? 'dist',
  publicPath: custom.publicPath ?? custom.env?.BK_STATIC_URL ?? process.env.BK_STATIC_URL,
  host: custom.host ?? custom.env?.BK_APP_HOST ?? process.env.BK_APP_HOST,
  port: custom.port ?? custom.env?.BK_APP_PORT ?? process.env.BK_APP_PORT,
  cache: true,
  open: true,
  typescript: true,
  forkTsChecker: false,
  bundleAnalysis: false,
  replaceStatic: false,
  target: 'web',
  lazyCompilation: false,
  lazyCompilationHost: 'localhost',
  envPrefix: 'BK_',
  copy: {
    from: './static',
    to: './dist/',
  },
  resource: {
    main: {
      entry: './src/main',
      html: {
        filename: 'index.html',
        template: './index.html',
        templateParameters: custom.env ?? process.env,
      },
    },
  },
  css: {
    scssLoaderOptions: {
      additionalData: '@import "./src/style/variables.scss";',
    },
  },
  configureWebpack() {
    return {
      resolve: {
        alias: {
          '@pluginHandler': resolveBase('./src/plugin-handler'),
        },
      },
      devServer: {
        hot: true, // 启用热模块替换
        liveReload: false, // 禁用整页刷新
        server: custom.server && custom.server,
      },
    };
  },
  chainWebpack: (config) => {
    config.module
      .rule('ts')
      .test(/\.m?ts$/)
      .use('swc-loader')
      .loader('swc-loader')
      .options({
        jsc: {
          parser: {
            syntax: 'typescript',
            decorators: true,
            tsx: true,
          },
          transform: {
            legacyDecorator: true,
            decoratorMetadata: true,
          },
          target: 'es2015',
        },
      });

    if (process.env.NODE_ENV === 'production') {
      config.plugin('buildHash').use(BuildHashPlugin);
    }

    return config;
  },
});

const targetEnv = argv._[1];
const customDevConfigPath = resolveBase(`env.${targetEnv || 'local'}.config.js`);
const isCustomDevConfigExist = fs.existsSync(customDevConfigPath);

let customConfig = () => {};
if (isCustomDevConfigExist) {
  customConfig = require(customDevConfigPath);
}

module.exports = getConfig(customConfig(targetEnv));
