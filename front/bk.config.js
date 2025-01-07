const path = require('path');
const fs = require('fs');
const argv = require('minimist')(process.argv.slice(2));

const appDir = fs.realpathSync(process.cwd());
const resolveBase = (relativePath) => path.resolve(appDir, relativePath);

const getConfig = (custom = {}) => ({
  assetsDir: '',
  outputAssetsDirName: '',
  outputDir: custom.outputDir ?? 'dist',
  publicPath: custom.publicPath ?? process.env.BK_STATIC_URL,
  host: custom.host ?? process.env.BK_APP_HOST,
  port: custom.port ?? process.env.BK_APP_PORT,
  cache: true,
  open: true,
  typescript: true,
  forkTsChecker: false,
  bundleAnalysis: false,
  replaceStatic: false,
  target: 'web',
  lazyCompilation: true,
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
        templateParameters: process.env,
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
        server: custom.server && custom.server,
      },
    };
  },
  chainWebpack: (config) => config,
});

const customDevConfigPath = resolveBase(`env.${argv._[1] || 'local'}.config.js`);
const isCustomDevConfigExist = fs.existsSync(customDevConfigPath);

let customConfig = {};
if (isCustomDevConfigExist) {
  customConfig = require(customDevConfigPath);
}

module.exports = getConfig(customConfig);
