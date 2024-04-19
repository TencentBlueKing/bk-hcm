const CopyWebpackPlugin = require('copy-webpack-plugin')
const { resolve } = require('path');
const replaceStaticUrlPlugin = require('./replace-static-url-plugin')
const isModeProduction = process.env.NODE_ENV === 'production';
const indexPath = isModeProduction ? './index.html' : './index-dev.html'
const env = require('./env')();
const apiMocker = require('./mock-server.js')
module.exports = {
  appConfig() {
    return {
      indexPath,
      mainPath: './src/main.ts',
      publicPath: env.publicPath,
      outputDir: env.outputDir,
      assetsDir: env.assetsDir,
      minChunkSize: 10000,
      // pages: {
      //   main: {
      //     entry: './src/main.ts',
      //     filename: 'index.html'
      //   },
      // },
      // needSplitChunks: false,
      css: {
        loaderOptions: {
          scss: {
            additionalData: '@import "./src/style/variables.scss";',
          },
        },
      },
      devServer : {
        host: env.DEV_HOST,
        port: 5000,
        historyApiFallback: true,
        disableHostCheck: true,
        before(app) {
          apiMocker(app, {
                // watch: [
                //   '/mock/api/v4/organization/user_info/',
                //   '/mock/api/v4/add/',
                //   '/mock/api/v4/get/',
                //   '/mock/api/v4/sync/',
                //   '/mock/api/v4/cloud/public_images/list/'
                // ],
                api: resolve(__dirname, './mock/api.ts')
            })
        },
        proxy: {
        }
      }
    }
  },
  configureWebpack(_webpackConfig) {
    webpackConfig = _webpackConfig;
    webpackConfig.plugins.push(
      new replaceStaticUrlPlugin(),
    )
    webpackConfig.plugins.push(
      new CopyWebpackPlugin({
        patterns: [
          {
            from: resolve('static/image'),
            to: resolve('dist'),
            globOptions: {
              ignore: [
                // 忽略所有 HTML 文件，如果有的话
                '**/*.html',
              ],
            },
          },
          {
            from: 'static/*.html', // 只匹配 static 目录下的 HTML 文件
            to: '[name][ext]', // 保持原文件名
          },
        ],
      })
    )
    
    // webpackConfig.externals = {
    //   'axios':'axios',
    //   'dayjs':'dayjs',
    // }
    webpackConfig.resolve = {
      ...webpackConfig.resolve,
      symlinks: false,
      extensions: ['.js', '.vue', '.json', '.ts', '.tsx'],
      alias: {
        ...webpackConfig.resolve?.alias,
        // extensions: ['.js', '.jsx', '.ts', '.tsx'],
        '@': resolve(__dirname, './src'),
        '@static': resolve(__dirname, './static'),
        '@charts': resolve(__dirname, './src/plugins/charts'),
        '@datasource': resolve(__dirname, './src/plugins/datasource'),
        '@modules': resolve(__dirname, './src/store/modules'),
      },
    };
  },
};
