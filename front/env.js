const { DEV_HOST } = require('./env.dev')
const config = {
    production: {
      NODE_ENV: JSON.stringify('production'),
      publicPath:'{{ BK_STATIC_URL }}',// 静态资源路径前缀
      // publicPath:'/static/dist/',// 静态资源路径前缀
      outputDir: __dirname + '/dist',// 打包输出目录
      assetsDir: '', // js/css/font资源归属目录
      AJAX_URL_PREFIX: '',
    },
    open: {
      NODE_ENV: JSON.stringify('production'),
      publicPath:'./',
      outputDir: __dirname + '/dist',
      assetsDir: '',
      AJAX_URL_PREFIX: '',
    },
    development: {
      NODE_ENV: JSON.stringify('development'),
      publicPath:'/',
      outputDir: __dirname + '/dist',
      assetsDir: './',
      AJAX_URL_PREFIX: '',
      DEV_HOST: DEV_HOST,
      AJAX_URL_PROXY: ''
    },
  };
  
  module.exports = () => {
    let env = config.production
    if (process.env?.version === 'open') {
      env = config.open
    }
    if (process.env.NODE_ENV === 'development') {
      env = config.development;
    }
    return env;
  };
  