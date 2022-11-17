/**
 * @file 替换 asset css 中的 BK_STATIC_URL，__webpack_public_path__ 没法解决 asset 里静态资源的 url
 * @author
 */

const { extname } = require('path');

class ReplaceCSSStaticUrlPlugin {
  apply(compiler) {
    // emit: 在生成资源并输出到目录之前
    compiler.hooks.emit.tapAsync('ReplaceCSSStaticUrlPlugin', (compilation, callback) => {
      const assets = Object.keys(compilation.assets);
      const assetsLen = assets.length;

      for (let i = 0; i < assetsLen; i++) {
        const fileName = assets[i];
        const name = extname(fileName);
        // if (extname(fileName) !== '.css'&&!fileName.includes('static/fonts/iconcool.')) {
        if (extname(fileName) !== '.css') {
          continue;
        }
        const asset = compilation.assets[fileName];
        let minifyFileContent = asset.source().toString()
          .replace(
            /\{\{\s*BK_STATIC_URL\s*\}\}/g,
            () => '../',
          );
        compilation.assets[fileName] = {
          // 返回文件内容
          source: () => minifyFileContent,
          // 返回文件大小
          size: () => Buffer.byteLength(minifyFileContent, 'utf8'),
        };
      }

      callback();
    });
  }
}


module.exports = ReplaceCSSStaticUrlPlugin;


