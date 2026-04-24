module.exports = {
  extends: ['stylelint-config-standard-scss', 'stylelint-config-standard-vue/scss', 'stylelint-config-html/vue'],
  rules: {
    'no-empty-source': null,
    'selector-pseudo-class-no-unknown': [
      true,
      {
        ignorePseudoClasses: ['global', 'deep'],
      },
    ],
  },
};
